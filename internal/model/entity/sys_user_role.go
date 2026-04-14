package entity

import (
	"time"
)

// SysUserRole 用户角色关联（多对多）
type SysUserRole struct {
	UserID    int64     `gorm:"column:user_id;type:bigint;not null;primaryKey;index:idx_sys_user_role_user_id"` // 用户ID
	RoleID    int64     `gorm:"column:role_id;type:bigint;not null;primaryKey;index:idx_sys_user_role_role_id"` // 角色ID
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                    // 创建时间
}

func (SysUserRole) TableName() string { return tableName("sys_user_role") }
