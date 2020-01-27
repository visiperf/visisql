package visisql

type JoinOption string

const (
	LeftJoin  JoinOption = "LEFT"
	InnerJoin JoinOption = "INNER"
	RightJoin JoinOption = "RIGHT"
)

type Join struct {
	option JoinOption
	table  string
	on     string
}

func NewJoin(option JoinOption, table string, on string) *Join {
	return &Join{option: option, table: table, on: on}
}
