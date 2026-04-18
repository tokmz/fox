package role

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/fox/pkg/datascope"
	"github.com/tokmz/fox/pkg/errcode"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 缓存 key 定义
const (
	roleDetailKey  = "role:detail:%d" // role:detail:{id}
	roleOptionsKey = "role:options"   // 角色选项列表缓存
)

// 缓存 TTL
const (
	roleDetailTTL  = 30 * time.Minute
	roleOptionsTTL = 10 * time.Minute
)

// Service 角色服务接口
type Service interface {
	// Create 创建角色（含菜单分配）
	Create(ctx context.Context, req *CreateReq, operatorID int64) error

	// Delete 删除角色（存在子角色或关联用户时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq, operatorID int64) error

	// Update 更新角色
	Update(ctx context.Context, req *UpdateReq, operatorID int64) error

	// UpdateStatus 修改角色状态
	UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error

	// Detail 查询角色详情（含已分配菜单ID），带缓存
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 返回角色选项列表（仅启用的角色），带缓存
	Options(ctx context.Context) ([]*OptionResp, error)

	// List 查询角色列表（返回树形结构）
	List(ctx context.Context, req *ListReq) ([]*TreeResp, error)

	// AssignMenus 全量替换角色菜单权限
	AssignMenus(ctx context.Context, req *AssignMenusReq, operatorID int64) error

	// AssignDepts 全量替换角色自定义部门（DataScope=2 时使用）
	AssignDepts(ctx context.Context, req *AssignDeptsReq, operatorID int64) error
}

// service 角色服务实现
type service struct {
	logger logger.Logger
	cache  cache.Cache
	db     *gorm.DB
}

// NewService 创建角色服务实例
func NewService(logger logger.Logger, cache cache.Cache, db *gorm.DB) Service {
	return &service{
		logger: logger,
		cache:  cache,
		db:     db,
	}
}

// ===== 写操作 =====

