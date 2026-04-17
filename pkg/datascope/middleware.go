package datascope

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/qi"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	UserIDKey string // 从上下文中获取用户ID的key，默认 "user_id"
}

// Middleware 数据权限中间件（使用默认配置）
// 解析用户数据权限并注入上下文，支持 Redis 缓存
func Middleware(db *gorm.DB, cache cache.Cache, log logger.Logger) qi.HandlerFunc {
	return MiddlewareWithConfig(db, cache, log, MiddlewareConfig{
		UserIDKey: DefaultUserIDKey,
	})
}

// MiddlewareWithConfig 数据权限中间件（自定义配置）
func MiddlewareWithConfig(db *gorm.DB, cache cache.Cache, log logger.Logger, config MiddlewareConfig) qi.HandlerFunc {
	// 设置默认值
	if config.UserIDKey == "" {
		config.UserIDKey = "user_id"
	}

	return func(c *qi.Context) {
		// 1. 从上下文中获取用户ID
		userID, exists := c.Get(config.UserIDKey)
		if !exists {
			// 未登录，跳过数据权限
			c.Next()
			return
		}

		// 安全的类型断言
		uid, ok := userID.(int64)
		if !ok {
			log.Warn("invalid user_id type",
				zap.String("key", config.UserIDKey),
				zap.Any("value", userID))
			c.Next()
			return
		}

		// 2. 尝试从缓存读取
		cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, uid)
		var ds DataScope

		var cachedData string
		err := cache.Get(c.Context(), cacheKey, &cachedData)
		if err == nil && cachedData != "" {
			// 缓存命中，尝试反序列化
			if err = json.Unmarshal([]byte(cachedData), &ds); err == nil {
				log.Debug("datascope cache hit", zap.Int64("user_id", uid))
				c.WithValue(DataScopeKey{}, &ds)
				c.Next()
				return
			}
			// 反序列化失败，记录日志并继续查询数据库
			log.Warn("datascope cache unmarshal failed",
				zap.Int64("user_id", uid),
				zap.Error(err))
		}

		// 3. 缓存未命中，查询数据库并计算权限
		log.Debug("datascope cache miss, building from db", zap.Int64("user_id", uid))
		ds = buildDataScope(c.Context(), db, log, uid)

		// 4. 写入缓存（TTL 30分钟）
		if data, err := json.Marshal(ds); err == nil {
			if err := cache.Set(c.Context(), cacheKey, string(data), DefaultCacheTTL); err != nil {
				log.Warn("datascope cache set failed",
					zap.Int64("user_id", uid),
					zap.Error(err))
			}
		} else {
			log.Warn("datascope marshal failed",
				zap.Int64("user_id", uid),
				zap.Error(err))
		}

		// 5. 注入上下文
		c.WithValue(DataScopeKey{}, &ds)
		c.Next()
	}
}

// buildDataScope 构建数据权限上下文（从数据库查询）
func buildDataScope(ctx context.Context, db *gorm.DB, log logger.Logger, userID int64) DataScope {
	ds := DataScope{
		UserID: userID,
		Scope:  DefaultScope, // 默认最严格（仅本人）
	}

	// 添加查询超时控制
	queryCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	// 1. 查询用户信息 + 角色 + 部门
	var user entity.SysUser
	err := db.WithContext(queryCtx).
		Preload("Roles", "status = ?", 1). // 仅加载启用的角色
		Preload("Dept").
		First(&user, userID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("user not found", zap.Int64("user_id", userID))
		} else {
			log.Error("query user failed",
				zap.Int64("user_id", userID),
				zap.Error(err))
		}
		return ds // 返回默认权限
	}

	ds.DeptID = user.DeptID

	// 2. 计算最宽松的权限范围（多角色取最小值）
	// 收集需要查询自定义部门的角色ID
	roleIDs := make([]int64, 0, len(user.Roles))

	for _, role := range user.Roles {
		if role.DataScope < ds.Scope {
			ds.Scope = role.DataScope
		}

		// 收集 DataScope=2 的角色ID
		if role.DataScope == ScopeCustom {
			roleIDs = append(roleIDs, role.ID)
		}
	}

	// 3. 批量查询自定义部门ID（优化 N+1 查询）
	if len(roleIDs) > 0 {
		var roleDepts []entity.SysRoleDept
		if err := db.WithContext(queryCtx).
			Where("role_id IN ?", roleIDs).
			Find(&roleDepts).Error; err != nil {
			log.Error("query role depts failed",
				zap.Int64s("role_ids", roleIDs),
				zap.Error(err))
		} else {
			for _, rd := range roleDepts {
				ds.DeptIDs = append(ds.DeptIDs, rd.DeptID)
			}
			// 去重
			if len(ds.DeptIDs) > 0 {
				ds.DeptIDs = uniqueInt64(ds.DeptIDs)
			}
		}
	}

	// 4. 预计算子部门ID列表（DataScope=4）
	if ds.Scope == ScopeDeptTree && user.Dept != nil {
		var childDepts []entity.SysDept
		if err := db.WithContext(queryCtx).
			Select("id").
			Where("id = ? OR tree LIKE ?", user.DeptID, user.Dept.Tree+",%").
			Find(&childDepts).Error; err != nil {
			log.Error("query child depts failed",
				zap.Int64("dept_id", user.DeptID),
				zap.Error(err))
		} else {
			ds.ChildDeptIDs = make([]int64, len(childDepts))
			for i, d := range childDepts {
				ds.ChildDeptIDs[i] = d.ID
			}
		}
	}

	return ds
}

// ClearCache 清除用户数据权限缓存
// 在用户角色变更、部门变更时调用
func ClearCache(ctx context.Context, cache cache.Cache, userID int64) error {
	cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, userID)
	return cache.Del(ctx, cacheKey)
}

// ClearCacheBatch 批量清除用户数据权限缓存
func ClearCacheBatch(ctx context.Context, cache cache.Cache, userIDs []int64) error {
	keys := make([]string, len(userIDs))
	for i, uid := range userIDs {
		keys[i] = fmt.Sprintf("%s%d", CacheKeyPrefix, uid)
	}
	return cache.Del(ctx, keys...)
}

// uniqueInt64 int64 切片去重
func uniqueInt64(slice []int64) []int64 {
	seen := make(map[int64]struct{}, len(slice))
	result := make([]int64, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
