package role

// ===== 请求 DTO =====

// CreateReq 创建角色请求
type CreateReq struct{}

// DeleteReq 删除角色请求
type DeleteReq struct{}

// UpdateReq 更新角色请求
type UpdateReq struct{}

// StatusReq 修改角色状态请求
type StatusReq struct{}

// DetailReq 查询角色详情请求
type DetailReq struct{}

// ListReq 查询角色列表请求
type ListReq struct{}

// AssignMenusReq 分配角色菜单权限请求
type AssignMenusReq struct{}

// AssignDeptsReq 分配角色自定义部门请求
type AssignDeptsReq struct{}

// ===== 响应 DTO =====

// DetailResp 角色详情响应（含已分配菜单ID）
type DetailResp struct{}

// OptionResp 角色选项（下拉选择用）
type OptionResp struct{}

// TreeResp 角色树形节点
type TreeResp struct{}
