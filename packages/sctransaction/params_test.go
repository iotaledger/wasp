package sctransaction

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicVariables(t *testing.T) {
	p := make(Params)

	_, ok := p.GetInt64("v1")
	assert.False(t, ok)

	assert.Equal(t, 0, len(p))

	p.SetInt64("v1", int64(42))
	vint, ok := p.GetInt64("v1")
	assert.True(t, ok)
	assert.Equal(t, int64(42), vint)

	assert.Equal(t, 1, len(p))

	_, ok = p.GetString("v2")
	assert.False(t, ok)

	p.SetString("v2", "a string")
	vstr, ok := p.GetString("v2")
	assert.True(t, ok)
	assert.Equal(t, "a string", vstr)

	assert.Equal(t, 2, len(p))
}

func TestMarshaling(t *testing.T) {
	p := make(Params)

	p.SetString("k1", "v1")
	p.SetInt64("k2", 42)
	p.SetInt64("k3", -42)

	var buf bytes.Buffer
	err := p.Write(&buf)
	assert.NoError(t, err)

	p2 := make(Params)
	err = p2.Read(bytes.NewReader(buf.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, 3, len(p2))

	s, ok := p2.GetString("k1")
	assert.Equal(t, true, ok)
	assert.Equal(t, "v1", s)

	n, ok := p2.GetInt64("k2")
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(42), n)

	n, ok = p2.GetInt64("k3")
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(-42), n)
}

func TestDetereminism(t *testing.T) {
	p1 := make(Params)
	p1.SetString("k1", "kuku")
	p1.SetInt64("k2", int64(42))
	p1.SetString("k3", "kuku")
	p1.SetInt64("k4", int64(2))

	p2 := make(Params)
	p2.SetInt64("k4", int64(2))
	p2.SetString("k3", "kuku")
	p2.SetInt64("k2", int64(42))
	p2.SetString("k1", "kuku")

	var buf1 bytes.Buffer
	err := p1.Write(&buf1)
	assert.NoError(t, err)

	var buf2 bytes.Buffer
	err = p2.Write(&buf2)
	assert.NoError(t, err)

	assert.Equal(t, buf1.Bytes(), buf2.Bytes())
}
