package buffered

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
)

func TestEmptyMutations(t *testing.T) {
	ms1 := NewMutations()
	ms2 := NewMutations()
	assert.EqualValues(t, util.GetHashValue(ms1), util.GetHashValue(ms2))
}

func TestMutationsMarshalling(t *testing.T) {
	ms := NewMutations()
	ms.Set("k1", []byte("v1"))
	ms.Del("k2")

	var buf bytes.Buffer
	err := ms.Write(&buf)
	assert.NoError(t, err)

	ms2 := NewMutations()
	err = ms2.Read(bytes.NewBuffer(buf.Bytes()))
	assert.NoError(t, err)

	assert.EqualValues(t, util.GetHashValue(ms), util.GetHashValue(ms2))
}

func TestMutationsMisc(t *testing.T) {
	m := NewMutations()
	require.True(t, !m.Contains("kuku"))
	m.Del("kuku")
	require.True(t, m.Contains("kuku"))
	m.Set("kuku", []byte("v"))
	require.True(t, m.Contains("kuku"))
	m.Del("kuku")
	require.True(t, m.Contains("kuku"))
}
