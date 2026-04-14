package entity

import (
	"time"
)

// SysMenuApi 菜单接口关联（多对多，替代逗号分隔的 permissions 字段）
type SysMenuApi struct {
	MenuID    int64     `gorm:"column:menu_id;type:bigint;not null;primaryKey;index:idx_sys_menu_api_menu_id"` // 菜单ID
	ApiID     int64     `gorm:"column:api_id;type:bigint;not null;primaryKey;index:idx_sys_menu_api_api_id"`   // 接口ID
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                   // 创建时间
}

func (SysMenuApi) TableName() string { return tableName("sys_menu_api") }
