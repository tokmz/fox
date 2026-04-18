package menu

import (
	"github.com/tokmz/fox/internal/system/entity"
)

// entityToDetailResp 将实体转换为详情响应
func entityToDetailResp(menu *entity.SysMenu, apiIDs []int64) *DetailResp {
	return &DetailResp{
		ID:           menu.ID,
		ParentID:     menu.ParentID,
		Title:        menu.Title,
		Key:          menu.Key,
		Path:         menu.Path,
		Component:    menu.Component,
		Redirect:     menu.Redirect,
		Query:        menu.Query,
		MenuType:     menu.MenuType,
		OpenType:     menu.OpenType,
		Icon:         menu.Icon,
		Sort:         menu.Sort,
		KeepAlive:    menu.KeepAlive,
		Hidden:       menu.Hidden,
		Affix:        menu.Affix,
		AlwaysShow:   menu.AlwaysShow,
		ActiveMenu:   menu.ActiveMenu,
		FrameSrc:     menu.FrameSrc,
		ExternalLink: menu.ExternalLink,
		Remark:       menu.Remark,
		Status:       menu.Status,
		ApiIDs:       apiIDs,
		CreatedAt:    menu.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    menu.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// entityToTreeResp 将实体转换为树形节点
func entityToTreeResp(menu *entity.SysMenu) *TreeResp {
	return &TreeResp{
		ID:           menu.ID,
		ParentID:     menu.ParentID,
		Title:        menu.Title,
		Key:          menu.Key,
		Path:         menu.Path,
		Component:    menu.Component,
		Redirect:     menu.Redirect,
		Icon:         menu.Icon,
		MenuType:     menu.MenuType,
		OpenType:     menu.OpenType,
		Sort:         menu.Sort,
		KeepAlive:    menu.KeepAlive,
		Hidden:       menu.Hidden,
		Affix:        menu.Affix,
		AlwaysShow:   menu.AlwaysShow,
		ActiveMenu:   menu.ActiveMenu,
		FrameSrc:     menu.FrameSrc,
		ExternalLink: menu.ExternalLink,
		Status:       menu.Status,
		Children:     make([]*TreeResp, 0),
	}
}

// entityToOptionResp 将实体转换为选项节点
func entityToOptionResp(menu *entity.SysMenu) *OptionResp {
	return &OptionResp{
		ID:       menu.ID,
		ParentID: menu.ParentID,
		Title:    menu.Title,
		Key:      menu.Key,
		MenuType: menu.MenuType,
		Children: make([]*OptionResp, 0),
	}
}
