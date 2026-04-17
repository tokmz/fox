package datascope

import (
	"context"
	"fmt"
)

// Rule 数据权限规则接口
type Rule interface {
	// BuildSQL 构建 SQL 条件片段
	// table: 表名或别名
	// 返回: (sql条件, 参数列表)
	BuildSQL(ctx context.Context, table string) (string, []any)
}

// buildColumn 构建安全的列名（防止 SQL 注入）
func buildColumn(table, column string) (string, bool) {
	// 验证字段名
	if !IsValidColumn(column) {
		return "", false
	}

	// 验证表名/别名
	if table != "" && !IsValidIdentifier(table) {
		return "", false
	}

	if table != "" {
		return fmt.Sprintf("%s.%s", table, column), true
	}
	return column, true
}

// AllDataRule 全部数据规则（不限制）
type AllDataRule struct{}

func (r *AllDataRule) BuildSQL(ctx context.Context, table string) (string, []any) {
	return "1 = 1", nil
}

// CustomDeptRule 自定义部门规则（通过 sys_role_dept 指定）
type CustomDeptRule struct {
	DeptColumn string // 部门字段名，如 "dept_id"
}

func (r *CustomDeptRule) BuildSQL(ctx context.Context, table string) (string, []any) {
	ds := FromContext(ctx)
	if ds == nil || len(ds.DeptIDs) == 0 {
		return "1 = 0", nil // 无授权部门，返回空结果
	}

	// 构建安全的列名
	column, ok := buildColumn(table, r.DeptColumn)
	if !ok {
		return "1 = 0", nil // 非法字段名，返回空结果
	}

	return fmt.Sprintf("%s IN ?", column), []any{ds.DeptIDs}
}

// DeptRule 本部门规则
type DeptRule struct {
	DeptColumn string
}

func (r *DeptRule) BuildSQL(ctx context.Context, table string) (string, []any) {
	ds := FromContext(ctx)
	if ds == nil {
		return "1 = 0", nil
	}

	// 构建安全的列名
	column, ok := buildColumn(table, r.DeptColumn)
	if !ok {
		return "1 = 0", nil // 非法字段名，返回空结果
	}

	return fmt.Sprintf("%s = ?", column), []any{ds.DeptID}
}

// DeptTreeRule 本部门及下级规则（使用预计算的子部门ID列表）
type DeptTreeRule struct {
	DeptColumn string
}

func (r *DeptTreeRule) BuildSQL(ctx context.Context, table string) (string, []any) {
	ds := FromContext(ctx)
	if ds == nil || len(ds.ChildDeptIDs) == 0 {
		return "1 = 0", nil
	}

	// 构建安全的列名
	column, ok := buildColumn(table, r.DeptColumn)
	if !ok {
		return "1 = 0", nil // 非法字段名，返回空结果
	}

	return fmt.Sprintf("%s IN ?", column), []any{ds.ChildDeptIDs}
}

// CreatorRule 仅本人规则（创建人）
type CreatorRule struct {
	CreatorColumn string // 创建人字段名，如 "created_by"
}

func (r *CreatorRule) BuildSQL(ctx context.Context, table string) (string, []any) {
	ds := FromContext(ctx)
	if ds == nil {
		return "1 = 0", nil
	}

	// 构建安全的列名
	column, ok := buildColumn(table, r.CreatorColumn)
	if !ok {
		return "1 = 0", nil // 非法字段名，返回空结果
	}

	return fmt.Sprintf("%s = ?", column), []any{ds.UserID}
}

// GetRule 根据 DataScope 获取对应规则
func GetRule(scope int8, deptCol, creatorCol string) Rule {
	switch scope {
	case 1:
		return &AllDataRule{}
	case 2:
		return &CustomDeptRule{DeptColumn: deptCol}
	case 3:
		return &DeptRule{DeptColumn: deptCol}
	case 4:
		return &DeptTreeRule{DeptColumn: deptCol}
	case 5:
		return &CreatorRule{CreatorColumn: creatorCol}
	default:
		return &CreatorRule{CreatorColumn: creatorCol} // 默认最严格
	}
}
