package role

import (
	"github.com/tokmz/qi"
)

// Handler 角色 HTTP 处理器
type Handler struct {
	svc Service
}

// NewHandler 创建角色处理器
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ===== 响应包装 =====

// listResp 列表响应（包装切片为结构体，满足 Bind 签名）
type listResp struct {
	List []*TreeResp `json:"list" desc:"角色树形列表"`
}

// optionsResp 选项响应（包装切片为结构体）
type optionsResp struct {
	List []*OptionResp `json:"list" desc:"角色选项列表"`
}

// ===== Handler 方法 =====

// Create 创建角色
// POST /roles  body: CreateReq
func (h *Handler) Create(c *qi.Context, req *CreateReq) error {
	return h.svc.Create(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Delete 批量删除角色
// POST /roles/delete  body: DeleteReq
func (h *Handler) Delete(c *qi.Context, req *DeleteReq) error {
	return h.svc.Delete(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Update 更新角色
// POST /roles/update  body: UpdateReq
func (h *Handler) Update(c *qi.Context, req *UpdateReq) error {
	return h.svc.Update(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// UpdateStatus 修改角色状态
// POST /roles/status  body: StatusReq
func (h *Handler) UpdateStatus(c *qi.Context, req *StatusReq) error {
	return h.svc.UpdateStatus(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Detail 查询角色详情
// GET /roles/detail?id=1
func (h *Handler) Detail(c *qi.Context, req *DetailReq) (*DetailResp, error) {
	return h.svc.Detail(c.Request().Context(), req)
}

// List 查询角色列表（树形结构）
// GET /roles/list?name=xxx
func (h *Handler) List(c *qi.Context, req *ListReq) (*listResp, error) {
	trees, err := h.svc.List(c.Request().Context(), req)
	if err != nil {
		return nil, err
	}
	return &listResp{List: trees}, nil
}

// Options 返回角色选项列表
// GET /roles/options
func (h *Handler) Options(c *qi.Context) (*optionsResp, error) {
	opts, err := h.svc.Options(c.Request().Context())
	if err != nil {
		return nil, err
	}
	return &optionsResp{List: opts}, nil
}

// AssignMenus 分配角色菜单权限
// POST /roles/menus  body: AssignMenusReq
func (h *Handler) AssignMenus(c *qi.Context, req *AssignMenusReq) error {
	return h.svc.AssignMenus(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// AssignDepts 分配角色自定义部门
// POST /roles/depts  body: AssignDeptsReq
func (h *Handler) AssignDepts(c *qi.Context, req *AssignDeptsReq) error {
	return h.svc.AssignDepts(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}
