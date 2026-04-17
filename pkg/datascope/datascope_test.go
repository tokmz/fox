package datascope_test

import (
	"context"
	"testing"

	"github.com/tokmz/fox/pkg/datascope"
	"github.com/stretchr/testify/assert"
)

func TestDataScope_BuildSQL(t *testing.T) {
	tests := []struct {
		name          string
		scope         int8
		deptIDs       []int64
		childDeptIDs  []int64
		expectedSQL   string
		expectedArgs  int
	}{
		{
			name:         "全部数据",
			scope:        1,
			expectedSQL:  "1 = 1",
			expectedArgs: 0,
		},
		{
			name:         "自定义部门",
			scope:        2,
			deptIDs:      []int64{1, 2, 3},
			expectedSQL:  "dept_id IN ?",
			expectedArgs: 1,
		},
		{
			name:         "本部门",
			scope:        3,
			expectedSQL:  "dept_id = ?",
			expectedArgs: 1,
		},
		{
			name:         "本部门及下级",
			scope:        4,
			childDeptIDs: []int64{1, 2, 3, 4},
			expectedSQL:  "dept_id IN ?",
			expectedArgs: 1,
		},
		{
			name:         "仅本人",
			scope:        5,
			expectedSQL:  "created_by = ?",
			expectedArgs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &datascope.DataScope{
				UserID:       100,
				DeptID:       10,
				Scope:        tt.scope,
				DeptIDs:      tt.deptIDs,
				ChildDeptIDs: tt.childDeptIDs,
			}

			ctx := datascope.InjectContext(context.Background(), ds)

			rule := datascope.GetRule(tt.scope, "dept_id", "created_by")
			sql, args := rule.BuildSQL(ctx, "")

			assert.Contains(t, sql, tt.expectedSQL)
			assert.Len(t, args, tt.expectedArgs)
		})
	}
}

func TestDataScope_WithTableAlias(t *testing.T) {
	ds := &datascope.DataScope{
		UserID: 100,
		DeptID: 10,
		Scope:  3, // 本部门
	}

	ctx := datascope.InjectContext(context.Background(), ds)

	rule := datascope.GetRule(3, "dept_id", "created_by")
	sql, args := rule.BuildSQL(ctx, "u")

	assert.Equal(t, "u.dept_id = ?", sql)
	assert.Equal(t, []any{int64(10)}, args)
}

func TestDataScope_FromContext(t *testing.T) {
	ds := &datascope.DataScope{
		UserID: 100,
		DeptID: 10,
		Scope:  3,
	}

	ctx := datascope.InjectContext(context.Background(), ds)

	retrieved := datascope.FromContext(ctx)
	assert.NotNil(t, retrieved)
	assert.Equal(t, int64(100), retrieved.UserID)
	assert.Equal(t, int64(10), retrieved.DeptID)
	assert.Equal(t, int8(3), retrieved.Scope)
}

func TestDataScope_FromContext_Nil(t *testing.T) {
	ctx := context.Background()
	retrieved := datascope.FromContext(ctx)
	assert.Nil(t, retrieved)
}
