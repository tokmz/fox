package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysApi 系统接口
// Method: HTTP 请求方法 (GET/POST/PUT/DELETE)
// Status: 1=启用 0=禁用
type SysApi struct {
	ID          int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                                   // 主键
	GroupID     int64          `gorm:"column:group_id;type:bigint;not null;default:0;index:idx_sys_api_group_id"`                        // 所属分组ID
	Method      string         `gorm:"column:method;type:varchar(16);not null;default:'';uniqueIndex:uk_sys_api_method_path,priority:1"` // 请求方法: GET/POST/PUT/DELETE
	Path        string         `gorm:"column:path;type:varchar(255);not null;default:'';uniqueIndex:uk_sys_api_method_path,priority:2"`  // 请求路径, 如 /api/system/user/list
	Name        string         `gorm:"column:name;type:varchar(64);not null;default:''"`                                                 // 接口名称, 如 获取用户列表
	Permission  string         `gorm:"column:permission;type:varchar(128);not null;default:'';index:idx_sys_api_permission"`             // 权限标识, 关联菜单 Permissions, 如 sys:user:list
	Description string         `gorm:"column:description;type:varchar(255);not null;default:''"`                                         // 接口描述
	Sort        int            `gorm:"column:sort;type:int;not null;default:0"`                                                          // 排序（升序）
	Status      int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_api_status"`                          // 状态: 1=启用 0=禁用
	CreatedBy   int64          `gorm:"column:created_by;type:bigint;not null;default:0"`                                                 // 创建人ID
	UpdatedBy   int64          `gorm:"column:updated_by;type:bigint;not null;default:0"`                                                 // 更新人ID
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                                      // 创建时间
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                                      // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_api_deleted_at"`                                 // 软删除
}

func (SysApi) TableName() string { return tableName("sys_api") }
