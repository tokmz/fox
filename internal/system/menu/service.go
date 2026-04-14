package menu

import "context"

// Service 菜单服务接口
type Service interface {
	// Create 创建菜单
	Create(ctx context.Context, req *CreateReq) error

	// Delete 删除菜单（存在子菜单时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq) error

	// Update 更新菜单
	Update(ctx context.Context, req *UpdateReq) error

	// UpdateStatus 修改菜单状态
	UpdateStatus(ctx context.Context, req *StatusReq) error

	// Detail 查询菜单详情
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 返回菜单选项树（用于权限分配）
	Options(ctx context.Context) ([]*OptionResp, error)

	// List 查询菜单列表（返回树形结构）
	List(ctx context.Context, req *ListReq) ([]*TreeResp, error)
}
