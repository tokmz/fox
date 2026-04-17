# DataScope 完整使用示例

## 1. 项目集成

### 1.1 主程序注册中间件

```go
// cmd/server/main.go
package main

import (
	"github.com/tokmz/fox/pkg/datascope"
	"github.com/tokmz/qi"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"gorm.io/gorm"
)

func main() {
	// 初始化依赖
	db := initDB()
	cache := initCache()
	log := initLogger()

	app := qi.New()

	// 1. 认证中间件（必须在数据权限之前，注入 user_id）
	app.Use(authMiddleware())

	// 2. 数据权限中间件（默认从 "user_id" 获取）
	app.Use(datascope.Middleware(db, cache, log))

	// 或自定义配置
	// app.Use(datascope.MiddlewareWithConfig(db, cache, log, datascope.MiddlewareConfig{
	//     UserIDKey: "uid", // 自定义key
	// }))

	// 3. 注册业务路由
	registerRoutes(app, db, cache, log)

	app.Run(":8080")
}

// authMiddleware 认证中间件示例（从 JWT 解析 user_id）
func authMiddleware() qi.HandlerFunc {
	return func(c *qi.Context) {
		// 从 JWT token 中解析用户ID
		token := c.GetHeader("Authorization")
		userID := parseJWT(token) // 你的 JWT 解析逻辑

		// 注入到上下文
		c.Set("user_id", userID)
		c.Next()
	}
}
```

### 1.2 Service 层使用

```go
// internal/system/user/service.go
package user

import (
	"context"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/fox/pkg/datascope"
	"github.com/tokmz/qi/pkg/cache"
	"github.com/tokmz/qi/pkg/logger"
	"gorm.io/gorm"
)

type Service interface {
	List(ctx context.Context, req *ListReq) ([]*ListItem, int64, error)
	Detail(ctx context.Context, id int64) (*DetailResp, error)
	Create(ctx context.Context, req *CreateReq) error
	Update(ctx context.Context, req *UpdateReq) error
	Delete(ctx context.Context, ids []int64) error
	AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error
}

type service struct {
	db    *gorm.DB
	cache cache.Cache
	log   logger.Logger
}

func NewService(log logger.Logger, cache cache.Cache, db *gorm.DB) Service {
	return &service{db: db, cache: cache, log: log}
}

// List 查询用户列表（带数据权限）
func (s *service) List(ctx context.Context, req *ListReq) ([]*ListItem, int64, error) {
	var users []entity.SysUser
	var total int64

	query := s.db.WithContext(ctx).Model(&entity.SysUser{})

	// ✅ 应用数据权限过滤
	query = query.Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))

	// 业务查询条件
	if req.Username != nil {
		query = query.Where("username LIKE ?", "%"+*req.Username+"%")
	}
	if req.DeptID != nil {
		query = query.Where("dept_id = ?", *req.DeptID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 分页
	query.Count(&total)
	query.Offset(req.Offset()).Limit(req.PageSize).Find(&users)

	return convertToListItems(users), total, nil
}

// Detail 查询用户详情（带数据权限）
func (s *service) Detail(ctx context.Context, id int64) (*DetailResp, error) {
	var user entity.SysUser

	err := s.db.WithContext(ctx).
		Scopes(datascope.Apply(ctx, "", "dept_id", "created_by")).
		Preload("Dept").
		Preload("Roles").
		Preload("Posts").
		First(&user, id).Error

	if err != nil {
		return nil, err
	}

	return convertToDetail(&user), nil
}

// AssignRoles 分配角色（需要清除权限缓存）
func (s *service) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	// 1. 删除旧角色
	s.db.Where("user_id = ?", userID).Delete(&entity.SysUserRole{})

	// 2. 添加新角色
	for _, roleID := range roleIDs {
		s.db.Create(&entity.SysUserRole{
			UserID: userID,
			RoleID: roleID,
		})
	}

	// 3. ✅ 清除权限缓存
	return datascope.ClearCache(ctx, s.cache, userID)
}
```

## 2. 复杂查询场景

### 2.1 JOIN 查询

