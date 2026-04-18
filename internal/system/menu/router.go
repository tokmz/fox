package menu

import (
	"github.com/tokmz/qi"
)

// RegisterRoutes 注册菜单模块路由
func RegisterRoutes(group *qi.RouterGroup, h *Handler, middleware ...qi.HandlerFunc) {
	menus := group.Group("/menus")
	api := menus.API

	// 创建菜单
	api().
		POST("", qi.BindE(h.Create)).
		Summary("创建菜单").
		Tags("菜单管理").
		Done()

	// 批量删除菜单
	api().
		POST("/delete", qi.BindE(h.Delete)).
		Summary("批量删除菜单").
		Tags("菜单管理").
		Done()

	// 更新菜单
	api().
		POST("/update", qi.BindE(h.Update)).
		Summary("更新菜单").
		Tags("菜单管理").
		Done()

	// 修改菜单状态
	api().
		POST("/status", qi.BindE(h.UpdateStatus)).
		Summary("修改菜单状态").
		Tags("菜单管理").
		Done()

	// 查询菜单详情
	api().
		GET("/detail", qi.Bind(h.Detail)).
		Summary("查询菜单详情").
		Tags("菜单管理").
		Done()

	// 查询菜单列表（树形）
	api().
		GET("/list", qi.Bind(h.List)).
		Summary("查询菜单列表").
		Tags("菜单管理").
		Done()

	// 菜单选项列表（树形，用于权限分配）
	api().
		GET("/options", qi.BindR(h.Options)).
		Summary("菜单选项列表").
		Tags("菜单管理").
		Done()

	// 分配菜单API权限
	api().
		POST("/assign-apis", qi.BindE(h.AssignApis)).
		Summary("分配菜单API权限").
		Tags("菜单管理").
		Done()
}
