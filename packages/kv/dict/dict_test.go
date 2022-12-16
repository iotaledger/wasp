package dict

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

func TestBasicKVMap(t *testing.T) {
	vars := New()

	v, err := vars.Get("k1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	vars.Set("k1", []byte("v1"))
	v, err = vars.Get("k1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	vars.Del("k1")
	v, err = vars.Get("v1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestBytes(t *testing.T) {
	vars1 := New()
	h1 := util.GetHashValue(vars1)

	vars2 := vars1.Clone()
	h2 := util.GetHashValue(vars2)

	assert.EqualValues(t, h1, h2)

	vars1.Set("k1", []byte("kuku"))
	vars2.Set("k1", []byte("kuku"))

	h11 := util.GetHashValue(vars1)
	h12 := util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)

	vars1.Set("k1", []byte("mumu"))
	h11 = util.GetHashValue(vars1)
	assert.False(t, h11 == h12)

	vars1.Del("k1")
	vars2.Del("k1")
	h11 = util.GetHashValue(vars1)
	h12 = util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)

	vars1.Set("k1", []byte{42})
	vars1.Set("k2", []byte("42"))

	vars2.Set("k2", []byte("42"))
	vars2.Set("k1", []byte{42})
	h11 = util.GetHashValue(vars1)
	h12 = util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)
}

func TestDeterminism(t *testing.T) {
	vars1 := New()
	h1 := util.GetHashValue(vars1)

	vars2 := vars1.Clone()
	h2 := util.GetHashValue(vars2)

	assert.EqualValues(t, h1, h2)

	vars1.Set("k1", []byte("kuku"))
	vars1.Set("k2", []byte{42})
	vars1.Set("k3", []byte("kuku"))
	vars1.Set("k4", []byte{2})

	vars2.Set("k4", []byte{2})
	vars2.Set("k3", []byte("kuku"))
	vars2.Set("k2", []byte{42})
	vars2.Set("k1", []byte("kuku"))

	h11 := util.GetHashValue(vars1)
	h12 := util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)

	vars1.Del("k3")
	vars2.Del("k3")
	assert.EqualValues(t, h11, h12)

	t.Logf("\n%s", vars1.String())
	t.Logf("\n%s", vars2.String())
}

func TestIterateSorted(t *testing.T) {
	d := New()
	d.Set("x", []byte("x"))
	d.Set("k5", []byte("v5"))
	d.Set("k1", []byte("v1"))
	d.Set("k3", []byte("v3"))
	d.Set("k2", []byte("v2"))
	d.Set("k4", []byte("v4"))

	var seen []kv.Key
	err := d.IterateSorted("", func(k kv.Key, v []byte) bool {
		seen = append(seen, k)
		return true
	})
	require.NoError(t, err)
	require.Equal(t, []kv.Key{"k1", "k2", "k3", "k4", "k5", "x"}, seen)

	seen = nil
	err = d.IterateSorted("k", func(k kv.Key, v []byte) bool {
		seen = append(seen, k)
		return true
	})
	require.NoError(t, err)
	require.Equal(t, []kv.Key{"k1", "k2", "k3", "k4", "k5"}, seen)
}

func TestMarshalling(t *testing.T) {
	vars1 := New()
	vars1.Set("k1", []byte("kuku"))
	vars1.Set("k2", []byte{42})
	vars1.Set("k3", []byte("kuku"))
	vars1.Set("k4", []byte{2})

	vars2, err := FromBytes(vars1.Bytes())
	assert.NoError(t, err)
	require.EqualValues(t, vars1.Bytes(), vars2.Bytes())
}

func TestJSONMarshalling(t *testing.T) {
	vars1 := New()
	vars1.Set("k1", []byte("kuku"))
	vars1.Set("k2", []byte{42})
	vars1.Set("k3", []byte("kuku"))
	vars1.Set("k4", []byte{2})

	b, err := json.Marshal(vars1)
	assert.NoError(t, err)

	var vars2 Dict
	err = json.Unmarshal(b, &vars2)
	assert.NoError(t, err)

	assert.EqualValues(t, util.GetHashValue(vars1), util.GetHashValue(vars2))
}
