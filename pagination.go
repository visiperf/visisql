package visisql

type Pagination struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

func NewPagination(start, limit int) *Pagination {
	return &Pagination{Start: start, Limit: limit}
}
