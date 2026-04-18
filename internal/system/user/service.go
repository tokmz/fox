package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/fox/pkg/datascope"
	"github.com/tokmz/fox/pkg/errcode"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service 用户服务接口
type Service interface {
	// Create 创建用户（含角色/岗位关联）
	Create(ctx context.Context, req *CreateReq, operatorID int64) error

	// Delete 删除用户（软删除，支持批量）
	Delete(ctx context.Context, req *DeleteReq, operatorID int64) error

	// Update 更新用户基本信息（不含角色/岗位）
	Update(ctx context.Context, req *UpdateReq, operatorID int64) error

	// UpdateStatus 修改用户状态（支持批量）
	UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error

	// Detail 获取用户详情（含关联数据：部门、角色、岗位），带缓存
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// List 用户列表（分页，支持多条件查询）
	List(ctx context.Context, req *ListReq) ([]*ListItemResp, int64, error)

	// Options 返回用户选项列表（仅启用的用户），带缓存
	Options(ctx context.Context) ([]*OptionResp, error)

	// ResetPassword 重置用户密码（管理员操作）
	ResetPassword(ctx context.Context, req *ResetPasswordReq, operatorID int64) error

	// AssignRoles 分配用户角色（全量替换）
	AssignRoles(ctx context.Context, req *AssignRolesReq, operatorID int64) error

	// AssignPosts 分配用户岗位（全量替换）
	AssignPosts(ctx context.Context, req *AssignPostsReq, operatorID int64) error
}

// 缓存 key 定义
const (
	userDetailKey  = "user:detail:%d" // user:detail:{id}
	userOptionsKey = "user:options"   // 用户选项列表缓存
)

// 缓存 TTL
const (
	userDetailTTL  = 30 * time.Minute
	userOptionsTTL = 10 * time.Minute
)

// service 用户服务实现
type service struct {
	logger logger.Logger
	cache  cache.Cache
	db     *gorm.DB
}

// NewService 创建用户服务实例
func NewService(logger logger.Logger, cache cache.Cache, db *gorm.DB) Service {
	return &service{
		logger: logger,
		cache:  cache,
		db:     db,
	}
}

// ===== 写操作 =====

