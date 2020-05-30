package variables

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestBasicVariables(t *testing.T) {
	vars := New(nil)
	_, ok := vars.Get("v1")
	assert.Equal(t, ok, false)

	vars.Set("v1", uint16(1))
	v, ok := vars.Get("v1")
	assert.Equal(t, ok, true)

	vint, ok := v.(uint16)
	assert.Equal(t, ok, true)
	assert.Equal(t, vint, uint16(1))

	vars.Set("v2", "kuku")
	v, ok = vars.Get("v2")
	assert.Equal(t, ok, true)

	vstr, ok := v.(string)
	assert.Equal(t, ok, true)
	assert.Equal(t, vstr, "kuku")

	vars.Set("v1", nil)
	v, ok = vars.Get("v1")
	assert.Equal(t, ok, false)
}

func TestBytes(t *testing.T) {
	vars1 := New(nil)
	h1 := hashing.GetHashValue(vars1)

	vars2 := New(vars1)
	h2 := hashing.GetHashValue(vars2)

	assert.Equal(t, h1 == h2, true)

	vars1.Set("k1", "kuku")
	vars2.Set("k1", "kuku")

	h11 := hashing.GetHashValue(vars1)
	h12 := hashing.GetHashValue(vars2)
	assert.Equal(t, h11 == h12, true)

	vars1.Set("k1", "mumu")
	h11 = hashing.GetHashValue(vars1)
	assert.Equal(t, h11 != h12, true)

	vars1.Set("k1", nil)
	vars2.Set("k1", nil)
	h11 = hashing.GetHashValue(vars1)
	h12 = hashing.GetHashValue(vars2)
	assert.Equal(t, h11 == h12, true)
	assert.Equal(t, h11 == h1, true)
	assert.Equal(t, h11 == h2, true)

	vars1.Set("k1", uint16(42))
	vars1.Set("k2", "42")

	vars2.Set("k2", "42")
	vars2.Set("k1", uint16(42))
	h11 = hashing.GetHashValue(vars1)
	h12 = hashing.GetHashValue(vars2)
	assert.Equal(t, h11 == h12, true)
}

func TestDetereminism(t *testing.T) {
	vars1 := New(nil)
	h1 := hashing.GetHashValue(vars1)

	vars2 := New(vars1)
	h2 := hashing.GetHashValue(vars2)

	assert.Equal(t, h1 == h2, true)

	vars1.Set("k1", "kuku")
	vars1.Set("k2", uint16(42))
	vars1.Set("k3", "kuku")
	vars1.Set("k4", uint16(2))

	vars2.Set("k4", uint16(2))
	vars2.Set("k3", "kuku")
	vars2.Set("k2", uint16(42))
	vars2.Set("k1", "kuku")

	h11 := hashing.GetHashValue(vars1)
	h12 := hashing.GetHashValue(vars2)
	assert.Equal(t, h11 == h12, true)

	vars1.Set("k3", nil)
	vars2.Set("k3", nil)
	assert.Equal(t, h11 == h12, true)
}
