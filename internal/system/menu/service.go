package menu

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
	"github.com/tokmz/qi/utils/pointer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 缓存 key 定义
const (
	menuDetailKey  = "menu:detail:%d" // menu:detail:{id}
	menuOptionsKey = "menu:options"   // 菜单选项树缓存
)

// 缓存 TTL
const (
	menuDetailTTL  = 30 * time.Minute
	menuOptionsTTL = 10 * time.Minute
)

// Service 菜单服务接口
type Service interface {
	// Create 创建菜单
	Create(ctx context.Context, req *CreateReq, operatorID int64) error

	// Delete 删除菜单（存在子菜单时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq, operatorID int64) error

	// Update 更新菜单
	Update(ctx context.Context, req *UpdateReq, operatorID int64) error

	// UpdateStatus 修改菜单状态
	UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error

	// Detail 查询菜单详情
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 返回菜单选项树（用于权限分配）
	Options(ctx context.Context) ([]*OptionResp, error)

	// List 查询菜单列表（返回树形结构）
	List(ctx context.Context, req *ListReq) ([]*TreeResp, error)

	// AssignApis 全量替换菜单API权限
	AssignApis(ctx context.Context, req *AssignApisReq, operatorID int64) error
}

// service 菜单服务实现
type service struct {
	logger logger.Logger
	cache  cache.Cache
	db     *gorm.DB
}

// NewService 创建菜单服务实例
func NewService(logger logger.Logger, cache cache.Cache, db *gorm.DB) Service {
	return &service{
		logger: logger,
		cache:  cache,
		db:     db,
	}
}

// ===== 写操作 =====

// Create 创建菜单
// 同一事务内：唯一性校验 → 计算物化路径 → 插入菜单 → 分配API权限
func (s *service) Create(ctx context.Context, req *CreateReq, operatorID int64) error {
	menu := &entity.SysMenu{
		ParentID:     req.ParentID,
		Title:        req.Title,
		Key:          req.Key,
		Path:         req.Path,
		Component:    req.Component,
		Redirect:     req.Redirect,
		Query:        req.Query,
		MenuType:     req.MenuType,
		OpenType:     req.OpenType,
		Icon:         req.Icon,
		Sort:         req.Sort,
		KeepAlive:    valueOrDefault(req.KeepAlive, 0),
		Hidden:       valueOrDefault(req.Hidden, 0),
		Affix:        valueOrDefault(req.Affix, 0),
		AlwaysShow:   valueOrDefault(req.AlwaysShow, 0),
		ActiveMenu:   req.ActiveMenu,
		FrameSrc:     req.FrameSrc,
		ExternalLink: req.ExternalLink,
		Remark:       req.Remark,
		Status:       valueOrDefault(req.Status, 1),
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 校验 Key 唯一性
		var count int64
		if err := tx.Model(&entity.SysMenu{}).Where("route_name = ?", req.Key).Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询菜单Key失败", zap.String("key", req.Key), zap.Error(err))
			return errcode.ErrMenuQuery.WithErr(err)
		}
		if count > 0 {
			s.logger.WarnContext(ctx, "菜单Key已存在", zap.String("key", req.Key))
			return errcode.ErrMenuKeyExists.WithMessagef("路由name [%s] 已存在", req.Key)
		}

		// 计算物化路径
		if err := computeTreePath(tx, menu); err != nil {
			s.logger.ErrorContext(ctx, "计算菜单路径失败", zap.Error(err))
			return err
		}

		// 插入菜单
		if err := tx.Create(menu).Error; err != nil {
			s.logger.ErrorContext(ctx, "插入菜单失败", zap.Error(err))
			return errcode.ErrMenuCreate.WithErr(err)
		}

		// 分配API权限
		if len(req.ApiIDs) > 0 {
			// 校验 API 是否存在
			var count int64
			if err := tx.Model(&entity.SysApi{}).Where("id IN ?", req.ApiIDs).Count(&count).Error; err != nil {
				s.logger.ErrorContext(ctx, "查询API失败", zap.Any("api_ids", req.ApiIDs), zap.Error(err))
				return errcode.ErrMenuQuery.WithErr(err)
			}
			if int(count) != len(req.ApiIDs) {
				s.logger.WarnContext(ctx, "部分API不存在", zap.Int("expected", len(req.ApiIDs)), zap.Int64("found", count))
				return errcode.ErrMenuCreate.WithMessagef("部分 API 不存在")
			}

			menuApis := make([]entity.SysMenuApi, 0, len(req.ApiIDs))
			for _, apiID := range req.ApiIDs {
				menuApis = append(menuApis, entity.SysMenuApi{
					MenuID: menu.ID,
					ApiID:  apiID,
				})
			}
			if err := tx.Create(&menuApis).Error; err != nil {
				s.logger.ErrorContext(ctx, "创建菜单API关联失败", zap.Error(err))
				return errcode.ErrMenuCreate.WithErr(err)
			}
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "创建菜单事务失败", zap.Error(err))
		return err
	}

	s.invalidateCache(ctx, []entity.SysMenu{*menu}, true) // Create 总是清理 options
	s.logger.InfoContext(ctx, "创建菜单成功", zap.Int64("id", menu.ID), zap.String("title", menu.Title))
	return nil
}

