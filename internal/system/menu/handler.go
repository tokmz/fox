package menu

import (
	"github.com/tokmz/qi"
)

// Handler 菜单处理器
type Handler struct {
	svc Service
}

// NewHandler 创建菜单处理器实例
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ===== 响应包装 =====

// listResp 列表响应（包装切片为结构体，满足 Bind 签名）
type listResp struct {
	List []*TreeResp `json:"list" desc:"菜单树形列表"`
}

// optionsResp 选项响应（包装切片为结构体）
type optionsResp struct {
	List []*OptionResp `json:"list" desc:"菜单选项列表"`
}

// ===== Handler 方法 =====

// Create 创建菜单
// POST /system/menu/create  body: CreateReq
func (h *Handler) Create(c *qi.Context, req *CreateReq) error {
	return h.svc.Create(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// Delete 删除菜单（支持批量）
// POST /system/menu/delete  body: DeleteReq
func (h *Handler) Delete(c *qi.Context, req *DeleteReq) error {
	return h.svc.Delete(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// Update 更新菜单
// POST /system/menu/update  body: UpdateReq
func (h *Handler) Update(c *qi.Context, req *UpdateReq) error {
	return h.svc.Update(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// UpdateStatus 修改菜单状态
// POST /system/menu/status  body: StatusReq
func (h *Handler) UpdateStatus(c *qi.Context, req *StatusReq) error {
	return h.svc.UpdateStatus(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// Detail 获取菜单详情
// GET /system/menu/detail?id=1
func (h *Handler) Detail(c *qi.Context, req *DetailReq) (*DetailResp, error) {
	return h.svc.Detail(c.Request().Context(), req)
}

// List 获取菜单列表（树形结构）
// GET /system/menu/list
func (h *Handler) List(c *qi.Context, req *ListReq) (*listResp, error) {
	items, err := h.svc.List(c.Request().Context(), req)
	if err != nil {
		return nil, err
	}
	return &listResp{List: items}, nil
}

// Options 获取菜单选项列表（树形结构，用于权限分配）
// GET /system/menu/options
func (h *Handler) Options(c *qi.Context) (*optionsResp, error) {
	opts, err := h.svc.Options(c.Request().Context())
	if err != nil {
		return nil, err
	}
	return &optionsResp{List: opts}, nil
}

// AssignApis 分配菜单API权限
// POST /system/menu/assign-apis  body: AssignApisReq
func (h *Handler) AssignApis(c *qi.Context, req *AssignApisReq) error {
	return h.svc.AssignApis(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}