// Create 创建用户（含角色/岗位关联）
// 同一事务内：用户名唯一性校验 → 插入用户 → 分配角色 → 分配岗位
func (s *service) Create(ctx context.Context, req *CreateReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 用户名唯一性校验
		var count int64
		if err := tx.Model(&entity.SysUser{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
			return errcode.ErrUserQuery.WithErr(err)
		}
		if count > 0 {
			return errcode.ErrUserExists.WithMessagef("用户名已存在: %s", req.Username)
		}

		// 创建用户
		user := createReqToEntity(req)
		user.CreatedBy = operatorID
		user.UpdatedBy = operatorID
		user.Password = hashPassword(req.Password)

		if err := tx.Create(user).Error; err != nil {
			s.logger.ErrorContext(ctx, "插入用户失败", zap.String("username", req.Username), zap.Error(err))
			return errcode.ErrUserCreate.WithErr(err)
		}

		// 分配角色
		if len(req.RoleIDs) > 0 {
			roleRecords := make([]entity.SysUserRole, 0, len(req.RoleIDs))
			for _, rid := range req.RoleIDs {
				roleRecords = append(roleRecords, entity.SysUserRole{UserID: user.ID, RoleID: rid})
			}
			if err := tx.Create(&roleRecords).Error; err != nil {
				s.logger.ErrorContext(ctx, "创建用户角色关联失败", zap.Int64("user_id", user.ID), zap.Error(err))
				return errcode.ErrUserCreate.WithErr(err)
			}
		}

		// 分配岗位
		if len(req.PostIDs) > 0 {
			postRecords := make([]entity.SysUserPost, 0, len(req.PostIDs))
			for _, pid := range req.PostIDs {
				postRecords = append(postRecords, entity.SysUserPost{UserID: user.ID, PostID: pid})
			}
			if err := tx.Create(&postRecords).Error; err != nil {
				s.logger.ErrorContext(ctx, "创建用户岗位关联失败", zap.Int64("user_id", user.ID), zap.Error(err))
				return errcode.ErrUserCreate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "创建用户事务失败", zap.String("username", req.Username), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, userOptionsKey)
	return nil
}

// Delete 批量删除用户（软删除）
func (s *service) Delete(ctx context.Context, req *DeleteReq, operatorID int64) error {
	if len(req.IDs) == 0 {
		return nil
	}

	// 查询待删除用户
	var users []entity.SysUser
	if err := s.db.WithContext(ctx).
		Where("id IN ?", req.IDs).
		Find(&users).Error; err != nil {
		s.logger.ErrorContext(ctx, "查询待删除用户失败", zap.Any("ids", req.IDs), zap.Error(err))
		return errcode.ErrUserQuery.WithErr(err)
	}
	if len(users) == 0 {
		s.logger.WarnContext(ctx, "待删除用户不存在", zap.Any("ids", req.IDs))
		return errcode.ErrUserNotFound
	}

	// 软删除（同时更新 updated_by）
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.SysUser{}).
			Where("id IN ?", req.IDs).
			Updates(map[string]any{"updated_by": operatorID}).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新用户操作人失败", zap.Any("ids", req.IDs), zap.Error(err))
			return errcode.ErrUserDelete.WithErr(err)
		}
		if err := tx.Delete(&entity.SysUser{}, req.IDs).Error; err != nil {
			s.logger.ErrorContext(ctx, "软删除用户失败", zap.Any("ids", req.IDs), zap.Error(err))
			return errcode.ErrUserDelete.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "批量删除用户事务失败",
			zap.Int("count", len(req.IDs)), zap.Error(err))
		return err
	}

	s.invalidateDetailAndOptions(ctx, users)
	return nil
}

// Update 更新用户基本信息（不含角色/岗位）
func (s *service) Update(ctx context.Context, req *UpdateReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user entity.SysUser
		if err := tx.First(&user, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.WarnContext(ctx, "用户不存在", zap.Int64("id", req.ID))
				return errcode.ErrUserNotFound
			}
			s.logger.ErrorContext(ctx, "查询用户失败", zap.Int64("id", req.ID), zap.Error(err))
			return errcode.ErrUserQuery.WithErr(err)
		}

		// 用户名唯一性校验
		if req.Username != "" && req.Username != user.Username {
			var count int64
			if err := tx.Model(&entity.SysUser{}).
				Where("username = ? AND id != ?", req.Username, req.ID).Count(&count).Error; err != nil {
				s.logger.ErrorContext(ctx, "查询用户名失败", zap.String("username", req.Username), zap.Error(err))
				return errcode.ErrUserQuery.WithErr(err)
			}
			if count > 0 {
				s.logger.WarnContext(ctx, "用户名已存在", zap.String("username", req.Username))
				return errcode.ErrUserExists.WithMessagef("用户名已存在: %s", req.Username)
			}
		}

		// 构建更新字段
		updates := map[string]any{"updated_by": operatorID}
		if req.Username != "" {
			updates["username"] = req.Username
		}
		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.Email != "" {
			updates["email"] = req.Email
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.Avatar != "" {
			updates["avatar"] = req.Avatar
		}
		if req.Gender != nil {
			updates["gender"] = *req.Gender
		}
		if req.DeptID != nil {
			updates["dept_id"] = *req.DeptID
		}
		if req.Remark != "" {
			updates["remark"] = req.Remark
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}

		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新用户数据失败", zap.Int64("id", req.ID), zap.Error(err))
			return errcode.ErrUserUpdate.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "更新用户事务失败", zap.Int64("user_id", req.ID), zap.Error(err))
		return err
	}

	// 仅在影响 options 字段时才清除 options 缓存
	_ = s.cache.Del(ctx, fmt.Sprintf(userDetailKey, req.ID))
	if req.Username != "" || req.Nickname != "" || req.DeptID != nil || req.Status != nil {
		_ = s.cache.Del(ctx, userOptionsKey)
	}
	// 部门变更时清除数据权限缓存
	if req.DeptID != nil {
		_ = datascope.ClearCache(ctx, s.cache, req.ID)
	}
	return nil
}

