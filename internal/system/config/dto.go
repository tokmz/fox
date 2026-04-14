package config

// ===== 分组 DTO =====

// CreateGroupReq 创建分组请求
type CreateGroupReq struct{}

// GetGroupReq 获取分组请求
type GetGroupReq struct{}

// GetGroupResp 获取分组响应（含配置项列表）
type GetGroupResp struct{}

// DeleteGroupReq 删除分组请求
type DeleteGroupReq struct{}

// ===== 配置项 DTO =====

// CreateItemReq 创建配置项请求
type CreateItemReq struct{}

// DeleteItemReq 删除配置项请求
type DeleteItemReq struct{}

// ListItemsReq 配置项列表请求（根据分组查询）
type ListItemsReq struct{}

// ItemResp 配置项响应
type ItemResp struct{}
