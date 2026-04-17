package datascope

import "context"

// DataScopeKey 上下文 key
type DataScopeKey struct{}

// DataScope 数据权限上下文
type DataScope struct {
	UserID       int64   `json:"user_id"`        // 当前用户ID
	DeptID       int64   `json:"dept_id"`        // 用户所属部门ID
	Scope        int8    `json:"scope"`          // 权限范围：1=全部 2=自定义 3=本部门 4=本部门及下级 5=仅本人
	DeptIDs      []int64 `json:"dept_ids"`       // 自定义部门ID列表（DataScope=2时）
	ChildDeptIDs []int64 `json:"child_dept_ids"` // 本部门+下级部门ID列表（DataScope=4时，预计算）
}

// FromContext 从上下文中获取数据权限
func FromContext(ctx context.Context) *DataScope {
	if ds, ok := ctx.Value(DataScopeKey{}).(*DataScope); ok {
		return ds
	}
	return nil
}

// InjectContext 将数据权限注入上下文
func InjectContext(ctx context.Context, ds *DataScope) context.Context {
	return context.WithValue(ctx, DataScopeKey{}, ds)
}
