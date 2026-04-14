package dept

// ===== 请求 DTO =====

// CreateReq 创建部门请求
type CreateReq struct{}

// DeleteReq 删除部门请求
type DeleteReq struct{}

// UpdateReq 更新部门请求
type UpdateReq struct{}

// StatusReq 修改部门状态请求
type StatusReq struct{}

// ListReq 查询部门列表请求
type ListReq struct{}

// DetailReq 查询部门详情请求
type DetailReq struct{}

// ===== 响应 DTO =====

// TreeResp 部门树形节点
type TreeResp struct{}

// DetailResp 部门详情响应
type DetailResp struct{}

// OptionResp 部门选项（下拉选择用）
type OptionResp struct{}
