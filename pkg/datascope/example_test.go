package datascope_test

import (
	"context"

	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/fox/pkg/datascope"
	"gorm.io/gorm"
)

// Example_simpleQuery 简单查询示例
func Example_simpleQuery() {
	var db *gorm.DB
	ctx := context.Background()

	// 假设中间件已注入数据权限上下文
	// ctx = datascope.InjectContext(ctx, &datascope.DataScope{...})

	var users []entity.SysUser

	// 应用数据权限过滤
	db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Scopes(datascope.Apply(ctx, "", "dept_id", "created_by")).
		Find(&users)
}

// Example_joinQuery JOIN 查询示例
func Example_joinQuery() {
	var db *gorm.DB
	ctx := context.Background()

	type Result struct {
		UserID   int64
		Username string
		DeptName string
	}
	var results []Result

	// JOIN 查询时指定表别名
	db.WithContext(ctx).
		Table("sys_user u").
		Select("u.id as user_id, u.username, d.name as dept_name").
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
		Scopes(datascope.Apply(ctx, "u", "dept_id", "created_by")).
		Find(&results)
}

// Example_complexQuery 复杂查询示例
func Example_complexQuery() {
	var db *gorm.DB
	ctx := context.Background()

	var users []entity.SysUser

	// 组合多个条件
	db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Where("status = ?", 1).                                       // 业务条件
		Where("username LIKE ?", "%admin%").                          // 业务条件
		Scopes(datascope.Apply(ctx, "", "dept_id", "created_by")).   // 数据权限
		Order("created_at DESC").
		Limit(10).
		Find(&users)
}

// Example_orCondition OR 条件示例
func Example_orCondition() {
	var db *gorm.DB
	ctx := context.Background()

	var users []entity.SysUser

	// 查询本部门数据 OR 本人创建的数据
	db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Scopes(datascope.ApplyOr(ctx, "", "dept_id", "created_by")).
		Find(&users)
}

// Example_subquery 子查询示例
func Example_subquery() {
	var db *gorm.DB
	ctx := context.Background()

	var users []entity.SysUser

	// 子查询也需要应用数据权限
	subQuery := db.WithContext(ctx).
		Model(&entity.SysUser{}).
		Select("dept_id").
		Where("status = ?", 1).
		Scopes(datascope.Apply(ctx, "", "dept_id", "created_by"))

	db.WithContext(ctx).
		Model(&entity.SysDept{}).
		Where("id IN (?)", subQuery).
		Find(&users)
}

// Example_clearCache 清除缓存示例
func Example_clearCache() {
	// var cache cache.Cache // 实际使用时从依赖注入获取
	ctx := context.Background()

	userID := int64(1)

	// 用户角色变更后清除缓存
	// datascope.ClearCache(ctx, cache, userID)

	// 批量清除
	userIDs := []int64{1, 2, 3}
	_ = userIDs // 避免未使用变量警告

	// datascope.ClearCacheBatch(ctx, cache, userIDs)
	_ = ctx
	_ = userID
}
