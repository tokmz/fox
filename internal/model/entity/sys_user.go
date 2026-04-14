package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysUser 系统管理员
// Status: 1=启用 0=禁用
type SysUser struct {
	ID        int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                             // 主键
	Username  string         `gorm:"column:username;type:varchar(64);not null;uniqueIndex:uk_sys_user_username"` // 用户名
	Password  string         `gorm:"column:password;type:varchar(128);not null" json:"-"`                        // 密码（加密存储）
	Nickname  string         `gorm:"column:nickname;type:varchar(64);not null;default:''"`                       // 昵称
	Email     string         `gorm:"column:email;type:varchar(128);not null;default:''"`                         // 邮箱
	Phone     string         `gorm:"column:phone;type:varchar(32);not null;default:''"`                          // 手机号
	Avatar    string         `gorm:"column:avatar;type:varchar(512);not null;default:''"`                        // 头像地址
	Gender    int8           `gorm:"column:gender;type:smallint;not null;default:0"`                             // 性别: 0=未知 1=男 2=女
	DeptID    int64          `gorm:"column:dept_id;type:bigint;not null;default:0;index:idx_sys_user_dept_id"`   // 所属部门ID
	Remark    string         `gorm:"column:remark;type:varchar(256);not null;default:''"`                        // 备注
	Status    int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_user_status"`   // 状态: 1=启用 0=禁用
	CreatedBy int64          `gorm:"column:created_by;type:bigint;not null;default:0"`                           // 创建人ID
	UpdatedBy int64          `gorm:"column:updated_by;type:bigint;not null;default:0"`                           // 更新人ID
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_user_deleted_at"`          // 软删除

	// 关联关系（不映射数据库列，Preload 时使用）
	Dept  *SysDept   `gorm:"foreignKey:DeptID;references:ID"  json:"dept,omitempty"`
	Roles []*SysRole `gorm:"many2many:sys_user_role;joinForeignKey:UserID;joinReferences:RoleID" json:"roles,omitempty"`
	Posts []*SysPost `gorm:"many2many:sys_user_post;joinForeignKey:UserID;joinReferences:PostID" json:"posts,omitempty"`
}

func (SysUser) TableName() string { return tableName("sys_user") }
