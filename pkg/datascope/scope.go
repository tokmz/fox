package datascope

import (
	"context"

	"gorm.io/gorm"
)

// Apply 应用数据权限过滤（GORM Scope 函数）
// table: 表名或别名（可选，用于 JOIN 查询）
// deptColumn: 部门字段名，如 "dept_id"
// creatorColumn: 创建人字段名，如 "created_by"
//
// 使用示例：
//
//	// 简单查询
//	db.Model(&entity.SysUser{}).Scopes(datascope.Apply(ctx, "", "dept_id", "created_by")).Find(&users)
//
//	// JOIN 查询
//	db.Table("sys_user u").
//	    Joins("LEFT JOIN sys_dept d ON u.dept_id = d.id").
//	    Scopes(datascope.Apply(ctx, "u", "dept_id", "created_by")).
//	    Find(&users)
func Apply(ctx context.Context, table, deptColumn, creatorColumn string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		ds := FromContext(ctx)
		if ds == nil {
			// 无权限上下文，不过滤（开发环境或内部调用）
			return db
		}

		rule := GetRule(ds.Scope, deptColumn, creatorColumn)
		sql, args := rule.BuildSQL(ctx, table)

		return db.Where(sql, args...)
	}
}

// ApplyOr 应用数据权限过滤（OR 条件）
// 用于需要同时检查部门权限和创建人权限的场景
//
// 使用示例：
//
//	// 查询本部门数据 OR 本人创建的数据
//	db.Model(&entity.SysUser{}).Scopes(datascope.ApplyOr(ctx, "", "dept_id", "created_by")).Find(&users)
func ApplyOr(ctx context.Context, table, deptColumn, creatorColumn string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		ds := FromContext(ctx)
		if ds == nil {
			return db
		}

		// 如果是全部数据权限，直接返回
		if ds.Scope == 1 {
			return db
		}

		// 构建部门权限条件
		deptRule := GetRule(ds.Scope, deptColumn, "")
		deptSQL, deptArgs := deptRule.BuildSQL(ctx, table)

		// 构建创建人权限条件
		creatorRule := &CreatorRule{CreatorColumn: creatorColumn}
		creatorSQL, creatorArgs := creatorRule.BuildSQL(ctx, table)

		// 合并参数
		args := append(deptArgs, creatorArgs...)

		// OR 条件
		return db.Where(deptSQL+" OR "+creatorSQL, args...)
	}
}
