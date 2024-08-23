package wbf_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/util/wbf"
	"github.com/stretchr/testify/require"
)

type BasicStruct struct {
	A int
	B string
	C int `wbf:"-"`
}

type IntWithLessBytes struct {
	A int `wbf:"bytes=2"`
}

type IntWithMoreBytes struct {
	A int16 `wbf:"bytes=4"`
}

type IntPtr struct {
	A *int
}

type IntOptional struct {
	A *int `wbf:"optional"`
}

type NestedStruct struct {
	A int
	B BasicStruct
}

type OptionalNestedStruct struct {
	A int
	B *BasicStruct `wbf:"optional"`
}

type EmbeddedStruct struct {
	BasicStruct
	C int
}

type OptionalEmbeddedStruct struct {
	*BasicStruct `wbf:"optional"`
	C            int
}

type WithSlice struct {
	A []int
}

type WithShortSlice struct {
	A []int `wbf:"len_bytes=2"`
}

type WithOptionalSlice struct {
	A *[]int `wbf:"optional"`
}

type WithArray struct {
	A [3]int
}

type WithBigIntPtr struct {
	A *big.Int
}

type WithBigIntVal struct {
	A big.Int
}

type WithTime struct {
	A time.Time
}

type WithCustomCodec struct {
}

func (w WithCustomCodec) WBFEncode(e *wbf.Encoder) error {
	e.Write([]byte{1, 2, 3})
	return nil
}

func (w *WithCustomCodec) WBFDecode(d *wbf.Decoder) error {
	b, err := d.Read(3)
	if err != nil {
		return err
	}

	if b[0] != 1 || b[1] != 2 || b[2] != 3 {
		return fmt.Errorf("invalid value: %v", b)
	}

	return nil
}

type WithNestedCustomCodec struct {
	A int `wbf:"bytes=1"`
	B WithCustomCodec
}

type ShortInt int

func (v ShortInt) WBFOptions() wbf.TypeOptions {
	return wbf.TypeOptions{Bytes: wbf.Value2Bytes}
}

type WithWBFOpts struct {
	A ShortInt
}

type WithWBFOptsOverride struct {
	A ShortInt `wbf:"bytes=1"`
}

type WitUnexported struct {
	A int
	b int
	c int `wbf:""`
	D int `wbf:"-"`
}

func TestEncoder(t *testing.T) {
	var err error

	r := must2(wbf.Encode(BasicStruct{A: 42, B: "aaa", C: 5}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97}, r)

	r = must2(wbf.Encode(IntWithLessBytes{A: 42}))
	require.Equal(t, []byte{42, 0}, r)

	intV := 42
	r = must2(wbf.Encode(IntPtr{A: &intV}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0}, r)

	_, err = wbf.Encode(IntPtr{A: nil})
	require.Error(t, err)

	r = must2(wbf.Encode(IntOptional{A: &intV}))
	require.Equal(t, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(IntOptional{A: nil}))
	require.Equal(t, []byte{0}, r)

	r = must2(wbf.Encode(NestedStruct{A: 42, B: BasicStruct{A: 43, B: "aaa"}}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97}, r)

	r = must2(wbf.Encode(OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97}, r)

	r = must2(wbf.Encode(OptionalNestedStruct{A: 42, B: nil}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(EmbeddedStruct{BasicStruct: BasicStruct{A: 42, B: "aaa"}, C: 43}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(OptionalEmbeddedStruct{BasicStruct: &BasicStruct{A: 42, B: "aaa"}, C: 43}))
	require.Equal(t, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(OptionalEmbeddedStruct{BasicStruct: nil, C: 43}))
	require.Equal(t, []byte{0, 43, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(WithSlice{A: []int{42, 43}}))
	require.Equal(t, []byte{2, 42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(WithSlice{A: []int{}}))
	require.Equal(t, []byte{0}, r)

	r = must2(wbf.Encode(WithSlice{A: nil}))
	require.Equal(t, []byte{0}, r)

	r = must2(wbf.Encode(WithShortSlice{A: []int{42, 43}}))
	require.Equal(t, []byte{2, 42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(WithOptionalSlice{A: &[]int{42, 43}}))
	require.Equal(t, []byte{1, 2, 42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(WithArray{A: [3]int{42, 43, 44}}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0, 44, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(WithBigIntPtr{A: big.NewInt(42)}))
	require.Equal(t, []byte{1, 42}, r)

	r = must2(wbf.Encode(WithBigIntVal{A: *big.NewInt(42)}))
	require.Equal(t, []byte{1, 42}, r)

	r = must2(wbf.Encode(WithTime{A: time.Unix(0, 42)}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0}, r)

	r = must2(wbf.Encode(WithCustomCodec{}))
	require.Equal(t, []byte{1, 2, 3}, r)

	r = must2(wbf.Encode(WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}))
	require.Equal(t, []byte{43, 1, 2, 3}, r)

	r = must2(wbf.Encode(WithWBFOpts{A: 42}))
	require.Equal(t, []byte{42, 0}, r)

	r = must2(wbf.Encode(WithWBFOptsOverride{A: 42}))
	require.Equal(t, []byte{42}, r)

	r = must2(wbf.Encode(WitUnexported{A: 42, b: 43, c: 44, D: 45}))
	require.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0, 44, 0, 0, 0, 0, 0, 0, 0}, r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[Res any](v Res, err error) Res {
	if err != nil {
		panic(err)
	}
	return v
}
