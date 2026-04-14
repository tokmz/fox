package role

import "context"

// Service 角色服务接口
type Service interface {
	// Create 创建角色（含菜单分配）
	Create(ctx context.Context, req *CreateReq) error

	// Delete 删除角色（存在子角色或关联用户时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq) error

	// Update 更新角色
	Update(ctx context.Context, req *UpdateReq) error

	// UpdateStatus 修改角色状态
	UpdateStatus(ctx context.Context, req *StatusReq) error

	// Detail 查询角色详情（含已分配菜单ID）
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 返回角色选项树（用于下拉选择）
	Options(ctx context.Context) ([]*OptionResp, error)

	// List 查询角色列表（返回树形结构）
	List(ctx context.Context, req *ListReq) ([]*TreeResp, error)

	// AssignMenus 分配角色菜单权限
	AssignMenus(ctx context.Context, req *AssignMenusReq) error

	// AssignDepts 分配角色自定义部门（DataScope=2 时使用）
	AssignDepts(ctx context.Context, req *AssignDeptsReq) error
}