```go
// 查询用户及部门信息
func (s *service) ListWithDept(ctx context.Context, req *ListReq) ([]*UserWithDept, error) {
	type Result struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
		DeptID   int64  `json:"dept_id"`
		DeptName string `json:"dept_name"`
	}
	var results []Result

	err := s.db.WithContext(ctx).
		Table("sys_user u").
		Select("u.id as user_id, u.username, u.dept_id, d.name as dept_name").
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
		Scopes(datascope.Apply(ctx, "u", "dept_id", "created_by")). // 指定表别名
		Where("u.status = ?", 1).
		Find(&results).Error

	return results, err
}
```

### 2.2 子查询

```go
// 查询有用户的部门列表
func (s *service) ListDeptsWithUsers(ctx context.Context) ([]*entity.SysDept, error) {
	var depts []entity.SysDept

	// 子查询：获取当前用户可见的用户的部门ID
	subQuery := s.db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Select("DISTINCT dept_id").
		Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))

	err := s.db.WithContext(ctx).
		Model(&entity.SysDept{}).
		Where("id IN (?)", subQuery).
		Find(&depts).Error

	return depts, err
}
```

### 2.3 聚合查询

```go
// 统计各部门用户数（带数据权限）
func (s *service) CountByDept(ctx context.Context) (map[int64]int64, error) {
	type Result struct {
		DeptID int64 `json:"dept_id"`
		Count  int64 `json:"count"`
	}
	var results []Result

	err := s.db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Select("dept_id, COUNT(*) as count").
		Scopes(datascope.Apply(ctx, "", "dept_id", "created_by")).
		Group("dept_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	countMap := make(map[int64]int64)
	for _, r := range results {
		countMap[r.DeptID] = r.Count
	}

	return countMap, nil
}
```

### 2.4 OR 条件查询

```go
// 查询本部门数据 OR 本人创建的数据
func (s *service) ListDeptOrOwn(ctx context.Context) ([]*entity.SysUser, error) {
	var users []entity.SysUser

	err := s.db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Scopes(datascope.ApplyOr(ctx, "", "dept_id", "created_by")).
		Find(&users).Error

	return users, err
}
```

## 3. 权限变更场景

### 3.1 用户角色变更

```go
// internal/system/user/service.go
func (s *service) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	// 更新角色
	s.db.Where("user_id = ?", userID).Delete(&entity.SysUserRole{})
	for _, roleID := range roleIDs {
		s.db.Create(&entity.SysUserRole{UserID: userID, RoleID: roleID})
	}

	// ✅ 清除单个用户缓存
	return datascope.ClearCache(ctx, s.cache, userID)
}
```

### 3.2 用户部门变更

```go
func (s *service) ChangeDept(ctx context.Context, userID, deptID int64) error {
	// 更新部门
	s.db.Model(&entity.SysUser{}).Where("id = ?", userID).Update("dept_id", deptID)

	// ✅ 清除缓存
	return datascope.ClearCache(ctx, s.cache, userID)
}
```

### 3.3 角色权限变更（批量清除）

```go
// internal/system/role/service.go
func (s *service) UpdateDataScope(ctx context.Context, roleID int64, scope int8, deptIDs []int64) error {
	// 1. 更新角色权限范围
	s.db.Model(&entity.SysRole{}).Where("id = ?", roleID).Update("data_scope", scope)

	// 2. 更新自定义部门
	if scope == 2 {
		s.db.Where("role_id = ?", roleID).Delete(&entity.SysRoleDept{})
		for _, deptID := range deptIDs {
			s.db.Create(&entity.SysRoleDept{RoleID: roleID, DeptID: deptID})
		}
	}

	// 3. ✅ 查询该角色下所有用户
	var userRoles []entity.SysUserRole
	s.db.Where("role_id = ?", roleID).Find(&userRoles)

	userIDs := make([]int64, len(userRoles))
	for i, ur := range userRoles {
		userIDs[i] = ur.UserID
	}

	// 4. ✅ 批量清除缓存
	return datascope.ClearCacheBatch(ctx, s.cache, userIDs)
}
```

### 3.4 部门树调整（批量清除）

```go
// internal/system/dept/service.go
func (s *service) MoveDept(ctx context.Context, deptID, newParentID int64) error {
	// 1. 更新部门父节点
	s.db.Model(&entity.SysDept{}).Where("id = ?", deptID).Update("parent_id", newParentID)

	// 2. 重新计算 tree 字段（略）

	// 3. ✅ 查询受影响部门下的所有用户
	var users []entity.SysUser
	s.db.Select("id").Where("dept_id = ?", deptID).Find(&users)

	userIDs := make([]int64, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// 4. ✅ 批量清除缓存
	return datascope.ClearCacheBatch(ctx, s.cache, userIDs)
}
```

