package query

// PageResp 分页响应
type PageResp[T any] struct {
	List  []T   `json:"list"  desc:"数据列表"`
	Total int64 `json:"total" desc:"总条数"`
}
