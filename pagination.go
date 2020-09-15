package visisql

type Pagination interface {
	GetStart() int
	GetLimit() int
}
