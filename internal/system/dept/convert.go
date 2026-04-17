package dept

import (
	"github.com/tokmz/fox/internal/system/entity"
	"github.com/tokmz/qi/utils/pointer"
)

// createReqToEntity 创建请求转实体
func createReqToEntity(req *CreateReq) *entity.SysDept {
	return &entity.SysDept{
		ParentID: req.ParentID,
		Name:     req.Name,
		Code:     req.Code,
		DeptType: req.DeptType,
		LeaderID: req.LeaderID,
		Sort:     req.Sort,
		Status:   pointer.GetOrDefault(req.Status, int8(1)),
	}
}

// entityToDetailResp 实体转详情响应
func entityToDetailResp(e *entity.SysDept) *DetailResp {
	return &DetailResp{
		ID:        e.ID,
		ParentID:  e.ParentID,
		Name:      e.Name,
		Code:      e.Code,
		DeptType:  e.DeptType,
		LeaderID:  e.LeaderID,
		Sort:      e.Sort,
		Status:    e.Status,
		CreatedBy: e.CreatedBy,
		UpdatedBy: e.UpdatedBy,
		CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: e.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// entityToTreeResp 实体转树形节点
func entityToTreeResp(e *entity.SysDept) *TreeResp {
	return &TreeResp{
		ID:       e.ID,
		ParentID: e.ParentID,
		Name:     e.Name,
		Code:     e.Code,
		DeptType: e.DeptType,
		LeaderID: e.LeaderID,
		Sort:     e.Sort,
		Status:   e.Status,
	}
}

// entityToOptionResp 实体转选项
func entityToOptionResp(e *entity.SysDept) *OptionResp {
	return &OptionResp{
		ID:   e.ID,
		Name: e.Name,
		Code: e.Code,
	}
}
