package visisql

const (
	OperatorIn    string = "IN"
	OperatorEqual string = "EQUALS"
	OperatorLike  string = "LIKE"
)

type Predicate struct {
	field    string
	operator string
	values   []interface{}
}

func NewPredicate(field string, operator string, values []interface{}) *Predicate {
	return &Predicate{field: field, operator: operator, values: values}
}

func (p *Predicate) IsOperator(operator string) bool {
	return p.operator == operator
}
