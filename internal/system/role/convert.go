package role

import (
	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/qi/utils/pointer"
)

// createReqToEntity 创建请求转实体
func createReqToEntity(req *CreateReq) *entity.SysRole {
	return &entity.SysRole{
		ParentID:          req.ParentID,
		Name:              req.Name,
		Code:              req.Code,
		DataScope:         req.DataScope,
		DeptCheckStrictly: pointer.GetOrDefault(req.DeptCheckStrictly, true),
		Sort:              req.Sort,
		Status:            pointer.GetOrDefault(req.Status, int8(1)),
	}
}

// entityToDetailResp 实体转详情响应
func entityToDetailResp(e *entity.SysRole, menuIDs, deptIDs []int64) *DetailResp {
	return &DetailResp{
		ID:                e.ID,
		ParentID:          e.ParentID,
		Name:              e.Name,
		Code:              e.Code,
		DataScope:         e.DataScope,
		DeptCheckStrictly: e.DeptCheckStrictly,
		Builtin:           e.Builtin,
		Sort:              e.Sort,
		Status:            e.Status,
		MenuIDs:           menuIDs,
		DeptIDs:           deptIDs,
		CreatedBy:         e.CreatedBy,
		UpdatedBy:         e.UpdatedBy,
		CreatedAt:         e.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         e.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// entityToTreeResp 实体转树形节点
func entityToTreeResp(e *entity.SysRole) *TreeResp {
	return &TreeResp{
		ID:        e.ID,
		ParentID:  e.ParentID,
		Name:      e.Name,
		Code:      e.Code,
		DataScope: e.DataScope,
		Sort:      e.Sort,
		Status:    e.Status,
		Builtin:   e.Builtin,
	}
}

// entityToOptionResp 实体转选项
func entityToOptionResp(e *entity.SysRole) *OptionResp {
	return &OptionResp{
		ID:   e.ID,
		Name: e.Name,
		Code: e.Code,
	}
}

