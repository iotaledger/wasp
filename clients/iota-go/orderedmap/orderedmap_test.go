package orderedmap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/orderedmap"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
)

func TestOrderedMap(t *testing.T) {
	t.Run("primitive type", func(t *testing.T) {
		m := orderedmap.New[string, int]()
		m.Insert("first", 1)
		m.Insert("second", 2)

		val, ok := m.Get("first")
		require.True(t, ok)
		require.Equal(t, 1, val)
		idx, ok := m.Find("first")
		require.True(t, ok)
		require.Equal(t, idx, 0)

		m.Insert("first", 3)
		val, ok = m.Get("first")
		require.True(t, ok)
		require.Equal(t, 3, val)
		idx, ok = m.Find("first")
		require.True(t, ok)
		require.Equal(t, idx, 0)

		var targetList []int = []int{6, 4}
		var testList []int
		m.ForEach(func(k string, v int) {
			testList = append(testList, v*2)
		})
		require.Equal(t, targetList, testList)
	})

	t.Run("customized type", func(t *testing.T) {
		m := orderedmap.New[sui2.BuilderArg, sui2.CallArg]()
		testBytes := [][]byte{
			{1, 4, 7},
			{2, 5, 8},
			{3, 6, 9},
			{10, 11, 12},
			{13, 14, 15},
		}
		m.Insert(sui2.BuilderArg{Pure: &testBytes[0]}, sui2.CallArg{Pure: &testBytes[0]})
		m.Insert(sui2.BuilderArg{Pure: &testBytes[1]}, sui2.CallArg{Pure: &testBytes[1]})

		val, ok := m.Get(sui2.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, sui2.CallArg{Pure: &testBytes[0]}, val)
		idx, ok := m.Find(sui2.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, idx, 0)

		m.Insert(sui2.BuilderArg{Pure: &testBytes[0]}, sui2.CallArg{Pure: &testBytes[2]})
		val, ok = m.Get(sui2.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, sui2.CallArg{Pure: &testBytes[2]}, val)
		idx, ok = m.Find(sui2.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, idx, 0)

		var targetList []sui2.CallArg = []sui2.CallArg{
			{Pure: &testBytes[3]},
			{Pure: &testBytes[4]},
		}
		var testList []sui2.CallArg
		i := 3
		m.ForEach(func(k sui2.BuilderArg, v sui2.CallArg) {
			testList = append(testList, sui2.CallArg{Pure: &testBytes[i]})
			i++
		})
		require.Equal(t, targetList, testList)
	})
}