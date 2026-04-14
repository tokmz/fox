package api

// ===== API 接口 DTO =====

// CreateApiReq 创建接口请求
type CreateApiReq struct{}

// UpdateApiReq 更新接口请求
type UpdateApiReq struct{}

// DeleteApiReq 删除接口请求
type DeleteApiReq struct{}

// DetailApiReq 查询接口详情请求
type DetailApiReq struct{}

// ListApiReq 查询接口列表请求
type ListApiReq struct{}

// DetailApiResp 接口详情响应
type DetailApiResp struct{}

// ApiItemResp  接口列表项
type ApiItemResp struct{}

// ===== Group 分组 DTO =====

// CreateGroupReq 创建分组请求
type CreateGroupReq struct{}

// UpdateGroupReq 更新分组请求
type UpdateGroupReq struct{}

// DeleteGroupReq 删除分组请求
type DeleteGroupReq struct{}

// GroupTreeReq 查询分组树请求
type GroupTreeReq struct{}

// GroupTreeResp 分组树形节点（含接口列表）
type GroupTreeResp struct{}