// Create 创建角色
// 同一事务内：唯一性校验 → 计算物化路径 → 插入角色 → 分配菜单
func (s *service) Create(ctx context.Context, req *CreateReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 名称唯一性校验
		var count int64
		if err := tx.Model(&entity.SysRole{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
			return errcode.ErrRoleQuery.WithErr(err)
		}
		if count > 0 {
			return errcode.ErrRoleExists.WithMessagef("角色名称已存在: %s", req.Name)
		}

		// 编码唯一性校验
		if err := tx.Model(&entity.SysRole{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
			return errcode.ErrRoleQuery.WithErr(err)
		}
		if count > 0 {
			return errcode.ErrRoleExists.WithMessagef("角色编码已存在: %s", req.Code)
		}

		// DTO → Entity
		role := createReqToEntity(req)
		role.CreatedBy = operatorID
		role.UpdatedBy = operatorID

		// 计算物化路径
		if err := computeTreePath(tx, role); err != nil {
			return err
		}

		// 插入角色
		if err := tx.Create(role).Error; err != nil {
			return errcode.ErrRoleCreate.WithErr(err)
		}

		// 分配菜单（如有）
		if len(req.MenuIDs) > 0 {
			records := make([]entity.SysRoleMenu, 0, len(req.MenuIDs))
			for _, mid := range req.MenuIDs {
				records = append(records, entity.SysRoleMenu{RoleID: role.ID, MenuID: mid})
			}
			if err := tx.Create(&records).Error; err != nil {
				return errcode.ErrRoleCreate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "创建角色失败", zap.String("name", req.Name), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, roleOptionsKey)
	return nil
}

// Delete 批量删除角色
// 批量查询校验 → 同一事务内：更新操作人 → 软删除
func (s *service) Delete(ctx context.Context, req *DeleteReq, operatorID int64) error {
	// 批量查询所有待删除角色（校验 + 缓存失效用）
	var roles []entity.SysRole
	if err := s.db.WithContext(ctx).
		Where("id IN ?", req.IDs).
		Find(&roles).Error; err != nil {
		return errcode.ErrRoleQuery.WithErr(err)
	}
	if len(roles) == 0 {
		return errcode.ErrRoleNotFound
	}

	// 校验内置角色
	for _, r := range roles {
		if r.Builtin {
			return errcode.ErrRoleBuiltin.WithMessagef("角色 [%s] 为内置角色", r.Name)
		}
	}

	// 批量校验子角色
	type RoleChild struct {
		ParentID int64
		Name     string
	}
	var children []RoleChild
	if err := s.db.WithContext(ctx).Model(&entity.SysRole{}).
		Select("parent_id, name").
		Where("parent_id IN ?", req.IDs).
		Limit(1).
		Find(&children).Error; err != nil {
		return errcode.ErrRoleQuery.WithErr(err)
	}
	if len(children) > 0 {
		for _, r := range roles {
			if r.ID == children[0].ParentID {
				return errcode.ErrRoleHasChildren.WithMessagef("角色 [%s] 存在子角色", r.Name)
			}
		}
	}

	// 批量校验关联用户
	type RoleUser struct {
		RoleID int64
	}
	var users []RoleUser
	if err := s.db.WithContext(ctx).Model(&entity.SysUserRole{}).
		Select("role_id").
		Where("role_id IN ?", req.IDs).
		Limit(1).
		Find(&users).Error; err != nil {
		return errcode.ErrRoleQuery.WithErr(err)
	}
	if len(users) > 0 {
		for _, r := range roles {
			if r.ID == users[0].RoleID {
				return errcode.ErrRoleHasUsers.WithMessagef("角色 [%s] 已分配用户", r.Name)
			}
		}
	}

	// 查询所有使用这些角色的用户ID（用于清理数据权限缓存）
	var userIDs []int64
	if err := s.db.WithContext(ctx).Model(&entity.SysUserRole{}).
		Where("role_id IN ?", req.IDs).
		Pluck("user_id", &userIDs).Error; err != nil {
		return errcode.ErrRoleQuery.WithErr(err)
	}

	// 同一事务内：更新操作人 → 软删除
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.SysRole{}).Where("id IN ?", req.IDs).
			Update("updated_by", operatorID).Error; err != nil {
			return errcode.ErrRoleDelete.WithErr(err)
		}
		if err := tx.Delete(&entity.SysRole{}, req.IDs).Error; err != nil {
			return errcode.ErrRoleDelete.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "批量删除角色失败",
			zap.Int("count", len(req.IDs)), zap.Error(err))
		return err
	}

	s.invalidateDetailAndOptions(ctx, roles)

	// 清理使用这些角色的用户的数据权限缓存
	for _, uid := range userIDs {
		_ = datascope.ClearCache(ctx, s.cache, uid)
	}

	return nil
}

// Update 更新角色
// 加载 → 校验 → 修改字段 → 父级变更时重算物化路径 → 写入
func (s *service) Update(ctx context.Context, req *UpdateReq, operatorID int64) error {
	var dataScopeChanged bool

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role entity.SysRole
		if err := tx.First(&role, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrRoleNotFound
			}
			return errcode.ErrRoleQuery.WithErr(err)
		}

		if role.Builtin {
			return errcode.ErrRoleBuiltin
		}

		updates := map[string]any{"updated_by": operatorID}

		// 名称唯一性校验
		if req.Name != "" {
			var count int64
			if err := tx.Model(&entity.SysRole{}).
				Where("name = ? AND id != ?", req.Name, req.ID).Count(&count).Error; err != nil {
				return errcode.ErrRoleQuery.WithErr(err)
			}
			if count > 0 {
				return errcode.ErrRoleExists.WithMessagef("角色名称已存在: %s", req.Name)
			}
			updates["name"] = req.Name
		}

		// 编码唯一性校验
		if req.Code != "" {
			var count int64
			if err := tx.Model(&entity.SysRole{}).
				Where("code = ? AND id != ?", req.Code, req.ID).Count(&count).Error; err != nil {
				return errcode.ErrRoleQuery.WithErr(err)
			}
			if count > 0 {
				return errcode.ErrRoleExists.WithMessagef("角色编码已存在: %s", req.Code)
			}
			updates["code"] = req.Code
		}

		// 父级变更校验循环引用
		if req.ParentID != nil {
			if *req.ParentID != 0 {
				var parent entity.SysRole
				if err := tx.Select("id, tree").First(&parent, *req.ParentID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errcode.ErrRoleNotFound.WithMessagef("父角色不存在: %d", *req.ParentID)
					}
					return errcode.ErrRoleQuery.WithErr(err)
				}
				// 检查循环引用
				if strings.Contains(parent.Tree, fmt.Sprintf(",%d,", req.ID)) ||
					strings.HasSuffix(parent.Tree, fmt.Sprintf(",%d", req.ID)) {
					return errcode.ErrRoleUpdate.WithMessagef("不能将角色的父级设置为其子孙角色")
				}
			}
			role.ParentID = req.ParentID
			if err := computeTreePath(tx, &role); err != nil {
				return err
			}
			updates["parent_id"] = req.ParentID
			updates["level"] = role.Level
			updates["tree"] = role.Tree
		}

		if req.DataScope != nil {
			dataScopeChanged = *req.DataScope != role.DataScope
			updates["data_scope"] = *req.DataScope
		}
		if req.DeptCheckStrictly != nil {
			updates["dept_check_strictly"] = *req.DeptCheckStrictly
		}
		if req.Sort != nil {
			updates["sort"] = *req.Sort
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}

		if err := tx.Model(&entity.SysRole{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			return errcode.ErrRoleUpdate.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "更新角色失败", zap.Int64("role_id", req.ID), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(roleDetailKey, req.ID))
	_ = s.cache.Del(ctx, roleOptionsKey)

	// DataScope 变更时清理使用该角色的用户数据权限缓存
	if dataScopeChanged {
		var userIDs []int64
		if err := s.db.WithContext(ctx).Model(&entity.SysUserRole{}).
			Where("role_id = ?", req.ID).
			Pluck("user_id", &userIDs).Error; err == nil {
			for _, uid := range userIDs {
				_ = datascope.ClearCache(ctx, s.cache, uid)
			}
		}
	}

	return nil
}

