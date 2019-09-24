package visisql

type Mutation struct {
	Set        map[string]interface{} `json:"set"`
	Predicates []*Predicate           `json:"predicates"`
}

func NewMutation(set map[string]interface{}, predicates []*Predicate) *Mutation {
	return &Mutation{Set: set, Predicates: predicates}
}
