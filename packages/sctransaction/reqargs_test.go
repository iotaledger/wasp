package sctransaction

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRequestArguments1(t *testing.T) {
	r := NewRequestArguments()
	r.Add("arg1", []byte("data1"))
	r.Add("arg2", []byte("data2"))
	r.Add("arg3", []byte("data3"))
	r.AddAsBlobHash("arg4", []byte("data4"))

	require.Len(t, r, 4)
	require.EqualValues(t, r["-arg1"], "data1")
	require.EqualValues(t, r["-arg2"], "data2")
	require.EqualValues(t, r["-arg3"], "data3")
	h := hashing.HashStrings("data4")
	require.EqualValues(t, r["*arg4"], h[:])

	var buf bytes.Buffer
	err := r.Write(&buf)
	require.NoError(t, err)

	rdr := bytes.NewReader(buf.Bytes())
	back := NewRequestArguments()
	err = back.Read(rdr)
	require.NoError(t, err)
}

func TestRequestArguments2(t *testing.T) {
	r := NewRequestArguments()
	r.Add("arg1", []byte("data1"))
	r.Add("arg2", []byte("data2"))
	r.Add("arg3", []byte("data3"))
	r.AddAsBlobHash("arg4", []byte("data4"))

	h := hashing.HashStrings("data4")

	require.Len(t, r, 4)
	require.EqualValues(t, r["-arg1"], "data1")
	require.EqualValues(t, r["-arg2"], "data2")
	require.EqualValues(t, r["-arg3"], "data3")
	require.EqualValues(t, r["*arg4"], h[:])

	var buf bytes.Buffer
	err := r.Write(&buf)
	require.NoError(t, err)

	rdr := bytes.NewReader(buf.Bytes())
	back := NewRequestArguments()
	err = back.Read(rdr)
	require.NoError(t, err)

	require.Len(t, back, 4)
	require.EqualValues(t, back["-arg1"], "data1")
	require.EqualValues(t, back["-arg2"], "data2")
	require.EqualValues(t, back["-arg3"], "data3")
	require.EqualValues(t, back["*arg4"], h[:])
}

func TestRequestArguments3(t *testing.T) {
	r := NewRequestArguments()
	r.Add("arg1", []byte("data1"))
	r.Add("arg2", []byte("data2"))
	r.Add("arg3", []byte("data3"))

	require.Len(t, r, 3)
	require.EqualValues(t, r["-arg1"], "data1")
	require.EqualValues(t, r["-arg2"], "data2")
	require.EqualValues(t, r["-arg3"], "data3")

	log := testutil.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)
	reg := registry.NewRegistry(nil, log, db)

	d, ok, err := r.DecodeRequestArguments(reg)
	require.NoError(t, err)
	require.True(t, ok)

	dec := kvdecoder.New(d)
	var s1, s2, s3 string
	require.NotPanics(t, func() {
		s1 = dec.MustGetString("arg1")
		s2 = dec.MustGetString("arg2")
		s3 = dec.MustGetString("arg3")
	})
	require.Len(t, d, 3)
	require.EqualValues(t, "data1", s1)
	require.EqualValues(t, "data2", s2)
	require.EqualValues(t, "data3", s3)
}

func TestRequestArguments4(t *testing.T) {
	r := NewRequestArguments()
	r.Add("arg1", []byte("data1"))
	r.Add("arg2", []byte("data2"))
	r.Add("arg3", []byte("data3"))
	data := []byte("data4")
	r.AddAsBlobHash("arg4", data)
	h := hashing.HashData(data)

	require.Len(t, r, 4)
	require.EqualValues(t, r["-arg1"], "data1")
	require.EqualValues(t, r["-arg2"], "data2")
	require.EqualValues(t, r["-arg3"], "data3")
	require.EqualValues(t, r["*arg4"], h[:])

	log := testutil.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)
	reg := registry.NewRegistry(nil, log, db)

	_, ok, err := r.DecodeRequestArguments(reg)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRequestArguments5(t *testing.T) {
	r := NewRequestArguments()
	r.Add("arg1", []byte("data1"))
	r.Add("arg2", []byte("data2"))
	r.Add("arg3", []byte("data3"))
	data := []byte("data4-data4-data4-data4-data4-data4-data4")
	r.AddAsBlobHash("arg4", data)
	h := hashing.HashData(data)

	require.Len(t, r, 4)
	require.EqualValues(t, r["-arg1"], "data1")
	require.EqualValues(t, r["-arg2"], "data2")
	require.EqualValues(t, r["-arg3"], "data3")
	require.EqualValues(t, r["*arg4"], h[:])

	log := testutil.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)
	reg := registry.NewRegistry(nil, log, db)

	hback, err := reg.PutBlob(data)
	require.NoError(t, err)
	require.EqualValues(t, h, hback)

	back, ok, err := r.DecodeRequestArguments(reg)
	require.NoError(t, err)
	require.True(t, ok)

	require.Len(t, back, 4)
	require.EqualValues(t, back["arg1"], "data1")
	require.EqualValues(t, back["arg2"], "data2")
	require.EqualValues(t, back["arg3"], "data3")
	require.EqualValues(t, back["arg4"], data)
}
