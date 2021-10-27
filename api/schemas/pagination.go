package schemas

type PaginationInput struct {
	Page     *int `schema:"page" validate:"omitempty,gte=1"`
	PageSize *int `schema:"pageSize" validate:"omitempty,gte=0,lte=100"`
}

type PaginatedList struct {
	TotalCount int64 `json:"totalCount"`
	HasNext    bool  `json:"hasNext"`
}
