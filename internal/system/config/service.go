package config

import "context"

// Service 系统配置服务
type Service interface {
	// ===== 分组 =====

	// CreateGroup 创建分组
	CreateGroup(ctx context.Context, req *CreateGroupReq) error

	// GetGroup 获取分组（含配置项列表）
	GetGroup(ctx context.Context, req *GetGroupReq) (*GetGroupResp, error)

	// DeleteGroup 删除分组（级联删除配置项）
	DeleteGroup(ctx context.Context, req *DeleteGroupReq) error

	// ===== 配置项 =====

	// CreateItem 创建配置项
	CreateItem(ctx context.Context, req *CreateItemReq) error

	// DeleteItem 删除配置项
	DeleteItem(ctx context.Context, req *DeleteItemReq) error

	// ListItems 查询配置项列表（根据分组）
	ListItems(ctx context.Context, req *ListItemsReq) ([]*ItemResp, error)

	// ===== 自定义更新（按业务场景扩展） =====
}
