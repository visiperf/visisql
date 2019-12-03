package visisql

import "fmt"

type Order string

const (
	OrderAsc  Order = "ASC"
	OrderDesc Order = "DESC"
)

type OrderBy struct {
	Field string `json:"field"`
	Order Order  `json:"order"`
}

func NewOrderBy(field string, order Order) *OrderBy {
	return &OrderBy{Field: field, Order: order}
}

func (ob *OrderBy) toString() string {
	return fmt.Sprintf("%s %s", ob.Field, ob.Order)
}
