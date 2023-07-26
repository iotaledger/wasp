package util

import (
	"github.com/samber/lo"
)

// Putting slice into a map is not acceptable as if you want to append to slice,
// you'll have to re-include the appended slice into the map.
type SliceStruct[E any] struct {
	slice []E
}

func NewSliceStruct[E any](elems ...E) *SliceStruct[E] {
	return &SliceStruct[E]{slice: elems}
}

func NewSliceStructLength[E any](length int) *SliceStruct[E] {
	return NewSliceStructLengthCapacity[E](length, length)
}

func NewSliceStructLengthCapacity[E any](length, capacity int) *SliceStruct[E] {
	return &SliceStruct[E]{slice: make([]E, length, capacity)}
}

func (s *SliceStruct[E]) Add(elem E) {
	s.slice = append(s.slice, elem)
}

func (s *SliceStruct[E]) Get(index int) E {
	return s.slice[index]
}

func (s *SliceStruct[E]) Set(index int, elem E) {
	s.slice[index] = elem
}

func (s *SliceStruct[E]) Length() int {
	return len(s.slice)
}

func (s *SliceStruct[E]) ForEach(forEachFun func(int, E) bool) bool {
	for index, elem := range s.slice {
		if !forEachFun(index, elem) {
			return false
		}
	}
	return true
}

// Returns a reference to new SliceStruct with exactly the same elements
func (s *SliceStruct[E]) Clone() *SliceStruct[E] {
	return s.CloneDeep(func(elem E) E { return elem }) // NOTE: this is not deep cloning as the passed function is a simple identity
}

// Returns a reference to new SliceStruct with every element of the old SliceStruct cloned using provided function
func (s *SliceStruct[E]) CloneDeep(cloneFun func(E) E) *SliceStruct[E] {
	result := make([]E, s.Length())
	s.ForEach(func(index int, elem E) bool {
		result[index] = cloneFun(elem)
		return true
	})
	return NewSliceStruct(result...)
}

func (s *SliceStruct[E]) ContainsBy(fun func(E) bool) bool {
	return lo.ContainsBy(s.slice, fun)
}

func (s *SliceStruct[E]) Find(fun func(E) bool) (E, bool) {
	return lo.Find(s.slice, fun)
}
