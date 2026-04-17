package post

import (
	"github.com/tokmz/fox/internal/system/entity"
)

func CreateReqToEntity(req *CreateReq) *entity.SysPost {
	return &entity.SysPost{
		DeptID: req.DeptID,
		Name:   req.Name,
		Code:   req.Code,
		Sort:   req.Sort,
		Remark: req.Remark,
	}
}

func UpdateReqToEntity(req *UpdateReq) *entity.SysPost {
	return &entity.SysPost{
		ID:     req.ID,
		DeptID: req.DeptID,
		Name:   req.Name,
		Code:   req.Code,
		Sort:   req.Sort,
		Remark: req.Remark,
	}
}

func EntityToDetailResp(post *entity.SysPost) *DetailResp {
	return &DetailResp{
		ID:        post.ID,
		DeptID:    post.DeptID,
		Name:      post.Name,
		Code:      post.Code,
		Sort:      post.Sort,
		Remark:    post.Remark,
		Status:    post.Status,
		CreatedBy: post.CreatedBy,
		UpdatedBy: post.UpdatedBy,
		CreatedAt: post.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: post.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
