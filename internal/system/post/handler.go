package post

import (
	"github.com/tokmz/qi"
)

// Handler 岗位 HTTP 处理器
type Handler struct {
	svc Service
}

// NewHandler 创建岗位处理器
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ===== 响应包装 =====

// optionsResp 选项响应（包装切片为结构体，满足 Bind 签名）
type optionsResp struct {
	List []*PostOptionItemResp `json:"list" desc:"岗位选项列表"`
}

// ===== Handler 方法 =====

// Create 创建岗位
// POST /posts  body: CreateReq
func (h *Handler) Create(c *qi.Context, req *CreateReq) error {
	return h.svc.Create(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Delete 批量删除岗位
// POST /posts/delete  body: DeleteReq
func (h *Handler) Delete(c *qi.Context, req *DeleteReq) error {
	return h.svc.Delete(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Update 更新岗位
// POST /posts/update  body: UpdateReq
func (h *Handler) Update(c *qi.Context, req *UpdateReq) error {
	return h.svc.Update(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// UpdateStatus 修改岗位状态
// POST /posts/status  body: StatusReq
func (h *Handler) UpdateStatus(c *qi.Context, req *StatusReq) error {
	return h.svc.UpdateStatus(c.Request().Context(), req, c.Gin().GetInt64("uid"))
}

// Detail 查询岗位详情
// GET /posts/detail?id=1
func (h *Handler) Detail(c *qi.Context, req *DetailReq) (*DetailResp, error) {
	return h.svc.Detail(c.Request().Context(), req)
}

// Options 查询岗位选项列表
// GET /posts/options?id=1
func (h *Handler) Options(c *qi.Context, req *OptionsReq) (*optionsResp, error) {
	items, err := h.svc.Options(c.Request().Context(), req)
	if err != nil {
		return nil, err
	}
	return &optionsResp{List: items}, nil
}
