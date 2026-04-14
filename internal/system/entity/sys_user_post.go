package entity

import (
	"time"
)

// SysUserPost 用户岗位关联（多对多）
type SysUserPost struct {
	UserID    int64     `gorm:"column:user_id;type:bigint;not null;primaryKey;index:idx_sys_user_post_user_id"` // 用户ID
	PostID    int64     `gorm:"column:post_id;type:bigint;not null;primaryKey;index:idx_sys_user_post_post_id"` // 岗位ID
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp(3);not null;autoCreateTime"`                    // 创建时间
}

func (SysUserPost) TableName() string { return tableName("sys_user_post") }
