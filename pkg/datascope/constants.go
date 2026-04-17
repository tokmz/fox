package datascope

import "time"

// 缓存配置
const (
	DefaultCacheTTL   = 30 * time.Minute // 默认缓存过期时间
	CacheKeyPrefix    = "datascope:"     // 缓存 key 前缀
	DefaultUserIDKey  = "user_id"        // 默认用户ID key
	QueryTimeout      = 3 * time.Second  // 数据库查询超时时间
)

// 权限范围
const (
	ScopeAll        int8 = 1 // 全部数据
	ScopeCustom     int8 = 2 // 自定义部门
	ScopeDept       int8 = 3 // 本部门
	ScopeDeptTree   int8 = 4 // 本部门及下级
	ScopeOwner      int8 = 5 // 仅本人
	DefaultScope    int8 = 5 // 默认权限范围（最严格）
)

// 字段名白名单（防止 SQL 注入）
var ValidColumns = map[string]bool{
	"dept_id":    true,
	"created_by": true,
	"updated_by": true,
	"user_id":    true,
	"creator_id": true,
	"owner_id":   true,
}