// UpdateStatus 修改用户状态（支持批量）
func (s *service) UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error {
	if len(req.IDs) == 0 {
		return nil
	}

	// 存在性校验
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.SysUser{}).
		Where("id IN ?", req.IDs).Count(&count).Error; err != nil {
		return errcode.ErrUserQuery.WithErr(err)
	}
	if count == 0 {
		return errcode.ErrUserNotFound
	}

	// 更新状态
	result := s.db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]any{
			"status":     req.Status,
			"updated_by": operatorID,
		})
	if result.Error != nil {
		s.logger.ErrorContext(ctx, "更新用户状态失败", zap.Int("count", len(req.IDs)), zap.Error(result.Error))
		return errcode.ErrUserUpdate.WithErr(result.Error)
	}

	for _, id := range req.IDs {
		_ = s.cache.Del(ctx, fmt.Sprintf(userDetailKey, id))
	}
	_ = s.cache.Del(ctx, userOptionsKey)
	return nil
}

// ResetPassword 重置用户密码（管理员操作）
func (s *service) ResetPassword(ctx context.Context, req *ResetPasswordReq, operatorID int64) error {
	// 存在性校验
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.SysUser{}).
		Where("id = ?", req.ID).Count(&count).Error; err != nil {
		return errcode.ErrUserQuery.WithErr(err)
	}
	if count == 0 {
		return errcode.ErrUserNotFound
	}

	result := s.db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Where("id = ?", req.ID).
		Updates(map[string]any{
			"password":   hashPassword(req.NewPassword),
			"updated_by": operatorID,
		})
	if result.Error != nil {
		s.logger.ErrorContext(ctx, "重置用户密码失败", zap.Int64("user_id", req.ID), zap.Error(result.Error))
		return errcode.ErrUserUpdate.WithErr(result.Error)
	}

	return nil
}

// AssignRoles 分配用户角色（全量替换）
func (s *service) AssignRoles(ctx context.Context, req *AssignRolesReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 存在性校验
		var count int64
		if err := tx.Model(&entity.SysUser{}).Where("id = ?", req.UserID).Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询用户失败", zap.Int64("user_id", req.UserID), zap.Error(err))
			return errcode.ErrUserQuery.WithErr(err)
		}
		if count == 0 {
			s.logger.WarnContext(ctx, "用户不存在", zap.Int64("user_id", req.UserID))
			return errcode.ErrUserNotFound
		}

		// 更新操作人
		if err := tx.Model(&entity.SysUser{}).Where("id = ?", req.UserID).
			Update("updated_by", operatorID).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新用户操作人失败", zap.Int64("user_id", req.UserID), zap.Error(err))
			return errcode.ErrUserUpdate.WithErr(err)
		}

		// 清空旧关联
		if err := tx.Where("user_id = ?", req.UserID).Delete(&entity.SysUserRole{}).Error; err != nil {
			s.logger.ErrorContext(ctx, "删除旧角色关联失败", zap.Int64("user_id", req.UserID), zap.Error(err))
			return errcode.ErrUserUpdate.WithErr(err)
		}

		// 写入新关联
		if len(req.RoleIDs) > 0 {
			records := make([]entity.SysUserRole, 0, len(req.RoleIDs))
			for _, rid := range req.RoleIDs {
				records = append(records, entity.SysUserRole{UserID: req.UserID, RoleID: rid})
			}
			if err := tx.Create(&records).Error; err != nil {
				s.logger.ErrorContext(ctx, "创建新角色关联失败",
					zap.Int64("user_id", req.UserID), zap.Int("count", len(req.RoleIDs)), zap.Error(err))
				return errcode.ErrUserUpdate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "分配用户角色失败",
			zap.Int64("user_id", req.UserID), zap.Int("count", len(req.RoleIDs)), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(userDetailKey, req.UserID))
	_ = datascope.ClearCache(ctx, s.cache, req.UserID)
	return nil
}

