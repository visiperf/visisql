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

const (
	operatorIn          = "IN"
	operatorEqual       = "EQUALS"
	operatorLike        = "LIKE"
	operatorIsNull      = "IS NULL"
	operatorLessThan    = "LESS THAN"
	operatorGreaterThan = "GREATER THAN"
)

type Predicate interface {
	GetField() string
	GetValues() []interface{}
	IsOperator(operator string) bool
}

func predicatesToStrings(predicates [][]Predicate, cond *sqlbuilder.Cond) ([]string, error) {
	var andExprs []string
	for _, pAnd := range predicates {

		var orExprs []string
		for _, pOr := range pAnd {
			if pOr.IsOperator(operatorIn) {
				orExprs = append(orExprs, cond.In(pOr.GetField(), pOr.GetValues()...))
			}
			if pOr.IsOperator(operatorEqual) {
				if len(pOr.GetValues()) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorEqual})
				}
				orExprs = append(orExprs, cond.Equal(pOr.GetField(), pOr.GetValues()[0]))
			}
			if pOr.IsOperator(operatorLike) {
				if len(pOr.GetValues()) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorLike})
				}
				orExprs = append(orExprs, cond.Like(pOr.GetField(), pOr.GetValues()[0]))
			}
			if pOr.IsOperator(operatorIsNull) {
				if len(pOr.GetValues()) > 0 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorIsNull})
				}
				orExprs = append(orExprs, cond.IsNull(pOr.GetField()))
			}
			if pOr.IsOperator(operatorLessThan) {
				if len(pOr.GetValues()) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorLessThan})
				}
				orExprs = append(orExprs, cond.LessThan(pOr.GetField(), pOr.GetValues()[0]))
			}
			if pOr.IsOperator(operatorGreaterThan) {
				if len(pOr.GetValues()) != 1 {
					return nil, fmt.Errorf("visisql predicates: %w", &QueryError{errOperatorGreaterThan})
				}
				orExprs = append(orExprs, cond.GreaterThan(pOr.GetField(), pOr.GetValues()[0]))
			}
		}

		andExprs = append(andExprs, fmt.Sprintf("( %s )", strings.Join(orExprs, " OR ")))
	}

	return andExprs, nil
}
