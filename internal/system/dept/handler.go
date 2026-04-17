package dept

import (
	"github.com/tokmz/qi"
)

// Handler 部门 HTTP 处理器
type Handler struct {
	svc Service
}

// NewHandler 创建部门处理器
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ===== 响应包装 =====

// listResp 列表响应（包装切片为结构体，满足 Bind 签名）
type listResp struct {
	List []*TreeResp `json:"list" desc:"部门树形列表"`
}

// optionsResp 选项响应（包装切片为结构体）
type optionsResp struct {
	List []*OptionResp `json:"list" desc:"部门选项列表"`
}

// ===== Handler 方法 =====

// Create 创建部门
// POST /depts  body: CreateReq
func (h *Handler) Create(c *qi.Context, req *CreateReq) error {
	return h.svc.Create(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Delete 批量删除部门
// POST /depts/delete  body: DeleteReq
func (h *Handler) Delete(c *qi.Context, req *DeleteReq) error {
	return h.svc.Delete(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Update 更新部门
// POST /depts/update  body: UpdateReq
func (h *Handler) Update(c *qi.Context, req *UpdateReq) error {
	return h.svc.Update(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// UpdateStatus 修改部门状态
// POST /depts/status  body: StatusReq
func (h *Handler) UpdateStatus(c *qi.Context, req *StatusReq) error {
	return h.svc.UpdateStatus(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Detail 查询部门详情
// GET /depts/detail?id=1
func (h *Handler) Detail(c *qi.Context, req *DetailReq) (*DetailResp, error) {
	return h.svc.Detail(c.Request().Context(), req)
}

// List 查询部门列表（树形结构）
// GET /depts/list?name=xxx
func (h *Handler) List(c *qi.Context, req *ListReq) (*listResp, error) {
	trees, err := h.svc.List(c.Request().Context(), req)
	if err != nil {
		return nil, err
	}
	return &listResp{List: trees}, nil
}

// Options 返回部门选项列表
// GET /depts/options
func (h *Handler) Options(c *qi.Context) (*optionsResp, error) {
	opts, err := h.svc.Options(c.Request().Context())
	if err != nil {
		return nil, err
	}
	return &optionsResp{List: opts}, nil
}
