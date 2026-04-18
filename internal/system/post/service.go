package post

import (
	"context"
	"errors"
	"fmt"
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
	postDetailKey  = "post:detail:%d"  // post:detail:{id}
	postOptionsKey = "post:options:%d" // post:options:{deptID}
)

// 缓存 TTL
const (
	postDetailTTL  = 30 * time.Minute
	postOptionsTTL = 10 * time.Minute
)

// Service 岗位服务接口
type Service interface {
	// Create 创建岗位
	Create(ctx context.Context, req *CreateReq, operatorID int64) error

	// Delete 删除岗位（存在关联用户时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq, operatorID int64) error

	// Update 更新岗位
	Update(ctx context.Context, req *UpdateReq, operatorID int64) error

	// UpdateStatus 修改岗位状态
	UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error

	// Detail 查询岗位详情
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 根据部门ID获取该部门下的岗位选项列表
	Options(ctx context.Context, req *OptionsReq) ([]*PostOptionItemResp, error)
}

type service struct {
	logger logger.Logger
	cache  cache.Cache
	db     *gorm.DB
}

// ===== 写操作 =====

func (s *service) Create(ctx context.Context, req *CreateReq, operatorID int64) error {
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&entity.SysPost{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询岗位名称失败", zap.String("name", req.Name), zap.Error(err))
			return errcode.ErrPostCreate.WithErr(err)
		}
		if count > 0 {
			s.logger.WarnContext(ctx, "岗位名称已存在", zap.String("name", req.Name))
			return errcode.ErrPostNameExists
		}

		if err := tx.Model(&entity.SysPost{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询岗位编码失败", zap.String("code", req.Code), zap.Error(err))
			return errcode.ErrPostCreate.WithErr(err)
		}
		if count > 0 {
			s.logger.WarnContext(ctx, "岗位编码已存在", zap.String("code", req.Code))
			return errcode.ErrPostCodeExists
		}

		post := CreateReqToEntity(req)
		post.CreatedBy = operatorID
		post.UpdatedBy = operatorID
		if err := tx.Create(post).Error; err != nil {
			s.logger.ErrorContext(ctx, "插入岗位失败", zap.String("name", req.Name), zap.Error(err))
			return errcode.ErrPostCreate.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "创建岗位失败", zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(postOptionsKey, req.DeptID))
	return nil
}

func (s *service) Delete(ctx context.Context, req *DeleteReq, operatorID int64) error {
	if len(req.IDs) == 0 {
		return nil
	}

	var posts []entity.SysPost
	if err := s.db.WithContext(ctx).
		Select("id, dept_id").
		Where("id IN ?", req.IDs).
		Find(&posts).Error; err != nil {
		s.logger.ErrorContext(ctx, "查询待删除岗位失败", zap.Error(err))
		return errcode.ErrPostQuery.WithErr(err)
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 记录删除操作人
		if err := tx.Model(&entity.SysPost{}).Where("id IN ?", req.IDs).
			Update("updated_by", operatorID).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新岗位操作人失败", zap.Int64("operator_id", operatorID), zap.Error(err))
			return errcode.ErrPostDelete.WithErr(err)
		}

		var count int64
		if err := tx.Model(&entity.SysUserPost{}).Where("post_id IN ?", req.IDs).Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询岗位关联用户失败", zap.Error(err))
			return errcode.ErrPostHasUsersQuery.WithErr(err)
		}
		if count > 0 && !req.Force {
			s.logger.WarnContext(ctx, "岗位已分配用户", zap.Int("count", len(req.IDs)), zap.Int64("user_count", count))
			return errcode.ErrPostHasUsers
		}

		if err := tx.Where("post_id IN ?", req.IDs).Delete(&entity.SysUserPost{}).Error; err != nil {
			s.logger.ErrorContext(ctx, "删除岗位用户关联失败", zap.Error(err))
			return errcode.ErrPostDeleteUsers.WithErr(err)
		}

		if err := tx.Delete(&entity.SysPost{}, req.IDs).Error; err != nil {
			s.logger.ErrorContext(ctx, "软删除岗位失败", zap.Int("count", len(req.IDs)), zap.Error(err))
			return errcode.ErrPostDelete.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "删除岗位失败", zap.Error(err))
		return err
	}

	s.invalidateDetailAndOptions(ctx, posts)
	return nil
}

