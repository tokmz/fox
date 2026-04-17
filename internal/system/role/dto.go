package role

// ===== 请求 DTO =====

// CreateReq 创建角色请求
type CreateReq struct {
	ParentID          *int64  `json:"parent_id" binding:"omitempty" desc:"父角色ID，nil表示顶级"`
	Name              string  `json:"name" binding:"required,max=64" desc:"角色名称" example:"管理员"`
	Code              string  `json:"code" binding:"required,max=64" desc:"角色编码" example:"admin"`
	DataScope         int8    `json:"data_scope" binding:"required,oneof=1 2 3 4 5" desc:"数据权限范围: 1=全部 2=自定义 3=本部门 4=本部门及下级 5=仅本人" example:"5"`
	DeptCheckStrictly *bool   `json:"dept_check_strictly" binding:"omitempty" desc:"部门树父子节点是否联动"`
	Sort              int     `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）" example:"0"`
	Status            *int8   `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
	MenuIDs           []int64 `json:"menu_ids" binding:"omitempty" desc:"已分配菜单ID列表"`
}

// UpdateReq 更新角色请求
type UpdateReq struct {
	ID                int64  `json:"id" binding:"required" desc:"角色ID"`
	ParentID          *int64 `json:"parent_id" binding:"omitempty" desc:"父角色ID，nil表示顶级"`
	Name              string `json:"name" binding:"omitempty,max=64" desc:"角色名称"`
	Code              string `json:"code" binding:"omitempty,max=64" desc:"角色编码"`
	DataScope         *int8  `json:"data_scope" binding:"omitempty,oneof=1 2 3 4 5" desc:"数据权限范围: 1=全部 2=自定义 3=本部门 4=本部门及下级 5=仅本人"`
	DeptCheckStrictly *bool  `json:"dept_check_strictly" binding:"omitempty" desc:"部门树父子节点是否联动"`
	Sort              *int   `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）"`
	Status            *int8  `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// StatusReq 修改角色状态请求
type StatusReq struct {
	ID     int64 `json:"id" binding:"required" desc:"角色ID"`
	Status int8  `json:"status" binding:"required,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
}

// DeleteReq 删除角色请求（支持批量删除）
type DeleteReq struct {
	IDs []int64 `json:"ids" binding:"required,min=1" desc:"角色ID列表"`
}

// DetailReq 查询角色详情请求
type DetailReq struct {
	ID int64 `json:"id" form:"id" binding:"required,min=1" desc:"角色ID" example:"1"`
}

// ListReq 查询角色列表请求
type ListReq struct {
	Name   string `json:"name" form:"name" binding:"omitempty,max=64" desc:"角色名称（模糊搜索）"`
	Code   string `json:"code" form:"code" binding:"omitempty,max=64" desc:"角色编码（模糊搜索）"`
	Status *int8  `json:"status" form:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// AssignMenusReq 分配角色菜单权限请求
type AssignMenusReq struct {
	RoleID  int64   `json:"role_id" binding:"required" desc:"角色ID"`
	MenuIDs []int64 `json:"menu_ids" binding:"required" desc:"菜单ID列表（全量替换）"`
}

// AssignDeptsReq 分配角色自定义部门请求
type AssignDeptsReq struct {
	RoleID  int64   `json:"role_id" binding:"required" desc:"角色ID"`
	DeptIDs []int64 `json:"dept_ids" binding:"required" desc:"部门ID列表（全量替换）"`
}

// ===== 响应 DTO =====

// DetailResp 角色详情响应（含已分配菜单ID）
type DetailResp struct {
	ID                int64    `json:"id" desc:"角色ID"`
	ParentID          *int64   `json:"parent_id" desc:"父角色ID"`
	Name              string   `json:"name" desc:"角色名称"`
	Code              string   `json:"code" desc:"角色编码"`
	DataScope         int8     `json:"data_scope" desc:"数据权限范围: 1=全部 2=自定义 3=本部门 4=本部门及下级 5=仅本人"`
	DeptCheckStrictly bool     `json:"dept_check_strictly" desc:"部门树父子节点是否联动"`
	Builtin           bool     `json:"builtin" desc:"是否内置角色"`
	Sort              int      `json:"sort" desc:"排序"`
	Status            int8     `json:"status" desc:"状态: 1=启用 0=禁用"`
	MenuIDs           []int64  `json:"menu_ids" desc:"已分配菜单ID列表"`
	DeptIDs           []int64  `json:"dept_ids" desc:"已分配部门ID列表"`
	CreatedBy         int64    `json:"created_by" desc:"创建人ID"`
	UpdatedBy         int64    `json:"updated_by" desc:"更新人ID"`
	CreatedAt         string   `json:"created_at" desc:"创建时间"`
	UpdatedAt         string   `json:"updated_at" desc:"更新时间"`
}

// TreeResp 角色树形节点
type TreeResp struct {
	ID        int64       `json:"id" desc:"角色ID"`
	ParentID  *int64      `json:"parent_id" desc:"父角色ID"`
	Name      string      `json:"name" desc:"角色名称"`
	Code      string      `json:"code" desc:"角色编码"`
	DataScope int8        `json:"data_scope" desc:"数据权限范围: 1=全部 2=自定义 3=本部门 4=本部门及下级 5=仅本人"`
	Sort      int         `json:"sort" desc:"排序"`
	Status    int8        `json:"status" desc:"状态: 1=启用 0=禁用"`
	Builtin   bool        `json:"builtin" desc:"是否内置角色"`
	Children  []*TreeResp `json:"children" desc:"子角色列表"`
}

// OptionResp 角色选项（下拉选择用）
type OptionResp struct {
	ID   int64  `json:"id" desc:"角色ID"`
	Name string `json:"name" desc:"角色名称"`
	Code string `json:"code" desc:"角色编码"`
}
