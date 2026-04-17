package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysPost 系统岗位
// Status: 1=启用 0=禁用
type SysPost struct {
	ID        int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                           // 主键
	DeptID    int64          `gorm:"column:dept_id;type:bigint;not null;index:idx_sys_post_dept_status,priority:1"`            // 所属部门ID
	Name      string         `gorm:"column:name;type:varchar(64);not null;default:''"`                                         // 岗位名称
	Code      string         `gorm:"column:code;type:varchar(64);not null;default:'';uniqueIndex:uk_sys_post_code"`            // 岗位编码
	Sort      int            `gorm:"column:sort;type:int;not null;default:0"`                                                  // 排序（升序）
	Remark    string         `gorm:"column:remark;type:varchar(256);not null;default:''"`                                      // 备注
	Status    int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_post_dept_status,priority:2"` // 状态: 1=启用 0=禁用
	CreatedBy int64          `gorm:"column:created_by;type:bigint;not null;default:0"`                                         // 创建人ID
	UpdatedBy int64          `gorm:"column:updated_by;type:bigint;not null;default:0"`                                         // 更新人ID
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                              // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                              // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_post_deleted_at"`                        // 软删除
}

func (SysPost) TableName() string { return tableName("sys_post") }