func (s *service) Update(ctx context.Context, req *UpdateReq, operatorID int64) error {
	var oldDeptID int64

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.SysPost
		if err := tx.Select("id, dept_id").First(&existing, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.WarnContext(ctx, "岗位不存在", zap.Int64("id", req.ID))
				return errcode.ErrPostNotFound
			}
			s.logger.ErrorContext(ctx, "查询岗位失败", zap.Int64("id", req.ID), zap.Error(err))
			return errcode.ErrPostUpdate.WithErr(err)
		}
		oldDeptID = existing.DeptID

		var count int64
		if req.Name != "" {
			if err := tx.Model(&entity.SysPost{}).
				Where("name = ? AND id != ?", req.Name, req.ID).
				Count(&count).Error; err != nil {
				s.logger.ErrorContext(ctx, "查询岗位名称失败", zap.String("name", req.Name), zap.Error(err))
				return errcode.ErrPostUpdate.WithErr(err)
			}
			if count > 0 {
				s.logger.WarnContext(ctx, "岗位名称已存在", zap.String("name", req.Name))
				return errcode.ErrPostNameExists
			}
		}
		if req.Code != "" {
			if err := tx.Model(&entity.SysPost{}).
				Where("code = ? AND id != ?", req.Code, req.ID).
				Count(&count).Error; err != nil {
				s.logger.ErrorContext(ctx, "查询岗位编码失败", zap.String("code", req.Code), zap.Error(err))
				return errcode.ErrPostUpdate.WithErr(err)
			}
			if count > 0 {
				s.logger.WarnContext(ctx, "岗位编码已存在", zap.String("code", req.Code))
				return errcode.ErrPostCodeExists
			}
		}

		post := UpdateReqToEntity(req)
		post.UpdatedBy = operatorID
		if err := tx.Model(post).
			Select("dept_id", "name", "code", "sort", "remark", "updated_by").
			Updates(post).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新岗位数据失败", zap.Int64("id", req.ID), zap.Error(err))
			return errcode.ErrPostUpdate.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "更新岗位失败", zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, fmt.Sprintf(postDetailKey, req.ID))
	_ = s.cache.Del(ctx, fmt.Sprintf(postOptionsKey, oldDeptID))
	if req.DeptID != oldDeptID {
		_ = s.cache.Del(ctx, fmt.Sprintf(postOptionsKey, req.DeptID))
	}
	return nil
}

func (s *service) UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error {
	var posts []entity.SysPost
	if err := s.db.WithContext(ctx).
		Select("id, dept_id").
		Where("id IN ?", req.IDs).
		Find(&posts).Error; err != nil {
		s.logger.ErrorContext(ctx, "查询岗位失败", zap.Error(err))
		return errcode.ErrPostQuery.WithErr(err)
	}

	result := s.db.WithContext(ctx).
		Model(&entity.SysPost{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]any{
			"status":     req.Status,
			"updated_by": operatorID,
		})
	if result.Error != nil {
		s.logger.ErrorContext(ctx, "修改岗位状态失败", zap.Int("count", len(req.IDs)), zap.Error(result.Error))
		return errcode.ErrPostUpdate.WithErr(result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.WarnContext(ctx, "岗位不存在", zap.Int64s("ids", req.IDs))
		return errcode.ErrPostNotFound
	}

	s.invalidateDetailAndOptions(ctx, posts)
	return nil
}

// ===== 读操作 =====

func (s *service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	key := fmt.Sprintf(postDetailKey, req.ID)
	var resp DetailResp
	if err := s.cache.GetOrSet(ctx, key, &resp, postDetailTTL, func() (any, error) {
		var post entity.SysPost
		if err := s.db.WithContext(ctx).First(&post, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, cache.ErrNotFound
			}
			return nil, errcode.ErrPostQuery.WithErr(err)
		}
		r := EntityToDetailResp(&post)
		return r, nil
	}); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return nil, errcode.ErrPostNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (s *service) Options(ctx context.Context, req *OptionsReq) ([]*PostOptionItemResp, error) {
	key := fmt.Sprintf(postOptionsKey, req.DeptID)
	var result []*PostOptionItemResp
	if err := s.cache.GetOrSet(ctx, key, &result, postOptionsTTL, func() (any, error) {
		var posts []*entity.SysPost
		if err := s.db.WithContext(ctx).
			Select("id, name, code").
			Where("dept_id = ? AND status = 1", req.DeptID).
			Find(&posts).Error; err != nil {
			return nil, errcode.ErrPostQuery.WithErr(err)
		}

		items := make([]*PostOptionItemResp, 0, len(posts))
		for _, p := range posts {
			items = append(items, &PostOptionItemResp{
				ID:   p.ID,
				Name: p.Name,
				Code: p.Code,
			})
		}
		return items, nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// ===== 缓存失效辅助 =====

func (s *service) invalidateDetailAndOptions(ctx context.Context, posts []entity.SysPost) {
	deptIDs := make(map[int64]struct{})
	detailKeys := make([]string, 0, len(posts))

	for _, p := range posts {
		detailKeys = append(detailKeys, fmt.Sprintf(postDetailKey, p.ID))
		deptIDs[p.DeptID] = struct{}{}
	}

	if len(detailKeys) > 0 {
		_ = s.cache.Del(ctx, detailKeys...)
	}

	for deptID := range deptIDs {
		_ = s.cache.Del(ctx, fmt.Sprintf(postOptionsKey, deptID))
	}
}

func NewService(logger logger.Logger, c cache.Cache, db *gorm.DB) Service {
	return &service{
		logger: logger,
		cache:  c,
		db:     db,
	}
}