// Update 更新菜单
// 使用 map 构建更新字段，支持父级循环引用校验
func (s *service) Update(ctx context.Context, req *UpdateReq, operatorID int64) error {
	var menu entity.SysMenu
	if err := s.db.WithContext(ctx).First(&menu, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.WarnContext(ctx, "菜单不存在", zap.Int64("id", req.ID))
			return errcode.ErrMenuNotFound
		}
		s.logger.ErrorContext(ctx, "查询菜单失败", zap.Int64("id", req.ID), zap.Error(err))
		return errcode.ErrMenuQuery.WithErr(err)
	}

	updates := map[string]any{"updated_by": operatorID}
	needRecomputePath := false
	needClearOptions := false // 是否需要清理 options 缓存

	if req.ParentID != nil && (menu.ParentID == nil || *menu.ParentID != *req.ParentID) {
		// 防止自引用
		if *req.ParentID == req.ID {
			s.logger.WarnContext(ctx, "菜单自引用", zap.Int64("id", req.ID))
			return errcode.ErrMenuCircularRef.WithMessagef("不能将菜单设置为自己的父级")
		}

		// 校验父级循环引用
		if *req.ParentID != 0 {
			var parent entity.SysMenu
			if err := s.db.WithContext(ctx).Select("tree").First(&parent, *req.ParentID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					s.logger.WarnContext(ctx, "父菜单不存在", zap.Int64("parent_id", *req.ParentID))
					return errcode.ErrMenuNotFound.WithMessagef("父菜单不存在")
				}
				s.logger.ErrorContext(ctx, "查询父菜单失败", zap.Int64("parent_id", *req.ParentID), zap.Error(err))
				return errcode.ErrMenuQuery.WithErr(err)
			}
			if strings.Contains(parent.Tree, fmt.Sprintf(",%d", req.ID)) {
				s.logger.WarnContext(ctx, "菜单循环引用", zap.Int64("id", req.ID), zap.Int64("parent_id", *req.ParentID))
				return errcode.ErrMenuCircularRef.WithMessagef("不能将菜单移动到其子菜单下")
			}
		}
		updates["parent_id"] = req.ParentID
		needRecomputePath = true
		needClearOptions = true // ParentID 变更影响树形结构
	}

	if req.Title != "" {
		updates["title"] = req.Title
		needClearOptions = true // Title 显示在 options 中
	}
	if req.Key != "" && req.Key != menu.Key {
		// 校验 Key 唯一性
		var count int64
		if err := s.db.WithContext(ctx).Model(&entity.SysMenu{}).
			Where("route_name = ? AND id != ?", req.Key, req.ID).
			Count(&count).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询菜单Key唯一性失败", zap.String("key", req.Key), zap.Error(err))
			return errcode.ErrMenuQuery.WithErr(err)
		}
		if count > 0 {
			s.logger.WarnContext(ctx, "菜单Key已存在", zap.String("key", req.Key), zap.Int64("menu_id", req.ID))
			return errcode.ErrMenuKeyExists.WithMessagef("路由name [%s] 已存在", req.Key)
		}
		updates["route_name"] = req.Key
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Component != "" {
		updates["component"] = req.Component
	}
	if req.Redirect != "" {
		updates["redirect"] = req.Redirect
	}
	if req.Query != "" {
		updates["query"] = req.Query
	}
	if req.MenuType != nil {
		updates["menu_type"] = pointer.GetOrDefault(req.MenuType, int8(0))
	}
	if req.OpenType != nil {
		updates["open_type"] = pointer.GetOrDefault(req.OpenType, int8(0))
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Sort != nil {
		updates["sort"] = pointer.GetOrDefault(req.Sort, 0)
	}
	if req.KeepAlive != nil {
		updates["keep_alive"] = pointer.GetOrDefault(req.KeepAlive, int8(0))
	}
	if req.Hidden != nil {
		updates["hidden"] = pointer.GetOrDefault(req.Hidden, int8(0))
		needClearOptions = true // Hidden 影响 options 可见性
	}
	if req.Affix != nil {
		updates["affix"] = pointer.GetOrDefault(req.Affix, int8(0))
	}
	if req.AlwaysShow != nil {
		updates["always_show"] = pointer.GetOrDefault(req.AlwaysShow, int8(0))
	}
	if req.ActiveMenu != "" {
		updates["active_menu"] = req.ActiveMenu
	}
	if req.FrameSrc != "" {
		updates["frame_src"] = req.FrameSrc
	}
	if req.ExternalLink != "" {
		updates["external_link"] = req.ExternalLink
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Status != nil {
		updates["status"] = pointer.GetOrDefault(req.Status, int8(0))
		needClearOptions = true // Status 影响 options 可用性
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if needRecomputePath {
			menu.ParentID = updates["parent_id"].(*int64)
			if err := computeTreePath(tx, &menu); err != nil {
				s.logger.ErrorContext(ctx, "重新计算菜单路径失败", zap.Int64("id", req.ID), zap.Error(err))
				return err
			}
			updates["level"] = menu.Level
			updates["tree"] = menu.Tree

			// 级联更新所有子菜单的 tree 和 level
			if err := updateChildrenTreePath(tx, req.ID, menu.Level, menu.Tree); err != nil {
				s.logger.ErrorContext(ctx, "更新子菜单路径失败", zap.Int64("id", req.ID), zap.Error(err))
				return err
			}
		}

		if err := tx.Model(&entity.SysMenu{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新菜单数据失败", zap.Int64("id", req.ID), zap.Error(err))
			return errcode.ErrMenuUpdate.WithErr(err)
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "更新菜单事务失败", zap.Int64("id", req.ID), zap.Error(err))
		return err
	}

	// 清理缓存（包括所有受影响的子菜单）
	if needRecomputePath {
		var allMenus []entity.SysMenu
		if err := s.db.WithContext(ctx).
			Where("tree LIKE ?", "%,"+fmt.Sprintf("%d", req.ID)+",%").
			Or("tree LIKE ?", "%,"+fmt.Sprintf("%d", req.ID)).
			Or("id = ?", req.ID).
			Find(&allMenus).Error; err == nil {
			s.invalidateCache(ctx, allMenus, true) // 树形结构变更总是清理 options
		} else {
			s.invalidateCache(ctx, []entity.SysMenu{menu}, needClearOptions)
		}
	} else {
		s.invalidateCache(ctx, []entity.SysMenu{menu}, needClearOptions)
	}
	s.logger.InfoContext(ctx, "更新菜单成功", zap.Int64("id", req.ID))
	return nil
}

// UpdateStatus 修改菜单状态
func (s *service) UpdateStatus(ctx context.Context, req *StatusReq, operatorID int64) error {
	var menu entity.SysMenu
	if err := s.db.WithContext(ctx).First(&menu, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.WarnContext(ctx, "菜单不存在", zap.Int64("id", req.ID))
			return errcode.ErrMenuNotFound
		}
		s.logger.ErrorContext(ctx, "查询菜单失败", zap.Int64("id", req.ID), zap.Error(err))
		return errcode.ErrMenuQuery.WithErr(err)
	}

	if err := s.db.WithContext(ctx).Model(&entity.SysMenu{}).
		Where("id = ?", req.ID).
		Updates(map[string]any{
			"status":     req.Status,
			"updated_by": operatorID,
		}).Error; err != nil {
		s.logger.ErrorContext(ctx, "修改菜单状态失败", zap.Int64("id", req.ID), zap.Error(err))
		return errcode.ErrMenuUpdate.WithErr(err)
	}

	s.invalidateCache(ctx, []entity.SysMenu{menu}, true) // UpdateStatus 影响 options
	return nil
}

// Delete 批量删除菜单
// 批量查询校验（子菜单） → 同一事务内：更新操作人 → 软删除
func (s *service) Delete(ctx context.Context, req *DeleteReq, operatorID int64) error {
	var menus []entity.SysMenu
	if err := s.db.WithContext(ctx).
		Where("id IN ?", req.IDs).
		Find(&menus).Error; err != nil {
		s.logger.ErrorContext(ctx, "查询待删除菜单失败", zap.Any("ids", req.IDs), zap.Error(err))
		return errcode.ErrMenuQuery.WithErr(err)
	}
	if len(menus) == 0 {
		s.logger.WarnContext(ctx, "待删除菜单不存在", zap.Any("ids", req.IDs))
		return errcode.ErrMenuNotFound
	}

	// 批量校验子菜单
	type MenuChild struct {
		ParentID int64
		Title    string
	}
	var children []MenuChild
	if err := s.db.WithContext(ctx).Model(&entity.SysMenu{}).
		Select("parent_id, title").
		Where("parent_id IN ?", req.IDs).
		Limit(1).
		Find(&children).Error; err != nil {
		s.logger.ErrorContext(ctx, "查询子菜单失败", zap.Any("ids", req.IDs), zap.Error(err))
		return errcode.ErrMenuQuery.WithErr(err)
	}
	if len(children) > 0 {
		for _, m := range menus {
			if m.ID == children[0].ParentID {
				s.logger.WarnContext(ctx, "菜单存在子菜单", zap.Int64("id", m.ID), zap.String("title", m.Title))
				return errcode.ErrMenuHasChildren.WithMessagef("菜单 [%s] 存在子菜单", m.Title)
			}
		}
	}

	// 同一事务内：更新操作人 → 软删除
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.SysMenu{}).Where("id IN ?", req.IDs).
			Update("updated_by", operatorID).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新菜单操作人失败", zap.Any("ids", req.IDs), zap.Error(err))
			return errcode.ErrMenuDelete.WithErr(err)
		}
		if err := tx.Delete(&entity.SysMenu{}, req.IDs).Error; err != nil {
			s.logger.ErrorContext(ctx, "软删除菜单失败", zap.Any("ids", req.IDs), zap.Error(err))
			return errcode.ErrMenuDelete.WithErr(err)
		}
		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "删除菜单事务失败", zap.Any("ids", req.IDs), zap.Error(err))
		return err
	}

	s.invalidateCache(ctx, menus, true) // Delete 影响 options
	s.logger.InfoContext(ctx, "批量删除菜单成功", zap.Int("count", len(req.IDs)))
	return nil
}

