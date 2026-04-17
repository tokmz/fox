package dept

// ===== 请求 DTO =====

// CreateReq 创建部门请求
type CreateReq struct {
	ParentID *int64 `json:"parent_id" binding:"omitempty" desc:"父部门ID，nil表示顶级"`
	Name     string `json:"name" binding:"required,min=1,max=64" desc:"部门名称" example:"技术部"`
	Code     string `json:"code" binding:"required,min=1,max=64" desc:"部门编码" example:"TECH"`
	DeptType int8   `json:"dept_type" binding:"required,oneof=1 2 3" desc:"部门类型: 1=公司 2=部门 3=小组" example:"2"`
	LeaderID *int64 `json:"leader_id" binding:"omitempty" desc:"部门负责人用户ID"`
	Sort     int    `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）" example:"1"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
}

// UpdateReq 更新部门请求
type UpdateReq struct {
	ID       int64  `json:"id" binding:"required" desc:"部门ID"`
	ParentID *int64 `json:"parent_id" binding:"omitempty" desc:"父部门ID，nil表示顶级"`
	Name     string `json:"name" binding:"omitempty,min=1,max=64" desc:"部门名称"`
	Code     string `json:"code" binding:"omitempty,min=1,max=64" desc:"部门编码"`
	DeptType *int8  `json:"dept_type" binding:"omitempty,oneof=1 2 3" desc:"部门类型: 1=公司 2=部门 3=小组"`
	LeaderID *int64 `json:"leader_id" binding:"omitempty" desc:"部门负责人用户ID"`
	Sort     *int   `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）"`
	Status   *int8  `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// StatusReq 修改部门状态请求
type StatusReq struct {
	IDs    []int64 `json:"ids" binding:"required,min=1" desc:"部门ID列表" example:"1,2"`
	Status int8    `json:"status" binding:"required,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
}

// DeleteReq 删除部门请求
type DeleteReq struct {
	IDs []int64 `json:"ids" binding:"required,min=1" desc:"部门ID列表"`
}

// DetailReq 查询部门详情请求
type DetailReq struct {
	ID int64 `json:"id" form:"id" binding:"required,min=1" desc:"部门ID" example:"1"`
}

// ListReq 查询部门列表请求
type ListReq struct {
	Name     string `json:"name" form:"name" binding:"omitempty,max=64" desc:"部门名称（模糊搜索）"`
	Code     string `json:"code" form:"code" binding:"omitempty,max=64" desc:"部门编码（模糊搜索）"`
	DeptType *int8  `json:"dept_type" form:"dept_type" binding:"omitempty,oneof=1 2 3" desc:"部门类型: 1=公司 2=部门 3=小组"`
	Status   *int8  `json:"status" form:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// ===== 响应 DTO =====

// DetailResp 部门详情响应
type DetailResp struct {
	ID        int64  `json:"id" desc:"部门ID"`
	ParentID  *int64 `json:"parent_id" desc:"父部门ID"`
	Name      string `json:"name" desc:"部门名称"`
	Code      string `json:"code" desc:"部门编码"`
	DeptType  int8   `json:"dept_type" desc:"部门类型: 1=公司 2=部门 3=小组"`
	LeaderID  *int64 `json:"leader_id" desc:"部门负责人用户ID"`
	Sort      int    `json:"sort" desc:"排序"`
	Status    int8   `json:"status" desc:"状态: 1=启用 0=禁用"`
	CreatedBy int64  `json:"created_by" desc:"创建人ID"`
	UpdatedBy int64  `json:"updated_by" desc:"更新人ID"`
	CreatedAt string `json:"created_at" desc:"创建时间"`
	UpdatedAt string `json:"updated_at" desc:"更新时间"`
}

// TreeResp 部门树形节点
type TreeResp struct {
	ID        int64      `json:"id" desc:"部门ID"`
	ParentID  *int64     `json:"parent_id" desc:"父部门ID"`
	Name      string     `json:"name" desc:"部门名称"`
	Code      string     `json:"code" desc:"部门编码"`
	DeptType  int8       `json:"dept_type" desc:"部门类型: 1=公司 2=部门 3=小组"`
	LeaderID  *int64     `json:"leader_id" desc:"部门负责人用户ID"`
	Sort      int        `json:"sort" desc:"排序"`
	Status    int8       `json:"status" desc:"状态: 1=启用 0=禁用"`
	Children  []*TreeResp `json:"children" desc:"子部门列表"`
}

// OptionResp 部门选项（下拉选择用）
type OptionResp struct {
	ID   int64  `json:"id" desc:"部门ID"`
	Name string `json:"name" desc:"部门名称"`
	Code string `json:"code" desc:"部门编码"`
}