## 4. 测试

### 4.1 单元测试

```go
// internal/system/user/service_test.go
package user

import (
	"context"
	"testing"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/fox/pkg/datascope"
	"github.com/stretchr/testify/assert"
)

func TestList_DataScope(t *testing.T) {
	db := setupTestDB(t)
	cache := setupTestCache(t)
	log := setupTestLogger(t)

	svc := NewService(log, cache, db)

	// 准备测试数据
	prepareTestData(db)

	tests := []struct {
		name     string
		userID   int64
		deptID   int64
		scope    int8
		expected int
	}{
		{
			name:     "全部数据",
			userID:   1,
			deptID:   1,
			scope:    1,
			expected: 10, // 所有用户
		},
		{
			name:     "本部门",
			userID:   2,
			deptID:   2,
			scope:    3,
			expected: 3, // 部门2的用户
		},
		{
			name:     "仅本人",
			userID:   3,
			deptID:   3,
			scope:    5,
			expected: 2, // 用户3创建的数据
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构造数据权限上下文
			ctx := datascope.InjectContext(context.Background(), &datascope.DataScope{
				UserID: tt.userID,
				DeptID: tt.deptID,
				Scope:  tt.scope,
			})

			// 查询
			users, total, err := svc.List(ctx, &ListReq{PageSize: 100})

			assert.NoError(t, err)
			assert.Equal(t, int64(tt.expected), total)
			assert.Len(t, users, tt.expected)
		})
	}
}
```

## 5. 性能监控

### 5.1 缓存命中率监控

```go
// pkg/datascope/middleware.go
func Middleware(db *gorm.DB, cache cache.Cache, log logger.Logger) qi.HandlerFunc {
	var cacheHit, cacheMiss int64

	return func(c *qi.Context) {
		// ... 原有逻辑

		if err == nil && cachedData != "" {
			atomic.AddInt64(&cacheHit, 1)
			// 每 1000 次打印一次命中率
			if (cacheHit+cacheMiss)%1000 == 0 {
				hitRate := float64(cacheHit) / float64(cacheHit+cacheMiss) * 100
				log.Info("datascope cache hit rate", zap.Float64("rate", hitRate))
			}
		} else {
			atomic.AddInt64(&cacheMiss, 1)
		}
	}
}
```

### 5.2 慢查询监控

```go
// 在 GORM 中添加回调
db.Callback().Query().Before("gorm:query").Register("datascope:before", func(db *gorm.DB) {
	db.InstanceSet("datascope:start_time", time.Now())
})

db.Callback().Query().After("gorm:query").Register("datascope:after", func(db *gorm.DB) {
	if startTime, ok := db.InstanceGet("datascope:start_time"); ok {
		elapsed := time.Since(startTime.(time.Time))
		if elapsed > 100*time.Millisecond {
			log.Warn("slow query with datascope", 
				zap.Duration("elapsed", elapsed),
				zap.String("sql", db.Statement.SQL.String()))
		}
	}
})
```

## 6. 常见问题

### Q1: 为什么查询结果为空？

检查是否正确注入了数据权限上下文：

```go
ds := datascope.FromContext(ctx)
if ds == nil {
	log.Warn("datascope context not found")
}
```

### Q2: JOIN 查询时权限不生效？

确保指定了正确的表别名：

```go
// ❌ 错误
Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))

// ✅ 正确
Scopes(datascope.Apply(ctx, "u", "dept_id", "created_by"))
```

### Q3: 缓存不失效？

检查是否在权限变更后调用了清除缓存：

```go
datascope.ClearCache(ctx, cache, userID)
```

### Q4: 性能问题？

1. 检查缓存命中率（应 > 90%）
2. 添加索引：`dept_id`, `created_by`
3. 检查是否有 N+1 查询

## 7. 最佳实践

1. ✅ 所有查询用户数据的接口都应用数据权限
2. ✅ 权限变更后立即清除缓存
3. ✅ 使用 `WithContext(ctx)` 传递上下文
4. ✅ JOIN 查询时指定表别名
5. ✅ 监控缓存命中率和慢查询
6. ❌ 不要在事务外清除缓存（可能导致不一致）
7. ❌ 不要跳过数据权限检查（安全风险）
