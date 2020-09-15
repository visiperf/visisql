package visisql

type OrderBy interface {
	GetField() string
	GetOrder() string
}
