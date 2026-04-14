package dept

import "context"

// Service 部门服务接口
type Service interface {
	// Create 创建部门
	Create(ctx context.Context, req *CreateReq) error

	// Delete 删除部门（存在子部门时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq) error

	// Update 更新部门
	Update(ctx context.Context, req *UpdateReq) error

	// UpdateStatus 修改部门状态
	UpdateStatus(ctx context.Context, req *StatusReq) error

	// Detail 查询部门详情
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 返回部门选项树（用于下拉选择）
	Options(ctx context.Context) ([]*OptionResp, error)

	// List 查询部门列表（返回树形结构）
	List(ctx context.Context, req *ListReq) ([]*TreeResp, error)
}
