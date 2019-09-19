package visisql

type Mutation struct {
	Set        map[string]interface{}
	Predicates []*Predicate
}

func NewMutation(set map[string]interface{}, predicates []*Predicate) *Mutation {
	return &Mutation{Set: set, Predicates: predicates}
}
