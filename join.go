package visisql

import "github.com/huandu/go-sqlbuilder"

type Join struct {
	table  string
	on     string
	option sqlbuilder.JoinOption
}

func NewJoin(table string, on string) *Join {
	return &Join{table: table, on: on}
}

func NewJoinOption(table string, on string, option sqlbuilder.JoinOption) *Join {
	return &Join{table: table, on: on, option: option}
}