// UpdateStatus 修改角色状态（仅修改 status 字段）
func (s *service) UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error {
	// 先检查角色是否为内置角色
	var role entity.SysRole
	if err := s.db.WithContext(ctx).Select("id, builtin").First(&role, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoleNotFound
		}
		return errcode.ErrRoleQuery.WithErr(err)
	}
	if role.Builtin {
		return errcode.ErrRoleBuiltin
	}

	result := s.db.WithContext(ctx).
		Model(&entity.SysRole{}).
		Where("id = ?", req.ID).
		Updates(map[string]any{
			"status":     req.Status,
			"updated_by": operatorID,
		})
	if result.Error != nil {
		s.logger.ErrorContext(ctx, "更新角色状态失败", zap.Int64("role_id", req.ID), zap.Error(result.Error))
		return errcode.ErrRoleUpdate.WithErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return errcode.ErrRoleNotFound
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(roleDetailKey, req.ID))
	_ = s.cache.Del(ctx, roleOptionsKey)
	return nil
}

// AssignMenus 全量替换角色菜单权限
// 同一事务内：更新操作人 → 清空旧关联 → 写入新关联
func (s *service) AssignMenus(ctx context.Context, req *AssignMenusReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 存在性校验
		var count int64
		if err := tx.Model(&entity.SysRole{}).Where("id = ?", req.RoleID).Count(&count).Error; err != nil {
			return errcode.ErrRoleQuery.WithErr(err)
		}
		if count == 0 {
			return errcode.ErrRoleNotFound
		}

		// 更新操作人
		if err := tx.Model(&entity.SysRole{}).Where("id = ?", req.RoleID).
			Update("updated_by", operatorID).Error; err != nil {
			return errcode.ErrRoleUpdate.WithErr(err)
		}

		// 清空旧关联
		if err := tx.Where("role_id = ?", req.RoleID).Delete(&entity.SysRoleMenu{}).Error; err != nil {
			return errcode.ErrRoleUpdate.WithErr(err)
		}

		// 写入新关联
		if len(req.MenuIDs) > 0 {
			records := make([]entity.SysRoleMenu, 0, len(req.MenuIDs))
			for _, mid := range req.MenuIDs {
				records = append(records, entity.SysRoleMenu{RoleID: req.RoleID, MenuID: mid})
			}
			if err := tx.Create(&records).Error; err != nil {
				return errcode.ErrRoleUpdate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "分配角色菜单失败",
			zap.Int64("role_id", req.RoleID), zap.Int("count", len(req.MenuIDs)), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(roleDetailKey, req.RoleID))
	return nil
}

// AssignDepts 全量替换角色自定义部门（DataScope=2 时使用）
// 同一事务内：更新操作人 → 清空旧关联 → 写入新关联
func (s *service) AssignDepts(ctx context.Context, req *AssignDeptsReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 存在性校验
		var count int64
		if err := tx.Model(&entity.SysRole{}).Where("id = ?", req.RoleID).Count(&count).Error; err != nil {
			return errcode.ErrRoleQuery.WithErr(err)
		}
		if count == 0 {
			return errcode.ErrRoleNotFound
		}

		// 更新操作人
		if err := tx.Model(&entity.SysRole{}).Where("id = ?", req.RoleID).
			Update("updated_by", operatorID).Error; err != nil {
			return errcode.ErrRoleUpdate.WithErr(err)
		}

		// 清空旧关联
		if err := tx.Where("role_id = ?", req.RoleID).Delete(&entity.SysRoleDept{}).Error; err != nil {
			return errcode.ErrRoleUpdate.WithErr(err)
		}

		// 写入新关联
		if len(req.DeptIDs) > 0 {
			records := make([]entity.SysRoleDept, 0, len(req.DeptIDs))
			for _, did := range req.DeptIDs {
				records = append(records, entity.SysRoleDept{RoleID: req.RoleID, DeptID: did})
			}
			if err := tx.Create(&records).Error; err != nil {
				return errcode.ErrRoleUpdate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "分配角色部门失败",
			zap.Int64("role_id", req.RoleID), zap.Int("count", len(req.DeptIDs)), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(roleDetailKey, req.RoleID))

	// 清理使用该角色的用户的数据权限缓存
	var userIDs []int64
	if err := s.db.WithContext(ctx).Model(&entity.SysUserRole{}).
		Where("role_id = ?", req.RoleID).
		Pluck("user_id", &userIDs).Error; err == nil {
		for _, uid := range userIDs {
			_ = datascope.ClearCache(ctx, s.cache, uid)
		}
	}

	return nil
}

// ===== 读操作 =====

// Detail 查询角色详情
// 通过 GetOrSet 原子加载：缓存命中直接返回，未命中查库并回写（内置 singleflight 防击穿）
func (s *service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	key := fmt.Sprintf(roleDetailKey, req.ID)

	var resp DetailResp
	if err := s.cache.GetOrSet(ctx, key, &resp, roleDetailTTL, func() (any, error) {
		var role entity.SysRole
		var menuIDs, deptIDs []int64

		// 使用事务保证数据一致性
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.First(&role, req.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return cache.ErrNotFound
				}
				return errcode.ErrRoleQuery.WithErr(err)
			}

			if err := tx.Model(&entity.SysRoleMenu{}).
				Where("role_id = ?", req.ID).
				Pluck("menu_id", &menuIDs).Error; err != nil {
				return errcode.ErrRoleQuery.WithErr(err)
			}

			if err := tx.Model(&entity.SysRoleDept{}).
				Where("role_id = ?", req.ID).
				Pluck("dept_id", &deptIDs).Error; err != nil {
				return errcode.ErrRoleQuery.WithErr(err)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		return entityToDetailResp(&role, menuIDs, deptIDs), nil
	}); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return nil, errcode.ErrRoleNotFound
		}
		return nil, err
	}

	return &resp, nil
}

