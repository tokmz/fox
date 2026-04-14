package entity

import (
	"time"

	"gorm.io/gorm"
)

// SysMenu 系统菜单
// MenuType: 1=目录 2=页面 3=按钮
// OpenType: 1=组件 2=内嵌iframe 3=外链
// Status:   1=启用 0=禁用
//
// naive-ui-admin 字段映射:
//
//	Title        → meta.title / label
//	Key          → key / name
//	Path         → path
//	Component    → component ('LAYOUT' | 'IFRAME' | view path)
//	Redirect     → redirect
//	Icon         → meta.icon
//	Sort         → sort
//	KeepAlive    → !meta.noKeepAlive
//	Hidden       → meta.hidden
//	Affix        → meta.affix
//	FrameSrc     → meta.frameSrc        (OpenType=2)
//	ExternalLink → meta.externalLink    (OpenType=3)
//	AlwaysShow   → meta.alwaysShow
//	ActiveMenu   → meta.activeMenu
type SysMenu struct {
	ID           int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`                                                                             // 主键
	ParentID     *int64         `gorm:"column:parent_id;type:bigint;index:idx_sys_menu_parent_status_sort,priority:1"`                                              // 父菜单ID，nil表示顶级
	Level        int            `gorm:"column:level;type:int;not null;default:0"`                                                                                   // 层级深度，顶级为0
	Tree         string         `gorm:"column:tree;type:varchar(255);not null;default:''"`                                                                          // 物化路径, 如 0,1,2
	Title        string         `gorm:"column:title;type:varchar(64);not null;default:''"`                                                                          // 菜单名称 → meta.title / label
	Key          string         `gorm:"column:route_name;type:varchar(64);not null;default:'';uniqueIndex:uk_sys_menu_route_name"`                                  // 路由 name → key / name, 唯一标识
	Path         string         `gorm:"column:path;type:varchar(256);not null;default:''"`                                                                          // 路由 path → path
	Component    string         `gorm:"column:component;type:varchar(256);not null;default:''"`                                                                     // 组件路径 → component ('LAYOUT'|'IFRAME'|view path)
	Redirect     string         `gorm:"column:redirect;type:varchar(256);not null;default:''"`                                                                      // 重定向路径 → redirect
	Query        string         `gorm:"column:query;type:varchar(256);not null;default:''"`                                                                         // 路由参数 JSON, 如 {"key":"val"}
	MenuType     int8           `gorm:"column:menu_type;type:smallint;not null;default:1;index:idx_sys_menu_type;index:idx_sys_menu_parent_status_sort,priority:4"` // 菜单类型: 1=目录 2=页面 3=按钮
	OpenType     int8           `gorm:"column:open_type;type:smallint;not null;default:1"`                                                                          // 打开方式: 1=组件 2=内嵌iframe 3=外链
	Icon         string         `gorm:"column:icon;type:varchar(128);not null;default:''"`                                                                          // 图标名称 → meta.icon
	Sort         int            `gorm:"column:sort;type:int;not null;default:0;index:idx_sys_menu_parent_status_sort,priority:3"`                                   // 排序（升序）→ sort
	KeepAlive    int8           `gorm:"column:keep_alive;type:smallint;not null;default:0"`                                                                         // 是否缓存页面 (0=否 1=是) → !meta.noKeepAlive
	Hidden       int8           `gorm:"column:hidden;type:smallint;not null;default:0"`                                                                             // 是否在菜单中隐藏 (0=否 1=是) → meta.hidden
	Affix        int8           `gorm:"column:affix;type:smallint;not null;default:0"`                                                                              // 是否固定在标签页 (0=否 1=是) → meta.affix
	AlwaysShow   int8           `gorm:"column:always_show;type:smallint;not null;default:0"`                                                                        // 是否强制显示根路由 (0=否 1=是) → meta.alwaysShow
	ActiveMenu   string         `gorm:"column:active_menu;type:varchar(64);not null;default:''"`                                                                    // 高亮指定菜单的 route_name → meta.activeMenu
	FrameSrc     string         `gorm:"column:frame_src;type:varchar(512);not null;default:''"`                                                                     // iframe 地址 (OpenType=2) → meta.frameSrc
	ExternalLink string         `gorm:"column:external_link;type:varchar(512);not null;default:''"`                                                                 // 外链地址 (OpenType=3) → meta.externalLink
	Remark       string         `gorm:"column:remark;type:varchar(255);not null;default:''"`                                                                        // 备注
	Status       int8           `gorm:"column:status;type:smallint;not null;default:1;index:idx_sys_menu_status;index:idx_sys_menu_parent_status_sort,priority:2"`  // 状态: 1=启用 0=禁用
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                                                                // 创建时间
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamp(3);not null;autoUpdateTime"`                                                                // 更新时间
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp(3);index:idx_sys_menu_deleted_at"`                                                          // 软删除

	// 关联关系（权限通过关联表获取，不再用逗号分隔字符串）
	Apis []*SysApi `gorm:"many2many:sys_menu_api;joinForeignKey:MenuID;joinReferences:ApiID" json:"apis,omitempty"`
}

func (SysMenu) TableName() string { return tableName("sys_menu") }
