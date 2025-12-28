package domain

type Elem[V any] struct {
	ID    uint64
	Value V
}
