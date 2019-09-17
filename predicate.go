package visisql

const (
	OperatorIn    string = "IN"
	OperatorEqual string = "EQUALS"
	OperatorLike  string = "LIKE"
)

type Predicate struct {
	Field    string
	Operator string
	Values   []interface{}
}

func NewPredicate(field string, operator string, values []interface{}) *Predicate {
	return &Predicate{Field: field, Operator: operator, Values: values}
}

func (p *Predicate) IsOperator(operator string) bool {
	return p.Operator == operator
}
