package dept

import (
	"github.com/tokmz/qi"
)

// RegisterRoutes 注册部门模块路由
func RegisterRoutes(group *qi.RouterGroup, h *Handler, middleware ...qi.HandlerFunc) {
	depts := group.Group("/depts")
	api := depts.API

	// 创建部门
	api().
		POST("", qi.BindE(h.Create)).
		Summary("创建部门").
		Tags("部门管理").
		Done()

	// 批量删除部门
	api().
		POST("/delete", qi.BindE(h.Delete)).
		Summary("批量删除部门").
		Tags("部门管理").
		Done()

	// 更新部门
	api().
		POST("/update", qi.BindE(h.Update)).
		Summary("更新部门").
		Tags("部门管理").
		Done()

	// 修改部门状态
	api().
		POST("/status", qi.BindE(h.UpdateStatus)).
		Summary("修改部门状态").
		Tags("部门管理").
		Done()

	// 查询部门详情
	api().
		GET("/detail", qi.Bind(h.Detail)).
		Summary("查询部门详情").
		Tags("部门管理").
		Done()

	// 查询部门列表（树形结构）
	api().
		GET("/list", qi.Bind(h.List)).
		Summary("查询部门列表").
		Tags("部门管理").
		Done()

	// 部门选项列表
	api().
		GET("/options", qi.BindR(h.Options)).
		Summary("部门选项列表").
		Tags("部门管理").
		Done()
}
