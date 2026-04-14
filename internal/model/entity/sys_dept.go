package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysDept 系统部门（树形结构）
// DeptType: 1=公司 2=部门 3=小组
// Status:   1=启用 0=禁用
type SysDept struct {
	ID        int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                                  // 主键
	ParentID  *int64         `gorm:"column:parent_id;type:bigint;index:idx_sys_dept_parent_status_sort,priority:1"`                   // 父部门ID，nil表示顶级
	Level     int            `gorm:"column:level;type:int;not null;default:0"`                                                        // 层级深度，顶级为0
	Tree      string         `gorm:"column:tree;type:varchar(255);not null;default:''"`                                               // 物化路径, 如 0,1,2
	Name      string         `gorm:"column:name;type:varchar(64);not null;default:''"`                                                // 部门名称
	Code      string         `gorm:"column:code;type:varchar(64);not null;default:'';uniqueIndex:uk_sys_dept_code"`                   // 部门编码
	DeptType  int8           `gorm:"column:dept_type;type:smallint;not null;default:2"`                                               // 部门类型
	LeaderID  *int64         `gorm:"column:leader_id;type:bigint;index:idx_sys_dept_leader_id"`                                       // 部门负责人用户ID，nil表示暂未指定
	Sort      int            `gorm:"column:sort;type:int;not null;default:0;index:idx_sys_dept_parent_status_sort,priority:3"`        // 排序（升序）
	Status    int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_dept_parent_status_sort,priority:2"` // 状态
	CreatedBy int64          `gorm:"column:created_by;type:bigint;not null;default:0"`                                                // 创建人ID
	UpdatedBy int64          `gorm:"column:updated_by;type:bigint;not null;default:0"`                                                // 更新人ID
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                                     // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                                     // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_dept_deleted_at"`                               // 软删除
}

func (SysDept) TableName() string { return tableName("sys_dept") }
