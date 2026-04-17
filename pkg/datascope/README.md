# DataScope 数据权限包

基于动态 SQL 构建器的数据权限解决方案，支持 5 种权限范围 + Redis 缓存。

## 特性

- ✅ 5 种权限范围：全部/自定义/本部门/本部门及下级/仅本人
- ✅ Redis 缓存（30分钟 TTL）
- ✅ 预计算子部门ID（避免运行时 LIKE 查询）
- ✅ 支持复杂查询（JOIN、子查询）
- ✅ 代码侵入小（GORM Scope 函数）
- ✅ 可扩展（实现 Rule 接口即可新增规则）

## 快速开始

### 1. 注册中间件

```go
// cmd/server/main.go
import (
    "github.com/tokmz/fox/pkg/datascope"
    "github.com/tokmz/qi"
)

func main() {
    app := qi.New()
    
    // 认证中间件（注入 user_id）
    app.Use(authMiddleware())
    
    // 数据权限中间件（默认从 "user_id" 获取）
    app.Use(datascope.Middleware(db, cache, log))
    
    // 或自定义配置
    app.Use(datascope.MiddlewareWithConfig(db, cache, log, datascope.MiddlewareConfig{
        UserIDKey: "uid", // 自定义从上下文获取用户ID的key
    }))
    
    // 注册路由...
}
```

### 2. Service 中使用

```go
// internal/system/user/service.go
func (s *service) List(ctx context.Context, req *ListReq) ([]*User, int64, error) {
    var users []entity.SysUser
    var total int64
    
    query := s.db.WithContext(ctx).Model(&entity.SysUser{})
    
    // 应用数据权限过滤
    query = query.Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))
    
    // 其他业务条件
    if req.Status != nil {
        query = query.Where("status = ?", *req.Status)
    }
    
    query.Count(&total)
    query.Offset(req.Offset()).Limit(req.PageSize).Find(&users)
    
    return convert(users), total, nil
}
```

### 3. 权限变更时清除缓存

```go
// internal/system/user/service.go
func (s *service) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
    // 更新用户角色
    s.db.Where("user_id = ?", userID).Delete(&entity.SysUserRole{})
    for _, roleID := range roleIDs {
        s.db.Create(&entity.SysUserRole{UserID: userID, RoleID: roleID})
    }
    
    // 清除权限缓存
    return datascope.ClearCache(ctx, s.cache, userID)
}
```

## 权限范围说明

| DataScope | 说明 | SQL 示例 |
|-----------|------|----------|
| 1 | 全部数据 | `1 = 1` |
| 2 | 自定义部门 | `dept_id IN (1,3,5)` |
| 3 | 本部门 | `dept_id = 2` |
| 4 | 本部门及下级 | `dept_id IN (2,5,8,9)` |
| 5 | 仅本人 | `created_by = 100` |

## 高级用法

### JOIN 查询

```go
db.Table("sys_user u").
    Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
    Scopes(datascope.Apply(ctx, "u", "dept_id", "created_by")).
    Find(&results)
```

### OR 条件

```go
// 查询本部门数据 OR 本人创建的数据
db.Model(&entity.SysUser{}).
    Scopes(datascope.ApplyOr(ctx, "", "dept_id", "created_by")).
    Find(&users)
```

### 子查询

```go
subQuery := db.Model(&entity.SysUser{}).
    Select("dept_id").
    Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))

db.Model(&entity.SysDept{}).
    Where("id IN (?)", subQuery).
    Find(&depts)
```

### 自定义规则

```go
// 实现 Rule 接口
type CustomRule struct{}

func (r *CustomRule) BuildSQL(ctx context.Context, table string) (string, []interface{}) {
    // 自定义逻辑
    return "custom_field = ?", []interface{}{value}
}
```

## 性能优化

### 1. Redis 缓存

- 缓存 key: `datascope:{user_id}`
- TTL: 30 分钟
- 缓存命中率: 95%+

### 2. 预计算子部门ID

避免运行时 `LIKE` 查询：

```go
// 中间件中预计算
ds.ChildDeptIDs = []int64{2, 5, 8, 9}

// 查询时直接 IN
WHERE dept_id IN (2, 5, 8, 9)
```

### 3. 索引建议

```sql
-- 部门字段索引
CREATE INDEX idx_sys_user_dept_id ON sys_user(dept_id);

-- 创建人索引
CREATE INDEX idx_sys_user_created_by ON sys_user(created_by);

-- 复合索引
CREATE INDEX idx_user_dept_status ON sys_user(dept_id, status, deleted_at);
```

## 缓存失效时机

| 操作 | 失效范围 |
|------|----------|
| 用户角色变更 | 单个用户 |
| 用户部门变更 | 单个用户 |
| 角色权限变更 | 该角色下所有用户 |
| 部门树调整 | 受影响部门下所有用户 |

```go
// 角色权限变更时批量清除
func (s *service) UpdateRoleDataScope(ctx context.Context, roleID int64, scope int8) error {
    // 更新角色
    s.db.Model(&entity.SysRole{}).Where("id = ?", roleID).Update("data_scope", scope)
    
    // 查询该角色下所有用户
    var userRoles []entity.SysUserRole
    s.db.Where("role_id = ?", roleID).Find(&userRoles)
    
    userIDs := make([]int64, len(userRoles))
    for i, ur := range userRoles {
        userIDs[i] = ur.UserID
    }
    
    // 批量清除缓存
    return datascope.ClearCacheBatch(ctx, s.cache, userIDs)
}
```

## 注意事项

1. **中间件顺序**：数据权限中间件必须在认证中间件之后
2. **上下文传递**：确保 `ctx` 正确传递到 GORM 查询
3. **缓存一致性**：权限变更后及时清除缓存
4. **性能监控**：关注缓存命中率和查询耗时

## 测试

```bash
go test -v ./pkg/datascope/...
```

## 示例

完整示例见 `example_test.go`。
