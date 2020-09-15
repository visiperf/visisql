package main

type pagination struct {
	start int
	limit int
}

func newPagination(start, limit int) *pagination {
	return &pagination{start: start, limit: limit}
}

func (p *pagination) GetStart() int {
	return p.start
}

func (p *pagination) GetLimit() int {
	return p.limit
}
