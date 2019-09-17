package visisql

type Operator string

const (
	OperatorIn    Operator = "IN"
	OperatorEqual Operator = "EQUALS"
	OperatorLike  Operator = "LIKE"
)

type Predicate struct {
	Field    string
	Operator Operator
	Values   []interface{}
}

func NewPredicate(field string, operator Operator, values []interface{}) *Predicate {
	return &Predicate{Field: field, Operator: operator, Values: values}
}

func (p *Predicate) IsOperator(operator Operator) bool {
	return p.Operator == operator
}
