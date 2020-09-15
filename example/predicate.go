package main

type operator string

const (
	operatorIn          operator = "IN"
	operatorEqual       operator = "EQUALS"
	operatorLike        operator = "LIKE"
	operatorIsNull      operator = "IS NULL"
	operatorLessThan    operator = "LESS THAN"
	operatorGreaterThan operator = "GREATER THAN"
)

type predicate struct {
	field string
	operator
	values []interface{}
}

func newPredicate(field string, operator operator, values []interface{}) *predicate {
	return &predicate{field: field, operator: operator, values: values}
}

func (p *predicate) GetField() string {
	return p.field
}

func (p *predicate) GetValues() []interface{} {
	return p.values
}

func (p *predicate) IsOperator(operator string) bool {
	return operator == string(p.operator)
}
