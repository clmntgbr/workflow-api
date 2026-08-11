package paginate

import "math"

const (
	OrderByAsc   = "asc"
	OrderByDesc  = "desc"
	DefaultLimit = 20
)

type PaginateQuery struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	SortBy  string `query:"sortBy"`
	OrderBy string `query:"orderBy"`
	Search  string `query:"search"`
}

type PaginateResponse struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
	Members    any `json:"members"`
}

func NewPaginateResponse(members any, total int, query PaginateQuery) PaginateResponse {
	totalPages := int(math.Ceil(float64(total) / float64(query.Limit)))
	return PaginateResponse{
		Members:    members,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}
}

func (p *PaginateQuery) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = DefaultLimit
	}
	if p.OrderBy != OrderByAsc && p.OrderBy != OrderByDesc {
		p.OrderBy = OrderByAsc
	}
}

func (p *PaginateQuery) Offset() int {
	return (p.Page - 1) * p.Limit
}
