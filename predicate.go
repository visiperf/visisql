package visisql

import (
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"strings"
)

type Operator string

const (
	OperatorIn     Operator = "IN"
	OperatorEqual  Operator = "EQUALS"
	OperatorLike   Operator = "LIKE"
	OperatorIsNull Operator = "IS NULL"
)

type Predicate struct {
	Field    string        `json:"field"`
	Operator Operator      `json:"operator"`
	Values   []interface{} `json:"values"`
}

func NewPredicate(field string, operator Operator, values []interface{}) *Predicate {
	return &Predicate{Field: field, Operator: operator, Values: values}
}

func (p *Predicate) IsOperator(operator Operator) bool {
	return p.Operator == operator
}

func predicatesToStrings(predicates [][]*Predicate, cond *sqlbuilder.Cond) ([]string, error) {
	var andExprs []string
	for _, pAnd := range predicates {

		var orExprs []string
		for _, pOr := range pAnd {
			if pOr.IsOperator(OperatorIn) {
				orExprs = append(orExprs, cond.In(pOr.Field, pOr.Values...))
			}
			if pOr.IsOperator(OperatorEqual) {
				if len(pOr.Values) != 1 {
					return nil, fmt.Errorf(`predicate must have only one value when operator is equal`)
				}
				orExprs = append(orExprs, cond.Equal(pOr.Field, pOr.Values[0]))
			}
			if pOr.IsOperator(OperatorLike) {
				if len(pOr.Values) != 1 {
					return nil, fmt.Errorf(`predicate must have only one value when operator is like`)
				}
				orExprs = append(orExprs, cond.Like(pOr.Field, pOr.Values[0]))
			}
			if pOr.IsOperator(OperatorIsNull) {
				if len(pOr.Values) > 0 {
					return nil, fmt.Errorf(`predicate should not have value(s) when operator is null`)
				}
				orExprs = append(orExprs, cond.IsNull(pOr.Field))
			}
		}

		andExprs = append(andExprs, fmt.Sprintf("( %s )", strings.Join(orExprs, " OR ")))
	}

	return andExprs, nil
}