// AssignApis 全量替换菜单API权限
func (s *service) AssignApis(ctx context.Context, req *AssignApisReq, operatorID int64) error {
	var menu entity.SysMenu
	if err := s.db.WithContext(ctx).First(&menu, req.MenuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.WarnContext(ctx, "菜单不存在", zap.Int64("menu_id", req.MenuID))
			return errcode.ErrMenuNotFound
		}
		s.logger.ErrorContext(ctx, "查询菜单失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
		return errcode.ErrMenuQuery.WithErr(err)
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 校验 API 是否存在
		if len(req.ApiIDs) > 0 {
			var count int64
			if err := tx.Model(&entity.SysApi{}).Where("id IN ?", req.ApiIDs).Count(&count).Error; err != nil {
				s.logger.ErrorContext(ctx, "查询API失败", zap.Any("api_ids", req.ApiIDs), zap.Error(err))
				return errcode.ErrMenuQuery.WithErr(err)
			}
			if int(count) != len(req.ApiIDs) {
				s.logger.WarnContext(ctx, "部分API不存在", zap.Int("expected", len(req.ApiIDs)), zap.Int64("found", count))
				return errcode.ErrMenuUpdate.WithMessagef("部分 API 不存在")
			}
		}

		// 删除旧关联
		if err := tx.Where("menu_id = ?", req.MenuID).Delete(&entity.SysMenuApi{}).Error; err != nil {
			s.logger.ErrorContext(ctx, "删除旧API关联失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
			return errcode.ErrMenuUpdate.WithErr(err)
		}

		// 插入新关联
		if len(req.ApiIDs) > 0 {
			menuApis := make([]entity.SysMenuApi, 0, len(req.ApiIDs))
			for _, apiID := range req.ApiIDs {
				menuApis = append(menuApis, entity.SysMenuApi{
					MenuID: menu.ID,
					ApiID:  apiID,
				})
			}
			if err := tx.Create(&menuApis).Error; err != nil {
				s.logger.ErrorContext(ctx, "创建新API关联失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
				return errcode.ErrMenuUpdate.WithErr(err)
			}
		}

		// 更新操作人
		if err := tx.Model(&entity.SysMenu{}).Where("id = ?", req.MenuID).
			Update("updated_by", operatorID).Error; err != nil {
			s.logger.ErrorContext(ctx, "更新菜单操作人失败", zap.Int64("menu_id", req.MenuID), zap.Error(err))
			return errcode.ErrMenuUpdate.WithErr(err)
		}

		return nil
	}); err != nil {
		s.logger.ErrorContext(ctx, "分配菜单API权限事务失败",
			zap.Int64("menu_id", req.MenuID), zap.Error(err))
		return err
	}

	s.invalidateCache(ctx, []entity.SysMenu{menu}, false) // AssignApis 不影响 options
	return nil
}

// ===== 读操作 =====

// Detail 查询菜单详情
// 通过 GetOrSet 原子加载：缓存命中直接返回，未命中查库并回写
func (s *service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	key := fmt.Sprintf(menuDetailKey, req.ID)

	var resp DetailResp
	if err := s.cache.GetOrSet(ctx, key, &resp, menuDetailTTL, func() (any, error) {
		var menu entity.SysMenu
		if err := s.db.WithContext(ctx).First(&menu, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, cache.ErrNotFound
			}
			s.logger.ErrorContext(ctx, "查询菜单详情失败", zap.Int64("id", req.ID), zap.Error(err))
			return nil, errcode.ErrMenuQuery.WithErr(err)
		}

		var apiIDs []int64
		if err := s.db.WithContext(ctx).Model(&entity.SysMenuApi{}).
			Where("menu_id = ?", req.ID).
			Pluck("api_id", &apiIDs).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询菜单API关联失败", zap.Int64("menu_id", req.ID), zap.Error(err))
			return nil, errcode.ErrMenuQuery.WithErr(err)
		}

		return entityToDetailResp(&menu, apiIDs), nil
	}); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return nil, errcode.ErrMenuNotFound
		}
		return nil, err
	}

	return &resp, nil
}

