package errcode

import (
	"net/http"

	qierr "github.com/tokmz/qi/pkg/errors"
)

// system 域业务错误码定义
// 错误码段规划（每模块占一个百位段）：
//   用户模块 100xx
//   角色模块 101xx
//   部门模块 102xx
//   岗位模块 103xx
//   菜单模块 104xx
//   API 模块 105xx
//   配置模块 106xx
//   认证模块 107xx

// ===== 用户模块 100xx =====

var (
	ErrUserNotFound   = qierr.NewWithStatus(10001, http.StatusNotFound, "用户不存在")
	ErrUserExists     = qierr.NewWithStatus(10002, http.StatusConflict, "用户已存在")
	ErrUserCreate     = qierr.NewWithStatus(10003, http.StatusInternalServerError, "创建用户失败")
	ErrUserUpdate     = qierr.NewWithStatus(10004, http.StatusInternalServerError, "更新用户失败")
	ErrUserDelete     = qierr.NewWithStatus(10005, http.StatusInternalServerError, "删除用户失败")
	ErrUserQuery      = qierr.NewWithStatus(10006, http.StatusInternalServerError, "查询用户失败")
	ErrUserRoleQuery  = qierr.NewWithStatus(10007, http.StatusInternalServerError, "查询用户角色失败")
	ErrUserPostQuery  = qierr.NewWithStatus(10008, http.StatusInternalServerError, "查询用户岗位失败")
)

// ===== 角色模块 101xx =====

var (
	ErrRoleNotFound    = qierr.NewWithStatus(10101, http.StatusNotFound, "角色不存在")
	ErrRoleExists      = qierr.NewWithStatus(10102, http.StatusConflict, "角色已存在")
	ErrRoleHasChildren = qierr.NewWithStatus(10103, http.StatusForbidden, "存在子角色，无法删除")
	ErrRoleHasUsers    = qierr.NewWithStatus(10104, http.StatusForbidden, "角色已分配用户，无法删除")
	ErrRoleBuiltin     = qierr.NewWithStatus(10105, http.StatusForbidden, "内置角色，不可操作")
	ErrRoleMenuQuery   = qierr.NewWithStatus(10106, http.StatusInternalServerError, "查询角色菜单失败")
	ErrRoleDeptQuery   = qierr.NewWithStatus(10107, http.StatusInternalServerError, "查询角色部门失败")
	ErrRoleCreate      = qierr.NewWithStatus(10108, http.StatusInternalServerError, "创建角色失败")
	ErrRoleUpdate      = qierr.NewWithStatus(10109, http.StatusInternalServerError, "更新角色失败")
	ErrRoleDelete      = qierr.NewWithStatus(10110, http.StatusInternalServerError, "删除角色失败")
	ErrRoleQuery       = qierr.NewWithStatus(10111, http.StatusInternalServerError, "查询角色失败")
)

// ===== 部门模块 102xx =====

var (
	ErrDeptNotFound    = qierr.NewWithStatus(10201, http.StatusNotFound, "部门不存在")
	ErrDeptCodeExists  = qierr.NewWithStatus(10202, http.StatusConflict, "部门编码已存在")
	ErrDeptHasChildren = qierr.NewWithStatus(10203, http.StatusForbidden, "存在子部门，无法删除")
	ErrDeptHasUsers    = qierr.NewWithStatus(10204, http.StatusForbidden, "部门下存在用户，无法删除")
	ErrDeptHasPosts    = qierr.NewWithStatus(10205, http.StatusForbidden, "部门下存在岗位，无法删除")
	ErrDeptCreate      = qierr.NewWithStatus(10206, http.StatusInternalServerError, "创建部门失败")
	ErrDeptUpdate      = qierr.NewWithStatus(10207, http.StatusInternalServerError, "更新部门失败")
	ErrDeptDelete      = qierr.NewWithStatus(10208, http.StatusInternalServerError, "删除部门失败")
	ErrDeptQuery       = qierr.NewWithStatus(10209, http.StatusInternalServerError, "查询部门失败")
)

// ===== 岗位模块 103xx =====

var (
	ErrPostNameExists    = qierr.NewWithStatus(10301, http.StatusConflict, "岗位名称已存在")
	ErrPostCodeExists    = qierr.NewWithStatus(10302, http.StatusConflict, "岗位编码已存在")
	ErrPostDeptQuery     = qierr.NewWithStatus(10303, http.StatusInternalServerError, "查询岗位部门失败")
	ErrPostCreate        = qierr.NewWithStatus(10304, http.StatusInternalServerError, "创建岗位失败")
	ErrPostDelete        = qierr.NewWithStatus(10305, http.StatusInternalServerError, "删除岗位失败")
	ErrPostHasUsers      = qierr.NewWithStatus(10306, http.StatusForbidden, "已分配用户该岗位")
	ErrPostHasUsersQuery = qierr.NewWithStatus(10307, http.StatusInternalServerError, "查询岗位用户失败")
	ErrPostDeleteUsers   = qierr.NewWithStatus(10308, http.StatusInternalServerError, "删除岗位用户关联失败")
	ErrPostNotFound      = qierr.NewWithStatus(10309, http.StatusNotFound, "岗位不存在")
	ErrPostUpdate        = qierr.NewWithStatus(10310, http.StatusInternalServerError, "更新岗位失败")
	ErrPostQuery         = qierr.NewWithStatus(10311, http.StatusInternalServerError, "查询岗位失败")
)
