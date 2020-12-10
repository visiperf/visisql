package visisql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/huandu/go-sqlbuilder"
)

var errOperatorEqual = errors.New("predicate must have only one value when operator is equal")
var errOperatorLike = errors.New("predicate must have only one value when operator is like")
var errOperatorIsNull = errors.New("predicate should not have value(s) when operator is null")
var errOperatorLessThan = errors.New("predicate must have only one value when operator is less than")
var errOperatorGreaterThan = errors.New("predicate must have only one value when operator is greater than")
var errOperatorBetween = errors.New("predicate must have two values when operator is between")

type Operator string

const (
	OperatorIn          Operator = "IN"
	OperatorEqual       Operator = "EQUALS"
	OperatorLike        Operator = "LIKE"
	OperatorIsNull      Operator = "IS NULL"
	OperatorLessThan    Operator = "LESS THAN"
	OperatorGreaterThan Operator = "GREATER THAN"
	OperatorBetween     Operator = "BETWEEN"
)

type Predicate struct {
	Field    string        `json:"field"`
	Operator Operator      `json:"operator"`
	Values   []interface{} `json:"values"`
	Funcs    []string
}

func NewPredicate(field string, operator Operator, values []interface{}, funcs ...string) *Predicate {
	return &Predicate{Field: field, Operator: operator, Values: values, Funcs: funcs}
}

func (p *Predicate) IsOperator(operator Operator) bool {
	return p.Operator == operator
}

func (p *Predicate) wrapFuncs(val string) string {
	var s = "%s"
	for _, f := range p.Funcs {
		s = fmt.Sprintf("%s(%s)", f, s)
	}

	return fmt.Sprintf(s, val)
}

func predicatesToStrings(predicates [][]*Predicate, cond *sqlbuilder.Cond) ([]string, error) {
	var andExprs []string
	for _, pAnd := range predicates {

		var orExprs []string
		for _, pOr := range pAnd {
			if pOr.IsOperator(OperatorIn) {
				vs := make([]string, 0, len(pOr.Values))

				for _, v := range pOr.Values {
					vs = append(vs, pOr.wrapFuncs(cond.Args.Add(v)))
				}

				orExprs = append(orExprs, fmt.Sprintf("%s IN (%s)", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field)), strings.Join(vs, ", ")))
			}
			if pOr.IsOperator(OperatorEqual) {
				if len(pOr.Values) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorEqual})
				}
				orExprs = append(orExprs, fmt.Sprintf("%s = %s", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field)), pOr.wrapFuncs(cond.Args.Add(pOr.Values[0]))))
			}
			if pOr.IsOperator(OperatorLike) {
				if len(pOr.Values) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorLike})
				}
				orExprs = append(orExprs, fmt.Sprintf("%s LIKE %s", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field)), pOr.wrapFuncs(cond.Args.Add(pOr.Values[0]))))
			}
			if pOr.IsOperator(OperatorIsNull) {
				if len(pOr.Values) > 0 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorIsNull})
				}
				orExprs = append(orExprs, fmt.Sprintf("%s IS NULL", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field))))
			}
			if pOr.IsOperator(OperatorLessThan) {
				if len(pOr.Values) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorLessThan})
				}
				orExprs = append(orExprs, fmt.Sprintf("%s < %s", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field)), pOr.wrapFuncs(cond.Args.Add(pOr.Values[0]))))
			}
			if pOr.IsOperator(OperatorGreaterThan) {
				if len(pOr.Values) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorGreaterThan})
				}
				orExprs = append(orExprs, fmt.Sprintf("%s > %s", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field)), pOr.wrapFuncs(cond.Args.Add(pOr.Values[0]))))
			}
			if pOr.IsOperator(OperatorBetween) {
				if len(pOr.Values) != 2 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorBetween})
				}
				orExprs = append(orExprs, fmt.Sprintf("%s BETWEEN %s AND %s", pOr.wrapFuncs(sqlbuilder.Escape(pOr.Field)), pOr.wrapFuncs(cond.Args.Add(pOr.Values[0])), pOr.wrapFuncs(cond.Args.Add(pOr.Values[1]))))
			}
		}

		andExprs = append(andExprs, fmt.Sprintf("( %s )", strings.Join(orExprs, " OR ")))
	}

	return andExprs, nil
}
