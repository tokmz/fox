package user

import (
	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/qi/utils/pointer"
)

// createReqToEntity 创建请求转实体
func createReqToEntity(req *CreateReq) *entity.SysUser {
	return &entity.SysUser{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Gender:   pointer.GetOrDefault(req.Gender, int8(0)),
		DeptID:   req.DeptID,
		Remark:   req.Remark,
		Status:   pointer.GetOrDefault(req.Status, int8(1)),
	}
}

// entityToDetailResp 实体转详情响应（含关联数据）
func entityToDetailResp(e *entity.SysUser, deptName string, roleIDs []int64, roleNames []string, postIDs []int64, postNames []string) *DetailResp {
	return &DetailResp{
		ID:        e.ID,
		Username:  e.Username,
		Nickname:  e.Nickname,
		Email:     e.Email,
		Phone:     e.Phone,
		Avatar:    e.Avatar,
		Gender:    e.Gender,
		DeptID:    e.DeptID,
		DeptName:  deptName,
		Remark:    e.Remark,
		Status:    e.Status,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
		PostIDs:   postIDs,
		PostNames: postNames,
		CreatedBy: e.CreatedBy,
		UpdatedBy: e.UpdatedBy,
		CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: e.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// entityToListItemResp 实体转列表项（含部门/角色/岗位名称）
func entityToListItemResp(e *entity.SysUser, deptName string, roleNames, postNames []string) *ListItemResp {
	return &ListItemResp{
		ID:        e.ID,
		Username:  e.Username,
		Nickname:  e.Nickname,
		Email:     e.Email,
		Phone:     e.Phone,
		Avatar:    e.Avatar,
		Gender:    e.Gender,
		DeptID:    e.DeptID,
		DeptName:  deptName,
		Status:    e.Status,
		RoleNames: roleNames,
		PostNames: postNames,
		CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// entityToOptionResp 实体转选项
func entityToOptionResp(e *entity.SysUser, deptName string) *OptionResp {
	return &OptionResp{
		ID:       e.ID,
		Username: e.Username,
		Nickname: e.Nickname,
		DeptName: deptName,
	}
}
