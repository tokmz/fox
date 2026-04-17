package post

import (
	"github.com/tokmz/qi"
)

// RegisterRoutes 注册岗位模块路由
func RegisterRoutes(group *qi.RouterGroup, h *Handler, middleware ...qi.HandlerFunc) {
	posts := group.Group("/posts")
	api := posts.API

	// 创建岗位
	api().
		POST("", qi.BindE(h.Create)).
		Summary("创建岗位").
		Tags("岗位管理").
		Done()

	// 批量删除岗位
	api().
		POST("/delete", qi.BindE(h.Delete)).
		Summary("批量删除岗位").
		Tags("岗位管理").
		Done()

	// 更新岗位
	api().
		POST("/update", qi.BindE(h.Update)).
		Summary("更新岗位").
		Tags("岗位管理").
		Done()

	// 修改岗位状态
	api().
		POST("/status", qi.BindE(h.UpdateStatus)).
		Summary("修改岗位状态").
		Tags("岗位管理").
		Done()

	// 查询岗位详情
	api().
		GET("/detail", qi.Bind(h.Detail)).
		Summary("查询岗位详情").
		Tags("岗位管理").
		Done()

	// 岗位选项列表
	api().
		GET("/options", qi.Bind(h.Options)).
		Summary("岗位选项列表").
		Tags("岗位管理").
		Done()
}
