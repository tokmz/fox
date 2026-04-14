package entity

import (
	"time"
)

// SysRoleDept 角色部门关联（多对多，用于 DataScope=2 自定义数据权限）
type SysRoleDept struct {
	RoleID    int64     `gorm:"column:role_id;type:bigint;not null;primaryKey;index:idx_sys_role_dept_role_id"` // 角色ID
	DeptID    int64     `gorm:"column:dept_id;type:bigint;not null;primaryKey;index:idx_sys_role_dept_dept_id"` // 部门ID
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                    // 创建时间
}

func (SysRoleDept) TableName() string { return tableName("sys_role_dept") }
