package menu

// ===== 请求 DTO =====

// CreateReq 创建菜单请求
type CreateReq struct {
	ParentID     *int64  `json:"parent_id" binding:"omitempty" desc:"父菜单ID，nil表示顶级"`
	Title        string  `json:"title" binding:"required,max=64" desc:"菜单名称" example:"系统管理"`
	Key          string  `json:"key" binding:"required,max=64" desc:"路由name（唯一标识）" example:"system"`
	Path         string  `json:"path" binding:"omitempty,max=256" desc:"路由path" example:"/system"`
	Component    string  `json:"component" binding:"omitempty,max=256" desc:"组件路径" example:"LAYOUT"`
	Redirect     string  `json:"redirect" binding:"omitempty,max=256" desc:"重定向路径"`
	Query        string  `json:"query" binding:"omitempty,max=256" desc:"路由参数JSON"`
	MenuType     int8    `json:"menu_type" binding:"required,oneof=1 2 3" desc:"菜单类型: 1=目录 2=页面 3=按钮" example:"1"`
	OpenType     int8    `json:"open_type" binding:"omitempty,oneof=1 2 3" desc:"打开方式: 1=组件 2=内嵌iframe 3=外链" example:"1"`
	Icon         string  `json:"icon" binding:"omitempty,max=128" desc:"图标名称" example:"icon-system"`
	Sort         int     `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）" example:"0"`
	KeepAlive    *int8   `json:"keep_alive" binding:"omitempty,oneof=0 1" desc:"是否缓存页面: 0=否 1=是" example:"0"`
	Hidden       *int8   `json:"hidden" binding:"omitempty,oneof=0 1" desc:"是否在菜单中隐藏: 0=否 1=是" example:"0"`
	Affix        *int8   `json:"affix" binding:"omitempty,oneof=0 1" desc:"是否固定在标签页: 0=否 1=是" example:"0"`
	AlwaysShow   *int8   `json:"always_show" binding:"omitempty,oneof=0 1" desc:"是否强制显示根路由: 0=否 1=是" example:"0"`
	ActiveMenu   string  `json:"active_menu" binding:"omitempty,max=64" desc:"高亮指定菜单的route_name"`
	FrameSrc     string  `json:"frame_src" binding:"omitempty,max=512" desc:"iframe地址（OpenType=2）"`
	ExternalLink string  `json:"external_link" binding:"omitempty,max=512" desc:"外链地址（OpenType=3）"`
	Remark       string  `json:"remark" binding:"omitempty,max=255" desc:"备注"`
	Status       *int8   `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
	ApiIDs       []int64 `json:"api_ids" binding:"omitempty" desc:"已分配API权限ID列表"`
}

// UpdateReq 更新菜单请求
type UpdateReq struct {
	ID           int64   `json:"id" binding:"required" desc:"菜单ID"`
	ParentID     *int64  `json:"parent_id" binding:"omitempty" desc:"父菜单ID，nil表示顶级"`
	Title        string  `json:"title" binding:"omitempty,max=64" desc:"菜单名称"`
	Key          string  `json:"key" binding:"omitempty,max=64" desc:"路由name（唯一标识）"`
	Path         string  `json:"path" binding:"omitempty,max=256" desc:"路由path"`
	Component    string  `json:"component" binding:"omitempty,max=256" desc:"组件路径"`
	Redirect     string  `json:"redirect" binding:"omitempty,max=256" desc:"重定向路径"`
	Query        string  `json:"query" binding:"omitempty,max=256" desc:"路由参数JSON"`
	MenuType     *int8   `json:"menu_type" binding:"omitempty,oneof=1 2 3" desc:"菜单类型: 1=目录 2=页面 3=按钮"`
	OpenType     *int8   `json:"open_type" binding:"omitempty,oneof=1 2 3" desc:"打开方式: 1=组件 2=内嵌iframe 3=外链"`
	Icon         string  `json:"icon" binding:"omitempty,max=128" desc:"图标名称"`
	Sort         *int    `json:"sort" binding:"omitempty,min=0" desc:"排序（升序）"`
	KeepAlive    *int8   `json:"keep_alive" binding:"omitempty,oneof=0 1" desc:"是否缓存页面: 0=否 1=是"`
	Hidden       *int8   `json:"hidden" binding:"omitempty,oneof=0 1" desc:"是否在菜单中隐藏: 0=否 1=是"`
	Affix        *int8   `json:"affix" binding:"omitempty,oneof=0 1" desc:"是否固定在标签页: 0=否 1=是"`
	AlwaysShow   *int8   `json:"always_show" binding:"omitempty,oneof=0 1" desc:"是否强制显示根路由: 0=否 1=是"`
	ActiveMenu   string  `json:"active_menu" binding:"omitempty,max=64" desc:"高亮指定菜单的route_name"`
	FrameSrc     string  `json:"frame_src" binding:"omitempty,max=512" desc:"iframe地址（OpenType=2）"`
	ExternalLink string  `json:"external_link" binding:"omitempty,max=512" desc:"外链地址（OpenType=3）"`
	Remark       string  `json:"remark" binding:"omitempty,max=255" desc:"备注"`
	Status       *int8   `json:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// StatusReq 修改菜单状态请求
type StatusReq struct {
	ID     int64 `json:"id" binding:"required" desc:"菜单ID"`
	Status int8  `json:"status" binding:"required,oneof=0 1" desc:"状态: 1=启用 0=禁用" example:"1"`
}

// DeleteReq 删除菜单请求（支持批量删除）
type DeleteReq struct {
	IDs []int64 `json:"ids" binding:"required,min=1" desc:"菜单ID列表"`
}

// DetailReq 查询菜单详情请求
type DetailReq struct {
	ID int64 `json:"id" form:"id" binding:"required,min=1" desc:"菜单ID" example:"1"`
}

// ListReq 查询菜单列表请求
type ListReq struct {
	Title    string `json:"title" form:"title" binding:"omitempty,max=64" desc:"菜单名称（模糊搜索）"`
	Key      string `json:"key" form:"key" binding:"omitempty,max=64" desc:"路由name（模糊搜索）"`
	MenuType *int8  `json:"menu_type" form:"menu_type" binding:"omitempty,oneof=1 2 3" desc:"菜单类型: 1=目录 2=页面 3=按钮"`
	Status   *int8  `json:"status" form:"status" binding:"omitempty,oneof=0 1" desc:"状态: 1=启用 0=禁用"`
}

// AssignApisReq 分配菜单API权限请求
type AssignApisReq struct {
	MenuID int64   `json:"menu_id" binding:"required" desc:"菜单ID"`
	ApiIDs []int64 `json:"api_ids" binding:"required" desc:"API ID列表（全量替换）"`
}

// ===== 响应 DTO =====

// DetailResp 菜单详情响应（含已分配API ID）
type DetailResp struct {
	ID           int64   `json:"id" desc:"菜单ID"`
	ParentID     *int64  `json:"parent_id" desc:"父菜单ID"`
	Title        string  `json:"title" desc:"菜单名称"`
	Key          string  `json:"key" desc:"路由name（唯一标识）"`
	Path         string  `json:"path" desc:"路由path"`
	Component    string  `json:"component" desc:"组件路径"`
	Redirect     string  `json:"redirect" desc:"重定向路径"`
	Query        string  `json:"query" desc:"路由参数JSON"`
	MenuType     int8    `json:"menu_type" desc:"菜单类型: 1=目录 2=页面 3=按钮"`
	OpenType     int8    `json:"open_type" desc:"打开方式: 1=组件 2=内嵌iframe 3=外链"`
	Icon         string  `json:"icon" desc:"图标名称"`
	Sort         int     `json:"sort" desc:"排序"`
	KeepAlive    int8    `json:"keep_alive" desc:"是否缓存页面: 0=否 1=是"`
	Hidden       int8    `json:"hidden" desc:"是否在菜单中隐藏: 0=否 1=是"`
	Affix        int8    `json:"affix" desc:"是否固定在标签页: 0=否 1=是"`
	AlwaysShow   int8    `json:"always_show" desc:"是否强制显示根路由: 0=否 1=是"`
	ActiveMenu   string  `json:"active_menu" desc:"高亮指定菜单的route_name"`
	FrameSrc     string  `json:"frame_src" desc:"iframe地址"`
	ExternalLink string  `json:"external_link" desc:"外链地址"`
	Remark       string  `json:"remark" desc:"备注"`
	Status       int8    `json:"status" desc:"状态: 1=启用 0=禁用"`
	ApiIDs       []int64 `json:"api_ids" desc:"已分配API权限ID列表"`
	CreatedAt    string  `json:"created_at" desc:"创建时间"`
	UpdatedAt    string  `json:"updated_at" desc:"更新时间"`
}

// TreeResp 菜单树形节点
type TreeResp struct {
	ID           int64       `json:"id" desc:"菜单ID"`
	ParentID     *int64      `json:"parent_id" desc:"父菜单ID"`
	Title        string      `json:"title" desc:"菜单名称"`
	Key          string      `json:"key" desc:"路由name"`
	Path         string      `json:"path" desc:"路由path"`
	Component    string      `json:"component" desc:"组件路径"`
	Redirect     string      `json:"redirect" desc:"重定向路径"`
	Icon         string      `json:"icon" desc:"图标名称"`
	MenuType     int8        `json:"menu_type" desc:"菜单类型: 1=目录 2=页面 3=按钮"`
	OpenType     int8        `json:"open_type" desc:"打开方式: 1=组件 2=内嵌iframe 3=外链"`
	Sort         int         `json:"sort" desc:"排序"`
	KeepAlive    int8        `json:"keep_alive" desc:"是否缓存页面"`
	Hidden       int8        `json:"hidden" desc:"是否隐藏"`
	Affix        int8        `json:"affix" desc:"是否固定标签页"`
	AlwaysShow   int8        `json:"always_show" desc:"是否强制显示根路由"`
	ActiveMenu   string      `json:"active_menu" desc:"高亮菜单"`
	FrameSrc     string      `json:"frame_src" desc:"iframe地址"`
	ExternalLink string      `json:"external_link" desc:"外链地址"`
	Status       int8        `json:"status" desc:"状态: 1=启用 0=禁用"`
	Children     []*TreeResp `json:"children" desc:"子菜单列表"`
}

// OptionResp 菜单选项（权限分配用）
type OptionResp struct {
	ID       int64          `json:"id" desc:"菜单ID"`
	ParentID *int64         `json:"parent_id" desc:"父菜单ID"`
	Title    string         `json:"title" desc:"菜单名称"`
	Key      string         `json:"key" desc:"路由name"`
	MenuType int8           `json:"menu_type" desc:"菜单类型: 1=目录 2=页面 3=按钮"`
	Children []*OptionResp  `json:"children" desc:"子菜单选项"`
}