// Options 返回启用的菜单选项树（权限分配用）
// 通过 GetOrSet 缓存，写操作后自动失效
func (s *service) Options(ctx context.Context) ([]*OptionResp, error) {
	var options []*OptionResp
	if err := s.cache.GetOrSet(ctx, menuOptionsKey, &options, menuOptionsTTL, func() (any, error) {
		var menus []*entity.SysMenu
		if err := s.db.WithContext(ctx).
			Where("status = 1").
			Order("sort ASC, id ASC").
			Find(&menus).Error; err != nil {
			s.logger.ErrorContext(ctx, "查询菜单选项失败", zap.Error(err))
			return nil, err
		}

		return buildOptionTree(menus), nil
	}); err != nil {
		return nil, err
	}

	return options, nil
}

// List 查询菜单列表并构建为树形结构
// 支持按标题/Key模糊搜索、菜单类型/状态精确过滤
func (s *service) List(ctx context.Context, req *ListReq) ([]*TreeResp, error) {
	tx := s.db.WithContext(ctx).Model(&entity.SysMenu{})

	if req.Title != "" {
		tx = tx.Where("title LIKE ?", "%"+req.Title+"%")
	}
	if req.Key != "" {
		tx = tx.Where("route_name LIKE ?", "%"+req.Key+"%")
	}
	if req.MenuType != nil {
		tx = tx.Where("menu_type = ?", *req.MenuType)
	}
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}

	var menus []*entity.SysMenu
	if err := tx.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		s.logger.ErrorContext(ctx, "查询菜单列表失败", zap.Error(err))
		return nil, err
	}

	return buildTree(menus), nil
}

