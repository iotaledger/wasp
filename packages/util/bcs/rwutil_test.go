package bcs_test

import (
	"io"
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

type StructWithRwUtilSupport struct {
	A int
	B int `bcs:"bytes=2"`
	C NestedStructWithRwUtilSupport
}

func (s *StructWithRwUtilSupport) Write(dest io.Writer) error {
	w := rwutil.NewWriter(dest)

	w.WriteInt64(int64(s.A))
	w.WriteInt16(int16(s.B))
	w.Write(&s.C)

	return nil
}

func (s *StructWithRwUtilSupport) Read(src io.Reader) error {
	r := rwutil.NewReader(src)

	s.A = int(r.ReadInt64())
	s.B = int(r.ReadInt16())
	r.Read(&s.C)

	return r.Err
}

type NestedStructWithRwUtilSupport struct {
	C int
	D []string `bcs:"len_bytes=2"`
}

func (s *NestedStructWithRwUtilSupport) Write(dest io.Writer) error {
	w := rwutil.NewWriter(dest)

	w.WriteInt64(int64(s.C))
	w.WriteSize16(len(s.D))
	for _, v := range s.D {
		w.WriteString(v)
	}

	return nil
}

func (s *NestedStructWithRwUtilSupport) Read(src io.Reader) error {
	r := rwutil.NewReader(src)

	s.C = int(r.ReadInt64())
	size := r.ReadSize16()
	s.D = make([]string, size)
	for i := range s.D {
		s.D[i] = r.ReadString()
	}

	return r.Err
}

func TestCompatibilityWithRwUtil(t *testing.T) {
	v := StructWithRwUtilSupport{
		A: 42,
		B: 43,
		C: NestedStructWithRwUtilSupport{
			C: 44,
			D: []string{"aaa", "bbb"},
		},
	}

	vEnc := lo.Must1(bcs.Marshal(v))
	vDec := lo.Must1(bcs.Unmarshal[StructWithRwUtilSupport](vEnc))
	require.Equal(t, v, vDec)

	written := rwutil.WriteToBytes(&v)
	require.Equal(t, written, vEnc)

	var read StructWithRwUtilSupport
	rwutil.ReadFromBytes(written, &read)
	require.Equal(t, v, read)

	var readFromEnc StructWithRwUtilSupport
	rwutil.ReadFromBytes(vEnc, &readFromEnc)
	require.Equal(t, v, readFromEnc)
}
