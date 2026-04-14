package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysConfig 系统配置（自引用树形结构）
//
// ParentID = nil → 分组行（一组配置的容器）
// ParentID != nil → 配置项行（挂在分组下）
//
// 分组行使用字段: ConfigKey(分组编码), Name(分组名称), Icon(图标), Description, Sort, Status
// 配置项使用字段: ConfigKey(配置键), Name(配置名称), ConfigValue, ValueType, IsEncrypted, Sort, Status
//
// ValueType: string=文本 number=数字 bool=开关 json=JSON image=图片
//
// 示例数据:
//
//	ID=1 ParentID=nil  ConfigKey="email"  Name="邮箱配置" Icon="mail"
//	ID=2 ParentID=&1   ConfigKey="host"   Name="SMTP地址"  ConfigValue="smtp.ex.com" ValueType="string"
//	ID=3 ParentID=&1   ConfigKey="port"   Name="SMTP端口"  ConfigValue="465"        ValueType="number"
//	ID=4 ParentID=&1   ConfigKey="password" Name="SMTP密码" ConfigValue="***"       ValueType="string" IsEncrypted=true
type SysConfig struct {
	ID          int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                      // 主键
	ParentID    *int64         `gorm:"column:parent_id;type:bigint;index:idx_sys_config_parent_key,priority:1"`             // 父ID，nil=分组，非nil=配置项
	ConfigKey   string         `gorm:"column:config_key;type:varchar(100);not null;index:idx_sys_config_parent_key,priority:2"` // 分组编码 or 配置键
	Name        string         `gorm:"column:name;type:varchar(64);not null;default:''"`                                    // 名称（分组名 or 配置项名）
	ConfigValue string         `gorm:"column:config_value;type:text"`                                                       // 配置值（仅配置项）
	ValueType   string         `gorm:"column:value_type;type:varchar(16);not null;default:'string'"`                        // 值类型: string/number/bool/json/image（仅配置项）
	Icon        string         `gorm:"column:icon;type:varchar(128);not null;default:''"`                                   // 图标（仅分组）
	Description string         `gorm:"column:description;type:varchar(255);not null;default:''"`                            // 说明
	IsEncrypted bool           `gorm:"column:is_encrypted;type:boolean;not null;default:false"`                             // 是否加密存储（仅配置项）
	Builtin     bool           `gorm:"column:builtin;type:boolean;not null;default:false"`                                  // 内置不可删除
	Sort        int            `gorm:"column:sort;type:int;not null;default:0"`                                             // 排序（升序）
	Status      int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_config_status"`           // 状态: 1=启用 0=禁用
	CreatedBy   int64          `gorm:"column:created_by;type:bigint;not null;default:0"`                                    // 创建人ID
	UpdatedBy   int64          `gorm:"column:updated_by;type:bigint;not null;default:0"`                                    // 更新人ID
	CreatedAt   time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                         // 创建时间
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                         // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_config_deleted_at"`                 // 软删除

	// 关联关系
	Children []*SysConfig `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
}

func (SysConfig) TableName() string { return tableName("sys_config") }