// AssignPosts 分配用户岗位（全量替换）
func (s *service) AssignPosts(ctx context.Context, req *AssignPostsReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 存在性校验
		var count int64
		if err := tx.Model(&entity.SysUser{}).Where("id = ?", req.UserID).Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询用户失败", zap.Int64("user_id", req.UserID), zap.Error(err))
			return errcode.ErrUserQuery.WithErr(err)
		}
		if count == 0 {
			s.logger.WarnContext(ctx, "用户不存在", zap.Int64("user_id", req.UserID))
			return errcode.ErrUserNotFound
		}

		// 更新操作人
		if err := tx.Model(&entity.SysUser{}).Where("id = ?", req.UserID).
			Update("updated_by", operatorID).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新用户操作人失败", zap.Int64("user_id", req.UserID), zap.Error(err))
			return errcode.ErrUserUpdate.WithErr(err)
		}

		// 清空旧关联
		if err := tx.Where("user_id = ?", req.UserID).Delete(&entity.SysUserPost{}).Error; err != nil {
			s.logger.ErrorContext(ctx, "删除旧岗位关联失败", zap.Int64("user_id", req.UserID), zap.Error(err))
			return errcode.ErrUserUpdate.WithErr(err)
		}

		// 写入新关联
		if len(req.PostIDs) > 0 {
			records := make([]entity.SysUserPost, 0, len(req.PostIDs))
			for _, pid := range req.PostIDs {
				records = append(records, entity.SysUserPost{UserID: req.UserID, PostID: pid})
			}
			if err := tx.Create(&records).Error; err != nil {
				s.logger.ErrorContext(ctx, "创建新岗位关联失败",
					zap.Int64("user_id", req.UserID), zap.Int("count", len(req.PostIDs)), zap.Error(err))
				return errcode.ErrUserUpdate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "分配用户岗位失败",
			zap.Int64("user_id", req.UserID), zap.Int("count", len(req.PostIDs)), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(userDetailKey, req.UserID))
	return nil
}

// ===== 读操作 =====

// Detail 获取用户详情（含关联数据：部门、角色、岗位），带缓存
func (s *service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	key := fmt.Sprintf(userDetailKey, req.ID)

	var resp DetailResp
	if err := s.cache.GetOrSet(ctx, key, &resp, userDetailTTL, func() (any, error) {
		var user entity.SysUser
		if err := s.db.WithContext(ctx).
			Preload("Dept").
			Preload("Roles").
			Preload("Posts").
			First(&user, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, cache.ErrNotFound
			}
			return nil, errcode.ErrUserQuery.WithErr(err)
		}

		// 提取关联数据
		deptName, roleIDs, roleNames, postIDs, postNames := extractUserAssociationsWithIDs(&user)

		return entityToDetailResp(&user, deptName, roleIDs, roleNames, postIDs, postNames), nil
	}); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return nil, errcode.ErrUserNotFound
		}
		return nil, err
	}

	return &resp, nil
}

