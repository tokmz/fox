package post

import "context"

// Service 岗位服务接口
type Service interface {
	// Create 创建岗位
	Create(ctx context.Context, req *CreateReq) error

	// Update 更新岗位
	Update(ctx context.Context, req *UpdateReq) error

	// UpdateStatus 修改岗位状态
	UpdateStatus(ctx context.Context, req *StatusReq) error

	// Delete 删除岗位（存在关联用户时拒绝删除）
	Delete(ctx context.Context, req *DeleteReq) error

	// Detail 查询岗位详情
	Detail(ctx context.Context, req *DetailReq) (*DetailResp, error)

	// Options 根据部门ID获取该部门下的岗位选项列表
	Options(ctx context.Context, req *OptionsReq) ([]*PostOptionItem, error)
}
