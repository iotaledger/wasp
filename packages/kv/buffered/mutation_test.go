package buffered

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
)

func TestEmptyMutations(t *testing.T) {
	ms1 := NewMutations()
	ms2 := NewMutations()
	require.EqualValues(t, util.GetHashValue(ms1), util.GetHashValue(ms2))
}

func TestMutationsMarshalling(t *testing.T) {
	ms := NewMutations()
	ms.Set("k1", []byte("v1"))
	ms.Del("k2", true)

	ms2, err := MutationsFromBytes(ms.Bytes())
	require.NoError(t, err)
	require.EqualValues(t, util.GetHashValue(ms), util.GetHashValue(ms2))
}

func TestMutationsMisc(t *testing.T) {
	m := NewMutations()
	require.True(t, !m.Contains("kuku"))
	m.Del("kuku", true)
	require.True(t, m.Contains("kuku"))
	m.Set("kuku", []byte("v"))
	require.True(t, m.Contains("kuku"))
	m.Del("kuku", false)
	require.False(t, m.Contains("kuku"))
}
