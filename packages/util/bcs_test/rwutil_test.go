package bcs_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

type BasicWithRwUtilCodec string

func (v BasicWithRwUtilCodec) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN([]byte{1, 2, 3})
	ww.WriteString(string(v))

	return ww.Err
}

func (v *BasicWithRwUtilCodec) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)

	b := make([]byte, 3)
	rr.ReadN(b)

	if b[0] != 1 || b[1] != 2 || b[2] != 3 {
		return fmt.Errorf("invalid value: %v", b)
	}

	s := rr.ReadString()

	*v = BasicWithRwUtilCodec(s)

	return rr.Err
}

func TestBasicTypesCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, BasicWithRwUtilCodec("aaa"), []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61})
}

type StructWithRwUtilSupport struct {
	A int
	B int `bcs:"type=i16"`
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

	vEnc := lo.Must1(bcs.Marshal(&v))
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

type WithRwUtilCodec struct{}

func (v WithRwUtilCodec) Write(w io.Writer) error {
	w.Write([]byte{1, 2, 3})
	return nil
}

func (v *WithRwUtilCodec) Read(r io.Reader) error {
	b := make([]byte, 3)
	if _, err := r.Read(b); err != nil {
		return err
	}

	if b[0] != 1 || b[1] != 2 || b[2] != 3 {
		return fmt.Errorf("invalid value: %v", b)
	}

	return nil
}

type WithNestedRwUtilCodec struct {
	A int `bcs:"type=i8"`
	B WithRwUtilCodec
}

func TestStructCustomCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, WithRwUtilCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithNestedRwUtilCodec{A: 43, B: WithRwUtilCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
}
