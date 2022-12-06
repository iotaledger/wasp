package util

type Cloneable[T any] interface {
	Clone() T
}

// CloneSlice clones every element in the the slice into cloned and returns the cloned slice.
func CloneSlice[T Cloneable[T]](base []T) []T {
	cloned := make([]T, len(base))

	for i := range base {
		cloned[i] = base[i].Clone()
	}

	return cloned
}

// CloneMap clones every element in the the map into cloned and returns the cloned map.
func CloneMap[K comparable, T Cloneable[T]](base map[K]T) map[K]T {
	cloned := make(map[K]T, len(base))

	for k, v := range base {
		cloned[k] = v.Clone()
	}

	return cloned
}

// KeyOnlyBy transforms a slice or an array of structs to a map based on a pivot callback.
func KeyOnlyBy[K comparable, V any](collection []V, iteratee func(V) K) map[K]struct{} {
	result := make(map[K]struct{}, len(collection))

	for _, v := range collection {
		k := iteratee(v)
		result[k] = struct{}{}
	}

	return result
}
