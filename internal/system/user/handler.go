package user

import (
	"github.com/tokmz/qi"
)

// Handler 用户处理器
type Handler struct {
	svc Service
}

// NewHandler 创建用户处理器实例
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ===== 响应包装 =====

// listResp 列表响应（包装切片为结构体，满足 Bind 签名）
type listResp struct {
	List  []*ListItemResp `json:"list" desc:"用户列表"`
	Total int64           `json:"total" desc:"总数"`
}

// optionsResp 选项响应（包装切片为结构体）
type optionsResp struct {
	List []*OptionResp `json:"list" desc:"用户选项列表"`
}

// ===== Handler 方法 =====

// Create 创建用户
// POST /system/user/create  body: CreateReq
func (h *Handler) Create(c *qi.Context, req *CreateReq) error {
	return h.svc.Create(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// Delete 删除用户（支持批量）
// POST /system/user/delete  body: DeleteReq
func (h *Handler) Delete(c *qi.Context, req *DeleteReq) error {
	return h.svc.Delete(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// Update 更新用户
// POST /system/user/update  body: UpdateReq
func (h *Handler) Update(c *qi.Context, req *UpdateReq) error {
	return h.svc.Update(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// UpdateStatus 修改用户状态
// POST /system/user/status  body: StatusReq
func (h *Handler) UpdateStatus(c *qi.Context, req *StatusReq) error {
	return h.svc.UpdateStatus(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// Detail 获取用户详情
// GET /system/user/detail?id=1
func (h *Handler) Detail(c *qi.Context, req *DetailReq) (*DetailResp, error) {
	return h.svc.Detail(c.Request().Context(), req)
}

// List 获取用户列表
// GET /system/user/list?page=1&size=10
func (h *Handler) List(c *qi.Context, req *ListReq) (*listResp, error) {
	items, total, err := h.svc.List(c.Request().Context(), req)
	if err != nil {
		return nil, err
	}
	return &listResp{List: items, Total: total}, nil
}

// Options 获取用户选项列表
// GET /system/user/options
func (h *Handler) Options(c *qi.Context) (*optionsResp, error) {
	opts, err := h.svc.Options(c.Request().Context())
	if err != nil {
		return nil, err
	}
	return &optionsResp{List: opts}, nil
}

// ResetPassword 重置用户密码
// POST /system/user/reset-password  body: ResetPasswordReq
func (h *Handler) ResetPassword(c *qi.Context, req *ResetPasswordReq) error {
	return h.svc.ResetPassword(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// AssignRoles 分配用户角色
// POST /system/user/assign-roles  body: AssignRolesReq
func (h *Handler) AssignRoles(c *qi.Context, req *AssignRolesReq) error {
	return h.svc.AssignRoles(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}

// AssignPosts 分配用户岗位
// POST /system/user/assign-posts  body: AssignPostsReq
func (h *Handler) AssignPosts(c *qi.Context, req *AssignPostsReq) error {
	return h.svc.AssignPosts(c.Request().Context(), req, c.Gin().GetInt64("user_id"))
}