// ===== 缓存失效辅助 =====

func (s *service) invalidateCache(ctx context.Context, menus []entity.SysMenu, clearOptions bool) {
	detailKeys := make([]string, 0, len(menus))
	for _, m := range menus {
		detailKeys = append(detailKeys, fmt.Sprintf(menuDetailKey, m.ID))
	}

	if len(detailKeys) > 0 {
		_ = s.cache.Del(ctx, detailKeys...)
	}

	if clearOptions {
		_ = s.cache.Del(ctx, menuOptionsKey)
	}
}

// ===== 辅助函数 =====

// computeTreePath 根据 ParentID 计算 Level 和 Tree
func computeTreePath(db *gorm.DB, menu *entity.SysMenu) error {
	if menu.ParentID == nil {
		menu.Level = 0
		menu.Tree = "0"
		return nil
	}

	// 防止自引用
	if menu.ID != 0 && *menu.ParentID == menu.ID {
		return errcode.ErrMenuCircularRef.WithMessagef("菜单不能引用自己")
	}

	var parent entity.SysMenu
	if err := db.Select("id, level, tree").First(&parent, *menu.ParentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrMenuNotFound.WithMessagef("父菜单不存在: %d", *menu.ParentID)
		}
		return errcode.ErrMenuQuery.WithErr(err)
	}

	// 检测循环引用（如果当前菜单已存在，检查父级路径中是否包含自己）
	if menu.ID != 0 {
		if strings.Contains(parent.Tree, fmt.Sprintf(",%d,", menu.ID)) ||
			strings.HasSuffix(parent.Tree, fmt.Sprintf(",%d", menu.ID)) {
			return errcode.ErrMenuCircularRef.WithMessagef("检测到循环引用")
		}
	}

	menu.Level = parent.Level + 1
	menu.Tree = fmt.Sprintf("%s,%d", parent.Tree, parent.ID)
	if !strings.HasPrefix(menu.Tree, "0") {
		menu.Tree = "0," + menu.Tree
	}
	return nil
}

