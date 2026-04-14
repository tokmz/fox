package entity

import (
	"strings"
	"sync"

	"gorm.io/gorm"
)

var (
	tablePrefix string
	tableOnce   sync.Once
)

// setTablePrefix 设置全局表前缀，仅允许设置一次。
// 重复调用会 panic，防止运行时意外覆盖。
func setTablePrefix(prefix string) {
	tableOnce.Do(func() {
		if prefix != "" && !strings.HasSuffix(prefix, "_") {
			prefix += "_"
		}
		tablePrefix = prefix
	})
}

// tableName 返回带前缀的表名。
func tableName(name string) string {
	return tablePrefix + name
}

// AutoMigrate 自动迁移所有实体表
// GORM many2many 声明在 SysUser/SysRole 上，关联表由框架自动管理外键
func AutoMigrate(db *gorm.DB, prefix ...string) error {
	if len(prefix) != 0 {
		setTablePrefix(prefix[0])
	}

	// 先迁移主表（被引用表必须先存在）
	if err := db.AutoMigrate(
		&SysUser{},
		&SysRole{},
		&SysMenu{},
		&SysDept{},
		&SysPost{},
		&SysApi{},
		&SysApiGroup{},
		&SysConfig{},
	); err != nil {
		return err
	}

	// 再迁移关联表
	return db.AutoMigrate(
		&SysUserRole{},
		&SysUserPost{},
		&SysRoleMenu{},
		&SysRoleDept{},
		&SysMenuApi{},
	)
}
