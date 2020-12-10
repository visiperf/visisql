package visisql

import (
	"errors"
	"testing"

	"github.com/huandu/go-sqlbuilder"
	"github.com/stretchr/testify/assert"
)

// @todo: use assert and add message in TestWrapFuncs
func TestWrapFuncs(t *testing.T) {
	type in struct {
		predicate *Predicate
		val       string
	}

	type out struct {
		res string
	}

	type test struct {
		in  *in
		out *out
	}

	var tests = []*test{{
		in: &in{
			predicate: NewPredicate("table.id", OperatorEqual, []interface{}{1}),
			val:       "table.id",
		},
		out: &out{
			res: "table.id",
		},
	}, {
		in: &in{
			predicate: NewPredicate("table.id", OperatorEqual, []interface{}{1}, "unaccent"),
			val:       "table.id",
		},
		out: &out{
			res: "unaccent(table.id)",
		},
	}, {
		in: &in{
			predicate: NewPredicate("table.id", OperatorEqual, []interface{}{1}, "unaccent", "lower"),
			val:       "table.id",
		},
		out: &out{
			res: "lower(unaccent(table.id))",
		},
	}, {
		in: &in{
			predicate: NewPredicate("table.id", OperatorEqual, []interface{}{1}, "unaccent", "lower"),
			val:       "$1",
		},
		out: &out{
			res: "lower(unaccent($1))",
		},
	}}

	for _, test := range tests {
		if res := test.in.predicate.wrapFuncs(test.in.val); res != test.out.res {
			t.Errorf("resp was incorrect, got: %s, want: %s", res, test.out.res)
		}
	}
}

func TestPredicatesToString(t *testing.T) {
	type in struct {
		predicates [][]*Predicate
		cond       *sqlbuilder.Cond
	}

	type out struct {
		res []string
		err error
	}

	type test struct {
		message string
		in      *in
		out     *out
	}

	var tests = []*test{{
		message: "invalid values length with equals operator",
		in: &in{
			predicates: [][]*Predicate{{
				NewPredicate("table.id", OperatorEqual, []interface{}{1, 2}),
			}},
			cond: &sqlbuilder.PostgreSQL.NewSelectBuilder().Cond,
		},
		out: &out{
			res: nil,
			err: &QueryError{errOperatorEqual},
		},
	}, {
		message: "equal operator without funcs",
		in: &in{
			predicates: [][]*Predicate{{
				NewPredicate("table.field_1", OperatorEqual, []interface{}{1}),
			}, {
				NewPredicate("table.field_2", OperatorEqual, []interface{}{2}),
				NewPredicate("table.field_3", OperatorEqual, []interface{}{3}),
			}},
			cond: &sqlbuilder.PostgreSQL.NewSelectBuilder().Cond,
		},
		out: &out{
			res: []string{
				"( table.field_1 = $0 )",
				"( table.field_2 = $1 OR table.field_3 = $2 )",
			},
			err: nil,
		},
	}, {
		message: "equal operator with funcs",
		in: &in{
			predicates: [][]*Predicate{{
				NewPredicate("table.field_1", OperatorEqual, []interface{}{1}),
			}, {
				NewPredicate("table.field_2", OperatorEqual, []interface{}{2}, "unaccent"),
				NewPredicate("table.field_3", OperatorEqual, []interface{}{3}, "unaccent", "lower"),
			}},
			cond: &sqlbuilder.PostgreSQL.NewSelectBuilder().Cond,
		},
		out: &out{
			res: []string{
				"( table.field_1 = $0 )",
				"( unaccent(table.field_2) = unaccent($1) OR lower(unaccent(table.field_3)) = lower(unaccent($2)) )",
			},
			err: nil,
		},
	}, {
		message: "all operators with funcs",
		in: &in{
			predicates: [][]*Predicate{{
				NewPredicate("table.field_in", OperatorIn, []interface{}{1, 2, 3}, "lower"),
			}, {
				NewPredicate("table.field_equal", OperatorEqual, []interface{}{1}, "upper"),
			}, {
				NewPredicate("table.field_like", OperatorLike, []interface{}{"%value%"}, "unaccent"),
			}, {
				NewPredicate("table.field_null", OperatorIsNull, nil, "lower"),
			}, {
				NewPredicate("table.field_less_than", OperatorLessThan, []interface{}{123.456}, "round"),
			}, {
				NewPredicate("table.field_greater_than", OperatorGreaterThan, []interface{}{123.456}, "trunc"),
			}, {
				NewPredicate("table.field_between", OperatorBetween, []interface{}{2.1, 8.2}, "round"),
			}},
			cond: &sqlbuilder.NewSelectBuilder().Cond,
		},
		out: &out{
			res: []string{
				"( lower(table.field_in) IN (lower($0), lower($1), lower($2)) )",
				"( upper(table.field_equal) = upper($3) )",
				"( unaccent(table.field_like) LIKE unaccent($4) )",
				"( lower(table.field_null) IS NULL )",
				"( round(table.field_less_than) < round($5) )",
				"( trunc(table.field_greater_than) > trunc($6) )",
				"( round(table.field_between) BETWEEN round($7) AND round($8) )",
			},
			err: nil,
		},
	}}

	for _, test := range tests {
		res, err := predicatesToStrings(test.in.predicates, test.in.cond)

		if test.out.err != nil {
			var qe *QueryError

			assert.True(t, errors.As(err, &qe), test.message)
			assert.Equal(t, test.out.err, qe, test.message)
		} else {
			assert.Nil(t, err, test.message)
		}

		if test.out.res != nil {
			assert.Equal(t, test.out.res, res, test.message)
		} else {
			assert.Nil(t, res, test.message)
		}
	}
}
