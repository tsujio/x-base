package schemas

type PaginationInput struct {
	Page     int `schema:"page" validate:"gte=1"`
	PageSize int `schema:"pageSize" validate:"gte=0"`
}

type PaginatedList struct {
	TotalCount int64 `json:"total_count"`
	HasNext    bool  `json:"has_next"`
}
