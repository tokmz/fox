package api

import "context"

// Service 接口管理服务
type Service interface {
	// ===== 接口（API） =====

	// CreateApi 创建接口
	CreateApi(ctx context.Context, req *CreateApiReq) error

	// UpdateApi 更新接口
	UpdateApi(ctx context.Context, req *UpdateApiReq) error

	// DeleteApi 删除接口
	DeleteApi(ctx context.Context, req *DeleteApiReq) error

	// DetailApi 查询接口详情
	DetailApi(ctx context.Context, req *DetailApiReq) (*DetailApiResp, error)

	// ListApi 查询接口列表
	ListApi(ctx context.Context, req *ListApiReq) ([]*ApiItemResp, error)

	// ===== 分组（Group） =====

	// CreateGroup 创建分组
	CreateGroup(ctx context.Context, req *CreateGroupReq) error

	// UpdateGroup 更新分组
	UpdateGroup(ctx context.Context, req *UpdateGroupReq) error

	// DeleteGroup 删除分组（存在子分组或关联接口时拒绝删除）
	DeleteGroup(ctx context.Context, req *DeleteGroupReq) error

	// GroupTree 查询分组树形结构（含每个分组下的接口列表）
	GroupTree(ctx context.Context, req *GroupTreeReq) ([]*GroupTreeResp, error)
}
