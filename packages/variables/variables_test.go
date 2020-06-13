package variables

import (
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicVariables(t *testing.T) {
	vars := New(nil)
	_, ok := vars.Get("v1")
	assert.False(t, ok)

	vars.Set("v1", uint16(1))
	v, ok := vars.Get("v1")
	assert.True(t, ok)

	vint, ok := v.(uint16)
	assert.True(t, ok)
	assert.Equal(t, vint, uint16(1))

	vars.Set("v2", "kuku")
	v, ok = vars.Get("v2")
	assert.True(t, ok)

	vstr, ok := v.(string)
	assert.True(t, ok)
	assert.Equal(t, vstr, "kuku")

	vars.Set("v1", nil)
	v, ok = vars.Get("v1")
	assert.True(t, ok)
	assert.Nil(t, v)
}

func TestBytes(t *testing.T) {
	vars1 := New(nil)
	h1 := util.GetHashValue(vars1)

	vars2 := New(vars1)
	h2 := util.GetHashValue(vars2)

	assert.EqualValues(t, h1, h2)

	vars1.Set("k1", "kuku")
	vars2.Set("k1", "kuku")

	h11 := util.GetHashValue(vars1)
	h12 := util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)

	vars1.Set("k1", "mumu")
	h11 = util.GetHashValue(vars1)
	assert.False(t, h11 == h12)

	vars1.Set("k1", nil)
	vars2.Set("k1", nil)
	h11 = util.GetHashValue(vars1)
	h12 = util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)

	vars1.Set("k1", uint16(42))
	vars1.Set("k2", "42")

	vars2.Set("k2", "42")
	vars2.Set("k1", uint16(42))
	h11 = util.GetHashValue(vars1)
	h12 = util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)
}

func TestDetereminism(t *testing.T) {
	vars1 := New(nil)
	h1 := util.GetHashValue(vars1)

	vars2 := New(vars1)
	h2 := util.GetHashValue(vars2)

	assert.EqualValues(t, h1, h2)

	vars1.Set("k1", "kuku")
	vars1.Set("k2", uint16(42))
	vars1.Set("k3", "kuku")
	vars1.Set("k4", uint16(2))

	vars2.Set("k4", uint16(2))
	vars2.Set("k3", "kuku")
	vars2.Set("k2", uint16(42))
	vars2.Set("k1", "kuku")

	h11 := util.GetHashValue(vars1)
	h12 := util.GetHashValue(vars2)
	assert.EqualValues(t, h11, h12)

	vars1.Set("k3", nil)
	vars2.Set("k3", nil)
	assert.EqualValues(t, h11, h12)

	t.Logf("\n%s", vars1.String())
	t.Logf("\n%s", vars2.String())
}

func TestApply(t *testing.T) {
	vars1 := New(nil)
	vars2 := New(vars1)

	vars1.Set("k1", "kuku")
	vars1.Set("k2", uint16(42))
	vars1.Set("k3", "kuku")
	vars1.Set("k4", uint16(2))

	t.Logf("vars1 pries:\n%s", vars1.String())

	vars2.Set("k4", nil)
	vars2.Set("k3", "mumu")
	vars2.Set("k2", uint16(314))
	t.Logf("vars2:\n%s", vars2.String())

	vars1.Apply(vars2)
	t.Logf("vars1 po:\n%s", vars1.String())
}
