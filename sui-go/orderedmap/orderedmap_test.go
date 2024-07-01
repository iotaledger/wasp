package orderedmap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/orderedmap"
	"github.com/iotaledger/wasp/sui-go/sui"
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
		m := orderedmap.New[sui.BuilderArg, sui.CallArg]()
		testBytes := [][]byte{
			{1, 4, 7},
			{2, 5, 8},
			{3, 6, 9},
			{10, 11, 12},
			{13, 14, 15},
		}
		m.Insert(sui.BuilderArg{Pure: &testBytes[0]}, sui.CallArg{Pure: &testBytes[0]})
		m.Insert(sui.BuilderArg{Pure: &testBytes[1]}, sui.CallArg{Pure: &testBytes[1]})

		val, ok := m.Get(sui.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, sui.CallArg{Pure: &testBytes[0]}, val)
		idx, ok := m.Find(sui.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, idx, 0)

		m.Insert(sui.BuilderArg{Pure: &testBytes[0]}, sui.CallArg{Pure: &testBytes[2]})
		val, ok = m.Get(sui.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, sui.CallArg{Pure: &testBytes[2]}, val)
		idx, ok = m.Find(sui.BuilderArg{Pure: &testBytes[0]})
		require.True(t, ok)
		require.Equal(t, idx, 0)

		var targetList []sui.CallArg = []sui.CallArg{
			{Pure: &testBytes[3]},
			{Pure: &testBytes[4]},
		}
		var testList []sui.CallArg
		i := 3
		m.ForEach(func(k sui.BuilderArg, v sui.CallArg) {
			testList = append(testList, sui.CallArg{Pure: &testBytes[i]})
			i++
		})
		require.Equal(t, targetList, testList)
	})
}