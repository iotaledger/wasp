// Package iscmove implements ISC move operations and related functionality.
package iscmove

type Bag = map[any]any

// Table is a generic map type. For now just experimenting with typings.
// Do we want to keep the isc structs as close as possible to the Sui types or just use map[K]V directly?
type Table[K comparable, V any] struct {
	Data map[K]V
}

func NewTable[K comparable, V any]() Table[K, V] {
	return Table[K, V]{Data: make(map[K]V)}
}
