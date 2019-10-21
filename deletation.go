package visisql

type Deletation struct {
	Predicates []*Predicate           `json:"predicates"`
}

/*
**	Usage example :
**
**	deletations := make(map[string]*visisql.Deletation)
**	deletations[pageRef] = visisql.NewDeletation(
**		[]*visisql.Predicate{
**			&visisql.Predicate{
**				Field: "facebook.pageref",
**				Operator: "IN",
v				Values: pageRefAsInterface,
**			},
**		},
**	)
*/

func NewDeletation(predicates []*Predicate) *Deletation {
	return &Deletation{Predicates: predicates}
}
