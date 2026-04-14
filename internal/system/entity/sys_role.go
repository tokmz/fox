package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysRole 系统角色
// Type:      1=后台角色 2=前台角色
// Status:    1=启用 0=禁用
// DataScope: 1=全部 2=自定义 3=本部门 4=本部门及下级 5=仅本人
type SysRole struct {
	ID                int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                // 主键
	ParentID          *int64         `gorm:"column:parent_id;type:bigint"`                                                   // 父角色ID，nil表示顶级
	Level             int            `gorm:"column:level;type:int;not null;default:0"`                                      // 层级深度，顶级为0
	Tree              string         `gorm:"column:tree;type:varchar(255);not null;default:''"`                             // 物化路径, 如 0,1,2
	Name              string         `gorm:"column:name;type:varchar(64);not null;default:'';uniqueIndex:uk_sys_role_name"` // 角色名称
	Code              string         `gorm:"column:code;type:varchar(64);not null;default:'';uniqueIndex:uk_sys_role_code"` // 角色编码，如 admin/editor
	DataScope         int8           `gorm:"column:data_scope;type:smallint;not null;default:5"`                            // 数据权限范围
	DeptCheckStrictly bool           `gorm:"column:dept_check_strictly;type:boolean;not null;default:true"`                 // 部门树父子节点是否联动
	Builtin           bool           `gorm:"column:builtin;type:boolean;not null;default:false"`                            // 内置角色，不可删除
	Sort              int            `gorm:"column:sort;type:int;not null;default:0"`                                       // 排序（升序）
	Status            int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_role_status"`      // 状态
	CreatedBy         int64          `gorm:"column:created_by;type:bigint;not null;default:0"`                              // 创建人ID
	UpdatedBy         int64          `gorm:"column:updated_by;type:bigint;not null;default:0"`                              // 更新人ID
	CreatedAt         time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                   // 创建时间
	UpdatedAt         time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                   // 更新时间
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_role_deleted_at"`             // 软删除

	// 关联关系
	Menus []*SysMenu `gorm:"many2many:sys_role_menu;joinForeignKey:RoleID;joinReferences:MenuID" json:"menus,omitempty"`
	Depts []*SysDept `gorm:"many2many:sys_role_dept;joinForeignKey:RoleID;joinReferences:DeptID" json:"depts,omitempty"`
}

func (SysRole) TableName() string { return tableName("sys_role") }
