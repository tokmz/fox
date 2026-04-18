package user

import (
	"github.com/tokmz/qi"
)

// RegisterRoutes 注册用户模块路由
func RegisterRoutes(group *qi.RouterGroup, h *Handler, middleware ...qi.HandlerFunc) {
	users := group.Group("/users")
	api := users.API

	// 创建用户
	api().
		POST("", qi.BindE(h.Create)).
		Summary("创建用户").
		Tags("用户管理").
		Done()

	// 批量删除用户
	api().
		POST("/delete", qi.BindE(h.Delete)).
		Summary("批量删除用户").
		Tags("用户管理").
		Done()

	// 更新用户
	api().
		POST("/update", qi.BindE(h.Update)).
		Summary("更新用户").
		Tags("用户管理").
		Done()

	// 修改用户状态
	api().
		POST("/status", qi.BindE(h.UpdateStatus)).
		Summary("修改用户状态").
		Tags("用户管理").
		Done()

	// 查询用户详情
	api().
		GET("/detail", qi.Bind(h.Detail)).
		Summary("查询用户详情").
		Tags("用户管理").
		Done()

	// 查询用户列表
	api().
		GET("/list", qi.Bind(h.List)).
		Summary("查询用户列表").
		Tags("用户管理").
		Done()

	// 用户选项列表
	api().
		GET("/options", qi.BindR(h.Options)).
		Summary("用户选项列表").
		Tags("用户管理").
		Done()

	// 重置用户密码
	api().
		POST("/reset-password", qi.BindE(h.ResetPassword)).
		Summary("重置用户密码").
		Tags("用户管理").
		Done()

	// 分配用户角色
	api().
		POST("/assign-roles", qi.BindE(h.AssignRoles)).
		Summary("分配用户角色").
		Tags("用户管理").
		Done()

	// 分配用户岗位
	api().
		POST("/assign-posts", qi.BindE(h.AssignPosts)).
		Summary("分配用户岗位").
		Tags("用户管理").
		Done()
}