// List 用户列表（分页，支持多条件查询）
func (s *service) List(ctx context.Context, req *ListReq) ([]*ListItemResp, int64, error) {
	tx := s.db.WithContext(ctx).Model(&entity.SysUser{})

	// 应用数据权限过滤
	tx = tx.Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))

	// 条件过滤
	if req.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Nickname != "" {
		tx = tx.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if req.Phone != "" {
		tx = tx.Where("phone = ?", req.Phone)
	}
	if req.Email != "" {
		tx = tx.Where("email = ?", req.Email)
	}
	if req.DeptID != nil {
		tx = tx.Where("dept_id = ?", *req.DeptID)
	}
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}

	// 统计总数
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrUserQuery.WithErr(err)
	}
	if total == 0 {
		return []*ListItemResp{}, 0, nil
	}

	// 分页查询
	var users []*entity.SysUser
	if err := tx.
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		Offset(req.Offset()).
		Limit(req.Size).
		Order("id DESC").
		Find(&users).Error; err != nil {
		return nil, 0, errcode.ErrUserQuery.WithErr(err)
	}

	// 转换为响应 DTO
	items := make([]*ListItemResp, 0, len(users))
	for _, u := range users {
		deptName, roleNames, postNames := extractUserAssociations(u)
		items = append(items, entityToListItemResp(u, deptName, roleNames, postNames))
	}

	return items, total, nil
}

// Options 返回用户选项列表（仅启用的用户），带缓存
func (s *service) Options(ctx context.Context) ([]*OptionResp, error) {
	var options []*OptionResp
	if err := s.cache.GetOrSet(ctx, userOptionsKey, &options, userOptionsTTL, func() (any, error) {
		var users []*entity.SysUser
		if err := s.db.WithContext(ctx).
			Preload("Dept").
			Where("status = 1").
			Order("id ASC").
			Find(&users).Error; err != nil {
			return nil, err
		}

		opts := make([]*OptionResp, 0, len(users))
		for _, u := range users {
			deptName := ""
			if u.Dept != nil {
				deptName = u.Dept.Name
			}
			opts = append(opts, entityToOptionResp(u, deptName))
		}
		return opts, nil
	}); err != nil {
		return nil, err
	}

	return options, nil
}

// ===== 缓存失效辅助 =====

func (s *service) invalidateDetailAndOptions(ctx context.Context, users []entity.SysUser) {
	detailKeys := make([]string, 0, len(users))
	for _, u := range users {
		detailKeys = append(detailKeys, fmt.Sprintf(userDetailKey, u.ID))
	}

	if len(detailKeys) > 0 {
		_ = s.cache.Del(ctx, detailKeys...)
	}

	_ = s.cache.Del(ctx, userOptionsKey)
}

// ===== 密码加密辅助 =====

// hashPassword 使用 bcrypt 加密密码
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// bcrypt.GenerateFromPassword 仅在 cost 参数无效时返回错误
		// 使用 DefaultCost 不会失败，这里保留 panic 作为防御性编程
		panic(fmt.Sprintf("密码加密失败: %v", err))
	}
	return string(hash)
}

// VerifyPassword 验证密码是否匹配
func VerifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// ===== 关联数据提取辅助 =====

// extractUserAssociations 提取用户关联数据（部门、角色、岗位名称）
func extractUserAssociations(u *entity.SysUser) (deptName string, roleNames, postNames []string) {
	if u.Dept != nil {
		deptName = u.Dept.Name
	}

	roleNames = make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roleNames = append(roleNames, r.Name)
	}

	postNames = make([]string, 0, len(u.Posts))
	for _, p := range u.Posts {
		postNames = append(postNames, p.Name)
	}

	return
}

// extractUserAssociationsWithIDs 提取用户关联数据（含 ID 列表）
func extractUserAssociationsWithIDs(u *entity.SysUser) (deptName string, roleIDs []int64, roleNames []string, postIDs []int64, postNames []string) {
	if u.Dept != nil {
		deptName = u.Dept.Name
	}

	roleIDs = make([]int64, 0, len(u.Roles))
	roleNames = make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roleIDs = append(roleIDs, r.ID)
		roleNames = append(roleNames, r.Name)
	}

	postIDs = make([]int64, 0, len(u.Posts))
	postNames = make([]string, 0, len(u.Posts))
	for _, p := range u.Posts {
		postIDs = append(postIDs, p.ID)
		postNames = append(postNames, p.Name)
	}

	return
}
