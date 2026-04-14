package menu

// ===== 请求 DTO =====

// CreateReq 创建菜单请求
type CreateReq struct{}

// DeleteReq 删除菜单请求
type DeleteReq struct{}

// UpdateReq 更新菜单请求
type UpdateReq struct{}

// StatusReq 修改菜单状态请求
type StatusReq struct{}

// DetailReq 查询菜单详情请求
type DetailReq struct{}

// ListReq 查询菜单列表请求
type ListReq struct{}

// ===== 响应 DTO =====

// DetailResp 菜单详情响应
type DetailResp struct{}

// OptionResp 菜单选项（权限分配用）
type OptionResp struct{}

// TreeResp 菜单树形节点
type TreeResp struct{}