// Options 返回启用的角色选项列表（下拉选择用）
// 通过 GetOrSet 缓存，写操作后自动失效
func (s *service) Options(ctx context.Context) ([]*OptionResp, error) {
	var options []*OptionResp
	if err := s.cache.GetOrSet(ctx, roleOptionsKey, &options, roleOptionsTTL, func() (any, error) {
		var roles []*entity.SysRole
		if err := s.db.WithContext(ctx).
			Where("status = 1").
			Order("sort ASC, id ASC").
			Find(&roles).Error; err != nil {
			return nil, err
		}

		opts := make([]*OptionResp, 0, len(roles))
		for _, r := range roles {
			opts = append(opts, entityToOptionResp(r))
		}
		return opts, nil
	}); err != nil {
		return nil, err
	}

	return options, nil
}

// List 查询角色列表并构建为树形结构
// 支持按名称/编码模糊搜索、状态精确过滤
func (s *service) List(ctx context.Context, req *ListReq) ([]*TreeResp, error) {
	tx := s.db.WithContext(ctx).Model(&entity.SysRole{})
	tx = tx.Scopes(datascope.Apply(ctx, "", "id", "created_by"))

	if req.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Code != "" {
		tx = tx.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}

	var roles []*entity.SysRole
	if err := tx.Order("sort ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}

	return buildTree(roles), nil
}

// ===== 缓存失效辅助 =====

func (s *service) invalidateDetailAndOptions(ctx context.Context, roles []entity.SysRole) {
	detailKeys := make([]string, 0, len(roles))
	for _, r := range roles {
		detailKeys = append(detailKeys, fmt.Sprintf(roleDetailKey, r.ID))
	}

	if len(detailKeys) > 0 {
		_ = s.cache.Del(ctx, detailKeys...)
	}

	_ = s.cache.Del(ctx, roleOptionsKey)
}

// ===== 辅助函数 =====

// computeTreePath 根据 ParentID 计算 Level 和 Tree
// 接受 *gorm.DB 以便在事务内复用
func computeTreePath(db *gorm.DB, role *entity.SysRole) error {
	if role.ParentID == nil {
		role.Level = 0
		role.Tree = "0"
		return nil
	}

	var parent entity.SysRole
	if err := db.Select("id, level, tree").First(&parent, *role.ParentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrRoleNotFound.WithMessagef("父角色不存在: %d", *role.ParentID)
		}
		return errcode.ErrRoleQuery.WithErr(err)
	}

	role.Level = parent.Level + 1
	role.Tree = fmt.Sprintf("%s,%d", parent.Tree, parent.ID)
	if !strings.HasPrefix(role.Tree, "0") {
		role.Tree = "0," + role.Tree
	}
	return nil
}

// buildTree 将平铺的角色列表构建为树形结构
// 通过 map 建立父子关系，ParentID=nil 的节点为根节点
func buildTree(roles []*entity.SysRole) []*TreeResp {
	respMap := make(map[int64]*TreeResp, len(roles))
	for _, r := range roles {
		respMap[r.ID] = entityToTreeResp(r)
	}

	var roots []*TreeResp
	for _, r := range roles {
		node := respMap[r.ID]
		if r.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := respMap[*r.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}
