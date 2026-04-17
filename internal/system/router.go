package system

import (
	"github.com/tokmz/fox/internal/system/dept"
	"github.com/tokmz/fox/internal/system/role"
	"github.com/tokmz/qi"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"gorm.io/gorm"
)

// RegisterRouter 注册 system 域所有模块路由
// db: 数据库实例
// log: 日志实例
// c: 缓存实例
func RegisterRouter(group *qi.RouterGroup, db *gorm.DB, log logger.Logger, c cache.Cache) {
	// 角色模块
	roleSvc := role.NewService(log, c, db)
	roleHandler := role.NewHandler(roleSvc)
	role.RegisterRoutes(group, roleHandler)

	// 部门模块
	deptSvc := dept.NewService(log, c, db)
	deptHandler := dept.NewHandler(deptSvc)
	dept.RegisterRoutes(group, deptHandler)

	// TODO: 菜单模块
	// TODO: 岗位模块
	// TODO: API 模块
	// TODO: 配置模块
	// TODO: 用户模块
	// TODO: 认证模块
}
