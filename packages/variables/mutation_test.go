package variables

import (
	"bytes"
	"testing"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
)

func TestApplyMutationSet(t *testing.T) {
	vars := New(nil)

	mset := NewMutationSet("k1", []byte("v1"))
	mset.ApplyTo(vars)

	v, ok := vars.Get("k1")
	assert.Equal(t, true, ok)
	assert.Equal(t, []byte("v1"), v)
}

func TestApplyMutationDel(t *testing.T) {
	vars := New(nil)
	vars.Set("k1", []byte("v1"))

	mset := NewMutationDel("k1")
	mset.ApplyTo(vars)

	_, ok := vars.Get("k1")
	assert.Equal(t, false, ok)
}

func TestEmptyMutationSequence(t *testing.T) {
	ms1 := NewMutationSequence()
	ms2 := NewMutationSequence()
	assert.EqualValues(t, util.GetHashValue(ms1), util.GetHashValue(ms2))
}

func TestMutationSequenceMarshaling(t *testing.T) {
	ms := NewMutationSequence()
	ms.Add(NewMutationSet("k1", []byte("v1")))
	ms.Add(NewMutationDel("k2"))

	var buf bytes.Buffer
	err := ms.Write(&buf)
	assert.NoError(t, err)

	ms2 := NewMutationSequence()
	err = ms2.Read(bytes.NewBuffer(buf.Bytes()))
	assert.NoError(t, err)

	assert.EqualValues(t, util.GetHashValue(ms), util.GetHashValue(ms2))
}
