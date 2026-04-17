package dept

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/fox/pkg/errcode"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 缓存 key 定义
const (
	deptDetailKey  = "dept:detail:%d" // dept:detail:{id}
	deptOptionsKey = "dept:options"   // 部门选项列表缓存
)

// 缓存 TTL
const (
	deptDetailTTL  = 30 * time.Minute
	deptOptionsTTL = 10 * time.Minute
)

// Service 部门服务接口
type Service interface {
	// Create 创建部门
	Create(ctx context.Context, req *CreateReq, operatorID int64) error

	// Delete 删除部门（存在子部门/关联用户/关联岗位时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq, operatorID int64) error

	// Update 更新部门
	Update(ctx context.Context, req *UpdateReq, operatorID int64) error

	// UpdateStatus 修改部门状态
	UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error

	// Detail 查询部门详情，带缓存
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 返回部门选项列表（仅启用的部门），带缓存
	Options(ctx context.Context) ([]*OptionResp, error)

	// List 查询部门列表（返回树形结构）
	List(ctx context.Context, req *ListReq) ([]*TreeResp, error)
}

// service 部门服务实现
type service struct {
	logger logger.Logger
	cache  cache.Cache
	db     *gorm.DB
}

// NewService 创建部门服务实例
func NewService(logger logger.Logger, cache cache.Cache, db *gorm.DB) Service {
	return &service{
		logger: logger,
		cache:  cache,
		db:     db,
	}
}

// ===== 写操作 =====

// Create 创建部门
// 同一事务内：编码唯一性校验 → 计算物化路径 → 插入
func (s *service) Create(ctx context.Context, req *CreateReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 编码唯一性校验
		var count int64
		if err := tx.Model(&entity.SysDept{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
			return errcode.ErrDeptQuery.WithErr(err)
		}
		if count > 0 {
			return errcode.ErrDeptCodeExists.WithMessagef("部门编码已存在: %s", req.Code)
		}

		dept := createReqToEntity(req)
		dept.CreatedBy = operatorID
		dept.UpdatedBy = operatorID

		// 计算物化路径
		if err := computeTreePath(tx, dept); err != nil {
			return err
		}

		if err := tx.Create(dept).Error; err != nil {
			return errcode.ErrDeptCreate.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "创建部门失败", zap.String("name", req.Name), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, deptOptionsKey)
	return nil
}

// Delete 批量删除部门
// 批量查询校验（子部门 + 关联用户 + 关联岗位） → 同一事务内：更新操作人 → 软删除
func (s *service) Delete(ctx context.Context, req *DeleteReq, operatorID int64) error {
	var depts []entity.SysDept
	if err := s.db.WithContext(ctx).
		Where("id IN ?", req.IDs).
		Find(&depts).Error; err != nil {
		return errcode.ErrDeptQuery.WithErr(err)
	}
	if len(depts) == 0 {
		return errcode.ErrDeptNotFound
	}

	// 逐条校验
	for _, d := range depts {
		// 存在子部门时拒绝删除
		var childCount int64
		if err := s.db.WithContext(ctx).Model(&entity.SysDept{}).
			Where("parent_id = ?", d.ID).Count(&childCount).Error; err != nil {
			return errcode.ErrDeptQuery.WithErr(err)
		}
		if childCount > 0 {
			return errcode.ErrDeptHasChildren.WithMessagef("部门 [%s] 存在子部门", d.Name)
		}

		// 关联用户时拒绝删除
		var userCount int64
		if err := s.db.WithContext(ctx).Model(&entity.SysUser{}).
			Where("dept_id = ?", d.ID).Count(&userCount).Error; err != nil {
			return errcode.ErrDeptQuery.WithErr(err)
		}
		if userCount > 0 {
			return errcode.ErrDeptHasUsers.WithMessagef("部门 [%s] 下存在用户", d.Name)
		}

		// 关联岗位时拒绝删除
		var postCount int64
		if err := s.db.WithContext(ctx).Model(&entity.SysPost{}).
			Where("dept_id = ?", d.ID).Count(&postCount).Error; err != nil {
			return errcode.ErrDeptQuery.WithErr(err)
		}
		if postCount > 0 {
			return errcode.ErrDeptHasPosts.WithMessagef("部门 [%s] 下存在岗位", d.Name)
		}
	}

	// 同一事务内：更新操作人 → 软删除
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.SysDept{}).Where("id IN ?", req.IDs).
			Update("updated_by", operatorID).Error; err != nil {
			return errcode.ErrDeptDelete.WithErr(err)
		}
		if err := tx.Delete(&entity.SysDept{}, req.IDs).Error; err != nil {
			return errcode.ErrDeptDelete.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "批量删除部门失败",
			zap.Int("count", len(req.IDs)), zap.Error(err))
		return err
	}

	s.invalidateDetailAndOptions(ctx, depts)
	return nil
}

