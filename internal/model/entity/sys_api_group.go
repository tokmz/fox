package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysApiGroup API分组 (支持树形嵌套)
// Status: 1=启用 0=禁用
type SysApiGroup struct {
	ID        int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                // 主键
	ParentID  *int64         `gorm:"column:parent_id;type:bigint"`                                                   // 父分组ID，nil表示顶级
	Level     int            `gorm:"column:level;type:int;not null;default:0"`                                      // 层级深度，顶级为0
	Tree      string         `gorm:"column:tree;type:varchar(255);not null;default:''"`                             // 物化路径, 如 0,1,2
	Name      string         `gorm:"column:name;type:varchar(64);not null;default:''"`                              // 分组名称, 如 用户管理、角色管理
	Sort      int            `gorm:"column:sort;type:int;not null;default:0"`                                       // 排序（升序）
	Status    int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_api_group_status"` // 状态: 1=启用 0=禁用
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                   // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                   // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_api_group_deleted_at"`        // 软删除
}

func (SysApiGroup) TableName() string { return tableName("sys_api_group") }
