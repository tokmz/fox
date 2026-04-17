package role

import (
	"github.com/tokmz/qi"
)

// RegisterRoutes 注册角色模块路由
func RegisterRoutes(group *qi.RouterGroup, h *Handler, middleware ...qi.HandlerFunc) {
	roles := group.Group("/roles")
	api := roles.API

	// 创建角色
	api().
		POST("", qi.BindE(h.Create)).
		Summary("创建角色").
		Tags("角色管理").
		Done()

	// 批量删除角色
	api().
		POST("/delete", qi.BindE(h.Delete)).
		Summary("批量删除角色").
		Tags("角色管理").
		Done()

	// 更新角色
	api().
		POST("/update", qi.BindE(h.Update)).
		Summary("更新角色").
		Tags("角色管理").
		Done()

	// 修改角色状态
	api().
		POST("/status", qi.BindE(h.UpdateStatus)).
		Summary("修改角色状态").
		Tags("角色管理").
		Done()

	// 查询角色详情
	api().
		GET("/detail", qi.Bind(h.Detail)).
		Summary("查询角色详情").
		Tags("角色管理").
		Done()

	// 查询角色列表（树形结构）
	api().
		GET("/list", qi.Bind(h.List)).
		Summary("查询角色列表").
		Tags("角色管理").
		Done()

	// 角色选项列表
	api().
		GET("/options", qi.BindR(h.Options)).
		Summary("角色选项列表").
		Tags("角色管理").
		Done()

	// 分配角色菜单权限
	api().
		POST("/menus", qi.BindE(h.AssignMenus)).
		Summary("分配角色菜单权限").
		Tags("角色管理").
		Done()

	// 分配角色自定义部门
	api().
		POST("/depts", qi.BindE(h.AssignDepts)).
		Summary("分配角色自定义部门").
		Tags("角色管理").
		Done()
}
