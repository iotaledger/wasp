package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type MockEquatable int

func NewMockEquatable(i int) *MockEquatable {
	result := MockEquatable(i)
	return &result
}

func (meT *MockEquatable) Equals(other Equatable) bool {
	meO, ok := other.(*MockEquatable)
	if !ok {
		return false
	}
	return *meT == *meO
}

func TestContains(t *testing.T) {
	me0 := NewMockEquatable(1)
	me1 := NewMockEquatable(10)
	me2 := NewMockEquatable(5)
	me3 := NewMockEquatable(2)
	sliceEmpty := []*MockEquatable{}
	require.False(t, Contains(me0, sliceEmpty))
	require.False(t, Contains(me1, sliceEmpty))
	require.False(t, Contains(me2, sliceEmpty))
	require.False(t, Contains(me3, sliceEmpty))

	sliceOne := []*MockEquatable{me1}
	require.False(t, Contains(me0, sliceOne))
	require.True(t, Contains(me1, sliceOne))
	require.False(t, Contains(me2, sliceOne))
	require.False(t, Contains(me3, sliceOne))

	sliceWoOne := []*MockEquatable{me0, me1, me3}
	require.True(t, Contains(me0, sliceWoOne))
	require.True(t, Contains(me1, sliceWoOne))
	require.False(t, Contains(me2, sliceWoOne))
	require.True(t, Contains(me3, sliceWoOne))

	sliceAll := []*MockEquatable{me0, me1, me2, me3}
	require.True(t, Contains(me0, sliceAll))
	require.True(t, Contains(me1, sliceAll))
	require.True(t, Contains(me2, sliceAll))
	require.True(t, Contains(me3, sliceAll))

	// Test if Equals is called
	me0 = NewMockEquatable(1)
	me1 = NewMockEquatable(10)
	me2 = NewMockEquatable(5)
	me3 = NewMockEquatable(2)
	require.True(t, Contains(me0, sliceAll))
	require.True(t, Contains(me1, sliceAll))
	require.True(t, Contains(me2, sliceAll))
	require.True(t, Contains(me3, sliceAll))
}

func TestRemove(t *testing.T) {
	me0 := NewMockEquatable(1)
	me1 := NewMockEquatable(10)
	me2 := NewMockEquatable(5)
	me3 := NewMockEquatable(2)
	slice := Remove(me2, []*MockEquatable{})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = Remove(me2, []*MockEquatable{me0, me1})
	require.True(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 2, len(slice))

	slice = Remove(me2, []*MockEquatable{me0, me1, me2, me3})
	require.True(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.True(t, Contains(me3, slice))
	require.Equal(t, 3, len(slice))

	slice = Remove(me0, []*MockEquatable{me0, me1, me2, me3})
	require.False(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.True(t, Contains(me3, slice))
	require.Equal(t, 3, len(slice))

	slice = Remove(me3, []*MockEquatable{me0, me1, me2, me3})
	require.True(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 3, len(slice))
}

func TestRemoveAll(t *testing.T) {
	me0 := NewMockEquatable(1)
	me1 := NewMockEquatable(10)
	me2 := NewMockEquatable(5)
	me3 := NewMockEquatable(2)
	slice := RemoveAll([]*MockEquatable{}, []*MockEquatable{})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = RemoveAll([]*MockEquatable{}, []*MockEquatable{me0, me1, me2, me3})
	require.True(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.True(t, Contains(me3, slice))
	require.Equal(t, 4, len(slice))

	slice = RemoveAll([]*MockEquatable{me0, me1, me2, me3}, []*MockEquatable{})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = RemoveAll([]*MockEquatable{me0, me1}, []*MockEquatable{me2, me3})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.True(t, Contains(me3, slice))
	require.Equal(t, 2, len(slice))

	slice = RemoveAll([]*MockEquatable{me1}, []*MockEquatable{me0, me1, me2, me3})
	require.True(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.True(t, Contains(me3, slice))
	require.Equal(t, 3, len(slice))

	slice = RemoveAll([]*MockEquatable{me3, me1}, []*MockEquatable{me0, me1, me2, me3})
	require.True(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 2, len(slice))

	slice = RemoveAll([]*MockEquatable{me3, me1, me0}, []*MockEquatable{me1, me2, me3})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 1, len(slice))

	slice = RemoveAll([]*MockEquatable{me2, me0, me3, me1}, []*MockEquatable{me0, me1, me2, me3})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))
}

func TestIntersection(t *testing.T) {
	me0 := NewMockEquatable(1)
	me1 := NewMockEquatable(10)
	me2 := NewMockEquatable(5)
	me3 := NewMockEquatable(2)
	slice := Intersection([]*MockEquatable{}, []*MockEquatable{})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = Intersection([]*MockEquatable{}, []*MockEquatable{me0, me1, me2, me3})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = Intersection([]*MockEquatable{me0, me1, me2, me3}, []*MockEquatable{})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = Intersection([]*MockEquatable{me0, me1}, []*MockEquatable{me2, me3})
	require.False(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 0, len(slice))

	slice = Intersection([]*MockEquatable{me0}, []*MockEquatable{me0, me1, me2, me3})
	require.True(t, Contains(me0, slice))
	require.False(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 1, len(slice))

	slice = Intersection([]*MockEquatable{me0, me1, me2, me3}, []*MockEquatable{me1})
	require.False(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.False(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 1, len(slice))

	slice = Intersection([]*MockEquatable{me0, me1, me2}, []*MockEquatable{me2, me1, me3})
	require.False(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.False(t, Contains(me3, slice))
	require.Equal(t, 2, len(slice))

	slice = Intersection([]*MockEquatable{me0, me1, me2, me3}, []*MockEquatable{me2, me1, me3, me0})
	require.True(t, Contains(me0, slice))
	require.True(t, Contains(me1, slice))
	require.True(t, Contains(me2, slice))
	require.True(t, Contains(me3, slice))
	require.Equal(t, 4, len(slice))
}

func TestAllDifferent(t *testing.T) {
	me0 := NewMockEquatable(1)
	me1 := NewMockEquatable(10)
	me2 := NewMockEquatable(5)
	me3 := NewMockEquatable(2)
	require.True(t, AllDifferent([]*MockEquatable{}))
	require.True(t, AllDifferent([]*MockEquatable{me0}))
	require.True(t, AllDifferent([]*MockEquatable{me0, me1, me2, me3}))
	require.False(t, AllDifferent([]*MockEquatable{me0, me1, me0, me2, me3}))
	require.False(t, AllDifferent([]*MockEquatable{me0, me0}))
	require.False(t, AllDifferent([]*MockEquatable{me0, me1, me2, me3, me2}))
}
