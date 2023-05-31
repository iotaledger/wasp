package sm_snapshots

import (
	"github.com/samber/lo"
)

type sliceStructImpl[E any] struct {
	slice []E
}

var _ SliceStruct[int] = &sliceStructImpl[int]{}

func NewSliceStruct[E any](elems ...E) SliceStruct[E] {
	return &sliceStructImpl[E]{slice: elems}
}

func NewSliceStructLength[E any](length int) SliceStruct[E] {
	return NewSliceStructLengthCapacity[E](length, length)
}

func NewSliceStructLengthCapacity[E any](length, capacity int) SliceStruct[E] {
	return &sliceStructImpl[E]{slice: make([]E, length, capacity)}
}

func (s *sliceStructImpl[E]) Add(elem E) {
	s.slice = append(s.slice, elem)
}

func (s *sliceStructImpl[E]) Get(index int) E {
	return s.slice[index]
}

func (s *sliceStructImpl[E]) Set(index int, elem E) {
	s.slice[index] = elem
}

func (s *sliceStructImpl[E]) Length() int {
	return len(s.slice)
}

func (s *sliceStructImpl[E]) ForEach(forEachFun func(int, E) bool) bool {
	for index, elem := range s.slice {
		if !forEachFun(index, elem) {
			return false
		}
	}
	return true
}

func (s *sliceStructImpl[E]) Clone() SliceStruct[E] {
	return s.CloneDeep(func(elem E) E { return elem }) // NOTE: this is not deep cloning as the passed function is a simple identity
}

func (s *sliceStructImpl[E]) CloneDeep(cloneFun func(E) E) SliceStruct[E] {
	result := make([]E, s.Length())
	s.ForEach(func(index int, elem E) bool {
		result[index] = cloneFun(elem)
		return true
	})
	return NewSliceStruct(result...)
}

func (s *sliceStructImpl[E]) ContainsBy(fun func(E) bool) bool {
	return lo.ContainsBy(s.slice, fun)
}

func (s *sliceStructImpl[E]) Find(fun func(E) bool) (E, bool) {
	return lo.Find(s.slice, fun)
}
