package dict

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
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

func TestDetereminism(t *testing.T) {
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

func TestMarshaling(t *testing.T) {
	vars1 := New()
	vars1.Set("k1", []byte("kuku"))
	vars1.Set("k2", []byte{42})
	vars1.Set("k3", []byte("kuku"))
	vars1.Set("k4", []byte{2})

	var buf bytes.Buffer
	err := vars1.Write(&buf)
	assert.NoError(t, err)

	vars2 := New()
	err = vars2.Read(bytes.NewBuffer(buf.Bytes()))
	assert.NoError(t, err)

	assert.EqualValues(t, util.GetHashValue(vars1), util.GetHashValue(vars2))
}

func TestJSONMarshaling(t *testing.T) {
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
