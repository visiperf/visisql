package visisql

type Deletation struct {
	Predicates []*Predicate           `json:"predicates"`
}

func NewDeletation(predicates []*Predicate) *Deletation {
	return &Deletation{Predicates: predicates}
}
