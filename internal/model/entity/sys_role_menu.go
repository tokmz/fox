package entity

import (
	"time"
)

// SysRoleMenu 角色菜单关联（多对多）
type SysRoleMenu struct {
	RoleID    int64     `gorm:"column:role_id;type:bigint;not null;primaryKey;index:idx_sys_role_menu_role_id"` // 角色ID
	MenuID    int64     `gorm:"column:menu_id;type:bigint;not null;primaryKey;index:idx_sys_role_menu_menu_id"` // 菜单ID
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                    // 创建时间
}

func (SysRoleMenu) TableName() string { return tableName("sys_role_menu") }
