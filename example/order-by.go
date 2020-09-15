package main

type order string

const (
	orderAsc  order = "ASC"
	orderDesc order = "DESC"
)

type orderBy struct {
	field string
	order order
}

func newOrderBy(field string, order order) *orderBy {
	return &orderBy{field: field, order: order}
}

func (ob *orderBy) GetField() string {
	return ob.field
}

func (ob *orderBy) GetOrder() string {
	return string(ob.order)
}
