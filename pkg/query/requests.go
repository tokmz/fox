package query

import (
	"errors"
	"time"
)

// PageReq 分页请求
type PageReq struct {
	Page int `form:"page" json:"page" binding:"omitempty,min=1"           desc:"页码"     example:"1"  default:"1"`
	Size int `form:"size" json:"size" binding:"omitempty,oneof=10 20 50 100" desc:"每页条数" enum:"10,20,50,100" example:"10" default:"10"`
}

// Offset 计算 SQL 偏移量
func (p *PageReq) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 10
	}
	return (p.Page - 1) * p.Size
}

// TimeRangeReq 时间范围查询
type TimeRangeReq struct {
	StartTime string `form:"start_time" json:"start_time" binding:"omitempty" desc:"开始时间" example:"2024-01-01"`
	EndTime   string `form:"end_time" json:"end_time" binding:"omitempty"   desc:"结束时间" example:"2024-12-31"`
}

// Parse 将时间字符串解析为 time.Time，返回 (开始, 结束, error)
// 支持格式：2006-01-02 或 2006-01-02 15:04:05
func (t *TimeRangeReq) Parse() (start, end time.Time, err error) {
	if t.StartTime != "" {
		start, err = parseTime(t.StartTime)
		if err != nil {
			return
		}
	}
	if t.EndTime != "" {
		end, err = parseTime(t.EndTime)
		if err != nil {
			return
		}
		// 结束时间补到当天 23:59:59.999
		end = end.Add(time.Millisecond*999 + time.Second*59 + time.Minute*59)
	}
	return
}

var errInvalidTimeFormat = errors.New("时间格式无效，支持: 2006-01-02 或 2006-01-02 15:04:05")

func parseTime(s string) (time.Time, error) {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errInvalidTimeFormat
}
