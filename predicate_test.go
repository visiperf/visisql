package visisql

import "testing"

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