// Update 更新部门
// 加载 → 校验 → 修改字段 → 父级变更时重算物化路径 → 写入
func (s *service) Update(ctx context.Context, req *UpdateReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dept entity.SysDept
		if err := tx.First(&dept, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrDeptNotFound
			}
			return errcode.ErrDeptQuery.WithErr(err)
		}

		// 编码唯一性校验
		if req.Code != "" {
			var count int64
			if err := tx.Model(&entity.SysDept{}).
				Where("code = ? AND id != ?", req.Code, req.ID).Count(&count).Error; err != nil {
				return errcode.ErrDeptQuery.WithErr(err)
			}
			if count > 0 {
				return errcode.ErrDeptCodeExists.WithMessagef("部门编码已存在: %s", req.Code)
			}
			dept.Code = req.Code
		}

		if req.Name != "" {
			dept.Name = req.Name
		}
		if req.ParentID != nil {
			dept.ParentID = req.ParentID
		}
		if req.DeptType != nil {
			dept.DeptType = *req.DeptType
		}
		if req.LeaderID != nil {
			dept.LeaderID = req.LeaderID
		}
		if req.Sort != nil {
			dept.Sort = *req.Sort
		}
		if req.Status != nil {
			dept.Status = *req.Status
		}

		// 父级变更时重算物化路径
		if req.ParentID != nil {
			if err := computeTreePath(tx, &dept); err != nil {
				return err
			}
		}

		dept.UpdatedBy = operatorID
		if err := tx.Model(&dept).
			Select("parent_id", "name", "code", "dept_type", "leader_id",
				"sort", "status", "level", "tree", "updated_by").
			Updates(&dept).Error; err != nil {
			return errcode.ErrDeptUpdate.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "更新部门失败", zap.Int64("dept_id", req.ID), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(deptDetailKey, req.ID))
	_ = s.cache.Del(ctx, deptOptionsKey)
	return nil
}

// UpdateStatus 修改部门状态
func (s *service) UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error {
	result := s.db.WithContext(ctx).
		Model(&entity.SysDept{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]any{
			"status":     req.Status,
			"updated_by": operatorID,
		})
	if result.Error != nil {
		s.logger.ErrorContext(ctx, "更新部门状态失败", zap.Int("count", len(req.IDs)), zap.Error(result.Error))
		return errcode.ErrDeptUpdate.WithErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return errcode.ErrDeptNotFound
	}

	for _, id := range req.IDs {
		_ = s.cache.Del(ctx, fmt.Sprintf(deptDetailKey, id))
	}
	_ = s.cache.Del(ctx, deptOptionsKey)
	return nil
}

// ===== 读操作 =====

// Detail 查询部门详情
func (s *service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	key := fmt.Sprintf(deptDetailKey, req.ID)

	var resp DetailResp
	if err := s.cache.GetOrSet(ctx, key, &resp, deptDetailTTL, func() (any, error) {
		var dept entity.SysDept
		if err := s.db.WithContext(ctx).First(&dept, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, cache.ErrNotFound
			}
			return nil, errcode.ErrDeptQuery.WithErr(err)
		}
		return entityToDetailResp(&dept), nil
	}); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return nil, errcode.ErrDeptNotFound
		}
		return nil, err
	}

	return &resp, nil
}

// Options 返回启用的部门选项列表（下拉选择用）
func (s *service) Options(ctx context.Context) ([]*OptionResp, error) {
	var options []*OptionResp
	if err := s.cache.GetOrSet(ctx, deptOptionsKey, &options, deptOptionsTTL, func() (any, error) {
		var depts []*entity.SysDept
		if err := s.db.WithContext(ctx).
			Where("status = 1").
			Order("sort ASC, id ASC").
			Find(&depts).Error; err != nil {
			return nil, err
		}

		opts := make([]*OptionResp, 0, len(depts))
		for _, d := range depts {
			opts = append(opts, entityToOptionResp(d))
		}
		return opts, nil
	}); err != nil {
		return nil, err
	}

	return options, nil
}

// List 查询部门列表并构建为树形结构
func (s *service) List(ctx context.Context, req *ListReq) ([]*TreeResp, error) {
	tx := s.db.WithContext(ctx).Model(&entity.SysDept{})

	if req.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Code != "" {
		tx = tx.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.DeptType != nil {
		tx = tx.Where("dept_type = ?", *req.DeptType)
	}
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}

	var depts []*entity.SysDept
	if err := tx.Order("sort ASC, id ASC").Find(&depts).Error; err != nil {
		return nil, err
	}

	return buildTree(depts), nil
}

// ===== 缓存失效辅助 =====

func (s *service) invalidateDetailAndOptions(ctx context.Context, depts []entity.SysDept) {
	detailKeys := make([]string, 0, len(depts))
	for _, d := range depts {
		detailKeys = append(detailKeys, fmt.Sprintf(deptDetailKey, d.ID))
	}

	if len(detailKeys) > 0 {
		_ = s.cache.Del(ctx, detailKeys...)
	}

	_ = s.cache.Del(ctx, deptOptionsKey)
}

// ===== 辅助函数 =====

// computeTreePath 根据 ParentID 计算 Level 和 Tree
func computeTreePath(db *gorm.DB, dept *entity.SysDept) error {
	if dept.ParentID == nil {
		dept.Level = 0
		dept.Tree = "0"
		return nil
	}

	var parent entity.SysDept
	if err := db.Select("id, level, tree").First(&parent, *dept.ParentID).Error; err != nil {
		return fmt.Errorf("父部门不存在: %w", err)
	}

	dept.Level = parent.Level + 1
	dept.Tree = fmt.Sprintf("%s,%d", parent.Tree, parent.ID)
	if !strings.HasPrefix(dept.Tree, "0") {
		dept.Tree = "0," + dept.Tree
	}
	return nil
}

// buildTree 将平铺的部门列表构建为树形结构
func buildTree(depts []*entity.SysDept) []*TreeResp {
	respMap := make(map[int64]*TreeResp, len(depts))
	for _, d := range depts {
		respMap[d.ID] = entityToTreeResp(d)
	}

	var roots []*TreeResp
	for _, d := range depts {
		node := respMap[d.ID]
		if d.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := respMap[*d.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}
