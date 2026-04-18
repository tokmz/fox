package user

import "github.com/tokmz/fox/pkg/query"

// ===== 请求 DTO =====

// CreateReq 创建用户请求
type CreateReq struct {
	Username string  `json:"username" binding:"required,min=3,max=64" desc:"用户名" example:"zhangsan"`
	Password string  `json:"password" binding:"required,min=6,max=32" desc:"密码" example:"123456"`
	Nickname string  `json:"nickname" binding:"required,min=1,max=64" desc:"昵称" example:"张三"`
	Email    string  `json:"email" binding:"omitempty,email,max=128" desc:"邮箱" example:"zhangsan@example.com"`
	Phone    string  `json:"phone" binding:"omitempty,max=32" desc:"手机号" example:"13800138000"`
	Avatar   string  `json:"avatar" binding:"omitempty,max=512" desc:"头像地址"`
	Gender   *int8   `json:"gender" binding:"omitempty,oneof=0 1 2" desc:"性别: 0=未知 1=男 2=女" example:"1"`
	DeptID   int64   `json:"dept_id" binding:"required,min=1" desc:"所属部门ID" example:"1"`
	Remark   string  `json:"remark" binding:"omitempty,max=256" desc:"备注"`
	Status   *int8   `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
	RoleIDs  []int64 `json:"role_ids" binding:"omitempty" desc:"角色ID列表"`
	PostIDs  []int64 `json:"post_ids" binding:"omitempty" desc:"岗位ID列表"`
}

// UpdateReq 更新用户请求
type UpdateReq struct {
	ID       int64   `json:"id" binding:"required" desc:"用户ID"`
	Username string  `json:"username" binding:"omitempty,min=3,max=64" desc:"用户名"`
	Nickname string  `json:"nickname" binding:"omitempty,min=1,max=64" desc:"昵称"`
	Email    string  `json:"email" binding:"omitempty,email,max=128" desc:"邮箱"`
	Phone    string  `json:"phone" binding:"omitempty,max=32" desc:"手机号"`
	Avatar   string  `json:"avatar" binding:"omitempty,max=512" desc:"头像地址"`
	Gender   *int8   `json:"gender" binding:"omitempty,oneof=0 1 2" desc:"性别: 0=未知 1=男 2=女"`
	DeptID   *int64  `json:"dept_id" binding:"omitempty,min=1" desc:"所属部门ID"`
	Remark   string  `json:"remark" binding:"omitempty,max=256" desc:"备注"`
	Status   *int8   `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// StatusReq 修改用户状态请求
type StatusReq struct {
	IDs    []int64 `json:"ids" binding:"required,min=1" desc:"用户ID列表" example:"1,2"`
	Status int8    `json:"status" binding:"required,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
}

// DeleteReq 删除用户请求（支持批量删除）
type DeleteReq struct {
	IDs []int64 `json:"ids" binding:"required,min=1" desc:"用户ID列表"`
}

// DetailReq 查询用户详情请求
type DetailReq struct {
	ID int64 `json:"id" form:"id" binding:"required,min=1" desc:"用户ID" example:"1"`
}

// ListReq 查询用户列表请求（分页）
type ListReq struct {
	query.PageReq
	Username string `json:"username" form:"username" binding:"omitempty,max=64" desc:"用户名（模糊搜索）"`
	Nickname string `json:"nickname" form:"nickname" binding:"omitempty,max=64" desc:"昵称（模糊搜索）"`
	Phone    string `json:"phone" form:"phone" binding:"omitempty,max=32" desc:"手机号（精确搜索）"`
	Email    string `json:"email" form:"email" binding:"omitempty,max=128" desc:"邮箱（精确搜索）"`
	DeptID   *int64 `json:"dept_id" form:"dept_id" binding:"omitempty,min=1" desc:"部门ID"`
	Status   *int8  `json:"status" form:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// ResetPasswordReq 重置用户密码请求（管理员操作）
type ResetPasswordReq struct {
	ID          int64  `json:"id" binding:"required" desc:"用户ID"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32" desc:"新密码" example:"123456"`
}

// AssignRolesReq 分配用户角色请求
type AssignRolesReq struct {
	UserID  int64   `json:"user_id" binding:"required" desc:"用户ID"`
	RoleIDs []int64 `json:"role_ids" binding:"required" desc:"角色ID列表（全量替换）"`
}

// AssignPostsReq 分配用户岗位请求
type AssignPostsReq struct {
	UserID  int64   `json:"user_id" binding:"required" desc:"用户ID"`
	PostIDs []int64 `json:"post_ids" binding:"required" desc:"岗位ID列表（全量替换）"`
}

// ===== 响应 DTO =====

// DetailResp 用户详情响应（含关联数据）
type DetailResp struct {
	ID        int64    `json:"id" desc:"用户ID"`
	Username  string   `json:"username" desc:"用户名"`
	Nickname  string   `json:"nickname" desc:"昵称"`
	Email     string   `json:"email" desc:"邮箱"`
	Phone     string   `json:"phone" desc:"手机号"`
	Avatar    string   `json:"avatar" desc:"头像地址"`
	Gender    int8     `json:"gender" desc:"性别: 0=未知 1=男 2=女"`
	DeptID    int64    `json:"dept_id" desc:"所属部门ID"`
	DeptName  string   `json:"dept_name" desc:"部门名称"`
	Remark    string   `json:"remark" desc:"备注"`
	Status    int8     `json:"status" desc:"状态: 1=启用 0=禁用"`
	RoleIDs   []int64  `json:"role_ids" desc:"角色ID列表"`
	RoleNames []string `json:"role_names" desc:"角色名称列表"`
	PostIDs   []int64  `json:"post_ids" desc:"岗位ID列表"`
	PostNames []string `json:"post_names" desc:"岗位名称列表"`
	CreatedBy int64    `json:"created_by" desc:"创建人ID"`
	UpdatedBy int64    `json:"updated_by" desc:"更新人ID"`
	CreatedAt string   `json:"created_at" desc:"创建时间"`
	UpdatedAt string   `json:"updated_at" desc:"更新时间"`
}

// ListItemResp 用户列表项响应
type ListItemResp struct {
	ID        int64    `json:"id" desc:"用户ID"`
	Username  string   `json:"username" desc:"用户名"`
	Nickname  string   `json:"nickname" desc:"昵称"`
	Email     string   `json:"email" desc:"邮箱"`
	Phone     string   `json:"phone" desc:"手机号"`
	Avatar    string   `json:"avatar" desc:"头像地址"`
	Gender    int8     `json:"gender" desc:"性别: 0=未知 1=男 2=女"`
	DeptID    int64    `json:"dept_id" desc:"所属部门ID"`
	DeptName  string   `json:"dept_name" desc:"部门名称"`
	Status    int8     `json:"status" desc:"状态: 1=启用 0=禁用"`
	RoleNames []string `json:"role_names" desc:"角色名称列表"`
	PostNames []string `json:"post_names" desc:"岗位名称列表"`
	CreatedAt string   `json:"created_at" desc:"创建时间"`
}

// OptionResp 用户选项（下拉选择用）
type OptionResp struct {
	ID       int64  `json:"id" desc:"用户ID"`
	Username string `json:"username" desc:"用户名"`
	Nickname string `json:"nickname" desc:"昵称"`
	DeptName string `json:"dept_name" desc:"部门名称"`
}