// updateChildrenTreePath 递归更新所有子菜单的 tree 和 level
func updateChildrenTreePath(db *gorm.DB, parentID int64, parentLevel int, parentTree string) error {
	var children []entity.SysMenu
	if err := db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return errcode.ErrMenuQuery.WithErr(err)
	}

	for _, child := range children {
		newLevel := parentLevel + 1
		newTree := fmt.Sprintf("%s,%d", parentTree, parentID)
		if !strings.HasPrefix(newTree, "0") {
			newTree = "0," + newTree
		}

		if err := db.Model(&entity.SysMenu{}).Where("id = ?", child.ID).
			Updates(map[string]any{
				"level": newLevel,
				"tree":  newTree,
			}).Error; err != nil {
			return errcode.ErrMenuUpdate.WithErr(err)
		}

		// 递归更新子菜单的子菜单
		if err := updateChildrenTreePath(db, child.ID, newLevel, newTree); err != nil {
			return err
		}
	}

	return nil
}

// buildTree 将平铺的菜单列表构建为树形结构
func buildTree(menus []*entity.SysMenu) []*TreeResp {
	respMap := make(map[int64]*TreeResp, len(menus))
	for _, m := range menus {
		respMap[m.ID] = entityToTreeResp(m)
	}

	var roots []*TreeResp
	for _, m := range menus {
		node := respMap[m.ID]
		if m.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := respMap[*m.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}

// buildOptionTree 将平铺的菜单列表构建为选项树
func buildOptionTree(menus []*entity.SysMenu) []*OptionResp {
	respMap := make(map[int64]*OptionResp, len(menus))
	for _, m := range menus {
		respMap[m.ID] = entityToOptionResp(m)
	}

	var roots []*OptionResp
	for _, m := range menus {
		node := respMap[m.ID]
		if m.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := respMap[*m.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
	return roots
}

// valueOrDefault 返回指针值或默认值
func valueOrDefault(ptr *int8, defaultVal int8) int8 {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
