package post

// ===== 请求 DTO =====

// CreateReq 创建岗位请求
type CreateReq struct {
	DeptID int64  `json:"dept_id" binding:"required,min=1" desc:"所属部门ID" example:"1"`
	Name   string `json:"name" binding:"required,min=1,max=64" desc:"岗位名称" example:"总经理"`
	Code   string `json:"code" binding:"required,min=1,max=64" desc:"岗位编码" example:"CEO"`
	Sort   int    `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）" example:"1"`
	Remark string `json:"remark" binding:"omitempty,max=256" desc:"备注" example:"公司最高管理岗位"`
}

// UpdateReq 更新岗位请求
type UpdateReq struct {
	ID     int64  `json:"id" binding:"required,min=1" desc:"岗位ID" example:"1"`
	DeptID int64  `json:"dept_id" binding:"required,min=1" desc:"所属部门ID" example:"1"`
	Name   string `json:"name" binding:"omitempty,min=1,max=64" desc:"岗位名称" example:"总经理"`
	Code   string `json:"code" binding:"omitempty,min=1,max=64" desc:"岗位编码" example:"CEO"`
	Sort   int    `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）" example:"1"`
	Remark string `json:"remark" binding:"omitempty,max=256" desc:"备注" example:"公司最高管理岗位"`
}

// StatusReq 修改岗位状态请求
type StatusReq struct {
	IDs    []int64 `json:"ids" binding:"required,min=1" desc:"岗位ID列表" example:"1,2"`
	Status int8    `json:"status" binding:"required,oneof=0 1" desc:"状态：1启用 0禁用" example:"1"`
}

// DeleteReq 删除岗位请求
type DeleteReq struct {
	IDs   []int64 `json:"ids" binding:"required,min=1" desc:"岗位ID列表"`
	Force bool    `json:"force" desc:"是否强制删除（忽略用户关联）"`
}

// DetailReq 查询岗位详情请求
type DetailReq struct {
	ID int64 `json:"id" form:"id" binding:"required,min=1" desc:"岗位ID" example:"1"`
}

// OptionsReq 岗位选项列表请求
type OptionsReq struct {
	DeptID int64 `json:"dept_id" form:"id" binding:"required,min=1" desc:"部门ID" example:"1"`
}

// ===== 响应 DTO =====

// DetailResp 岗位详情响应
type DetailResp struct {
	ID        int64  `json:"id" desc:"岗位ID"`
	DeptID    int64  `json:"dept_id" desc:"所属部门ID"`
	Name      string `json:"name" desc:"岗位名称"`
	Code      string `json:"code" desc:"岗位编码"`
	Sort      int    `json:"sort" desc:"排序"`
	Remark    string `json:"remark" desc:"备注"`
	Status    int8   `json:"status" desc:"状态：1启用 0禁用"`
	CreatedBy int64  `json:"created_by" desc:"创建人ID"`
	UpdatedBy int64  `json:"updated_by" desc:"更新人ID"`
	CreatedAt string `json:"created_at" desc:"创建时间"`
	UpdatedAt string `json:"updated_at" desc:"更新时间"`
}

// PostOptionItemResp 岗位选项项响应
type PostOptionItemResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
