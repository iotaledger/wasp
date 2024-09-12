package bcs_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

type BasicStruct struct {
	A int64
	B string
	C int64 `bcs:"-"`
}

type IntWithLessBytes struct {
	A int64 `bcs:"bytes=2"`
}

type IntWithMoreBytes struct {
	A int16 `bcs:"bytes=4"`
}

type IntPtr struct {
	A *int64
}

type IntMultiPtr struct {
	A **int64
}

type IntOptional struct {
	A *int64 `bcs:"optional"`
}

type IntOptionalPtr struct {
	A **int64 `bcs:"optional"`
}

type NestedStruct struct {
	A int64
	B BasicStruct
}

type OptionalNestedStruct struct {
	A int64
	B *BasicStruct `bcs:"optional"`
}

type EmbeddedStruct struct {
	BasicStruct
	C int64
}

type OptionalEmbeddedStruct struct {
	*BasicStruct `bcs:"optional"`
	C            int64
}

type WithByteArr struct {
	A string
	B int64 `bcs:"bytes=4,bytearr"`
	C string
}

type WithSlice struct {
	A []int32
}

type WithShortSlice struct {
	A []int32 `bcs:"len_bytes=2"`
}

type WithOptionalSlice struct {
	A *[]int32 `bcs:"optional"`
}

type WithArray struct {
	A [3]int16
}

type WithByteArrElem struct {
	ByteArrayVal []BasicStruct `bcs_elem:"bytearr"`
}

type WithByteArrByte struct {
	A []byte `bcs_elem:"bytearr"`
}

type WithByteArrInt struct {
	ByteArrVal []int32 `bcs_elem:"bytearr"`
}

type WithMap struct {
	A map[int16]bool
}

type WithOptionalMap struct {
	A map[int16]bool `bcs:"optional"`
}

type WithOptionalMapPtr struct {
	A *map[int16]bool `bcs:"optional"`
}

type WithShortMap struct {
	A map[int16]bool `bcs:"len_bytes=2"`
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

func (w WithCustomCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Write([]byte{1, 2, 3})
	return nil
}

func (w *WithCustomCodec) UnmarshalBCS(d *bcs.Decoder) error {
	b := make([]byte, 3)
	if _, err := d.Read(b); err != nil {
		return err
	}

	if b[0] != 1 || b[1] != 2 || b[2] != 3 {
		return fmt.Errorf("invalid value: %v", b)
	}

	return nil
}

type WithInit struct {
	A int
}

func (w *WithInit) BCSInit() error {
	w.A = w.A + 10
	return nil
}

type WithCustomAndInit struct {
	WithCustomCodec
	A int
}

func (w *WithCustomAndInit) BCSInit() error {
	w.A = 100
	return nil
}

type WithNestedCustomCodec struct {
	A int `bcs:"bytes=1"`
	B WithCustomCodec
}

type WithNestedCustomPtrCodec struct {
	A int `bcs:"bytes=1"`
	B BasicWithCustomPtrCodec
}

type WithNestedPtrCustomPtrCodec struct {
	A int `bcs:"bytes=1"`
	B *BasicWithCustomPtrCodec
}

type ShortInt int64

func (v ShortInt) BCSOptions() bcs.TypeOptions {
	return bcs.TypeOptions{Bytes: bcs.Value2Bytes}
}

type WithBCSOpts struct {
	A ShortInt
}

type WithBCSOptsOverride struct {
	A ShortInt `bcs:"bytes=1"`
}

type WitUnexported struct {
	A int
	b int
	c int `bcs:""`
	D int `bcs:"-"`
}

func TestStructCodec(t *testing.T) {
	bcs.TestCodecAndBytesVsRef(t, BasicStruct{A: 42, B: "aaa"}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytes(t, IntWithLessBytes{A: 42}, []byte{42, 0})
	bcs.TestCodecAndBytes(t, IntWithMoreBytes{A: 42}, []byte{42, 0, 0, 0})
	vI := int64(42)
	pVI := &vI
	bcs.TestCodecAndBytesVsRef(t, IntPtr{A: &vI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestEncodeErr(t, IntPtr{A: nil})
	bcs.TestCodecAndBytesVsRef(t, IntMultiPtr{A: &pVI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestEncodeErr(t, IntMultiPtr{A: nil})
	bcs.TestCodecAndBytesVsRef(t, IntOptional{A: &vI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, IntOptionalPtr{A: &pVI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, NestedStruct{A: 42, B: BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, &OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, OptionalNestedStruct{A: 42, B: nil}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, EmbeddedStruct{BasicStruct: BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, OptionalEmbeddedStruct{BasicStruct: &BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, OptionalEmbeddedStruct{BasicStruct: nil, C: 43}, []byte{0, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytes(t, WithByteArr{A: "aaa", B: 42, C: "ccc"}, []byte{0x3, 0x61, 0x61, 0x61, 0x4, 0x2a, 0x0, 0x0, 0x0, 0x3, 0x63, 0x63, 0x63})
	bcs.TestCodecAndBytesVsRef(t, WithSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, WithSlice{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytesVsRef(t, WithOptionalSlice{A: &[]int32{42, 43}}, []byte{1, 0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, WithOptionalSlice{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, WithShortSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, WithArray{A: [3]int16{42, 43, 44}}, []byte{0x2a, 0x0, 0x2b, 0x0, 0x2c, 0x0})
	bcs.TestCodecAndBytes(t, WithByteArrElem{ByteArrayVal: []BasicStruct{{A: 42, B: "aaa"}, {A: 43, B: "bbb"}}}, []byte{0x2, 0xc, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x61, 0x61, 0x61, 0xc, 0x2b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x62, 0x62, 0x62})
	bcs.TestCodecAndBytes(t, WithByteArrInt{ByteArrVal: []int32{1, 2, 3}}, []byte{0x3, 0x4, 0x1, 0x0, 0x0, 0x0, 0x4, 0x2, 0x0, 0x0, 0x0, 0x4, 0x3, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithByteArrByte{A: []byte{1, 2, 3}}, []byte{0x3, 0x1, 0x1, 0x1, 0x2, 0x1, 0x3})
	bcs.TestCodecAndBytes(t, WithMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	bcs.TestCodecAndBytes(t, WithMap{A: map[int16]bool{}}, []byte{0x0})
	bcs.TestEncodeErr(t, WithMap{A: nil})
	bcs.TestCodecAndBytes(t, WithOptionalMap{A: map[int16]bool{}}, []byte{0x1, 0x0})
	bcs.TestCodecAndBytes(t, WithOptionalMap{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, WithOptionalMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	bcs.TestCodecAndBytes(t, WithOptionalMapPtr{A: &map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	var m map[int16]bool
	bcs.TestEncodeErr(t, WithOptionalMapPtr{A: &m})
	bcs.TestCodecAndBytes(t, WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, &WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, &WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithNestedCustomPtrCodec{A: 43, B: BasicWithCustomPtrCodec("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, &WithNestedCustomPtrCodec{A: 43, B: BasicWithCustomPtrCodec("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, WithNestedPtrCustomPtrCodec{A: 43, B: lo.ToPtr[BasicWithCustomPtrCodec]("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, &WithNestedPtrCustomPtrCodec{A: 43, B: lo.ToPtr[BasicWithCustomPtrCodec]("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	bcs.TestCodecAndBytes(t, &WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	bcs.TestCodecAndBytes(t, WithBCSOptsOverride{A: 42}, []byte{0x2A})
	bcs.TestCodecAndBytes(t, &WithBCSOptsOverride{A: 42}, []byte{0x2A})
	bcs.TestCodecAndBytes(t, WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, &WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithBigIntVal{A: *big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithTime{A: time.Unix(12345, 6789)}, []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})
}

func TestStructWithInit(t *testing.T) {
	vInit := WithInit{A: 71}
	vEnc := bcs.MustMarshal(&vInit)
	require.Equal(t, []byte{0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, vEnc)
	vInitDec := bcs.MustUnmarshal[WithInit](vEnc)
	require.NotEqual(t, vInit, vInitDec)
	vInit.A += 10
	require.Equal(t, vInit, vInitDec)

	vCustomAndInit := WithCustomAndInit{A: 71}
	vEnc = bcs.MustMarshal(&vCustomAndInit)
	require.Equal(t, []byte{0x1, 0x2, 0x3}, vEnc)
	vCustomAndInitDec := bcs.MustUnmarshal[WithCustomAndInit](vEnc)
	require.NotEqual(t, vCustomAndInit, vCustomAndInitDec)
	vCustomAndInit.A = 100
	require.Equal(t, vCustomAndInit, vCustomAndInitDec)
}

func TestUnexportedFieldsCodec(t *testing.T) {
	v := WitUnexported{A: 42, b: 43, c: 44, D: 45}
	vEnc := lo.Must1(bcs.Marshal(&v))
	require.Equal(t, []byte{42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 44, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, vEnc)
	vDec := lo.Must1(bcs.Unmarshal[WitUnexported](vEnc))
	require.NotEqual(t, v, vDec)
	require.Equal(t, 0, vDec.b)
	require.Equal(t, 0, vDec.D)
	vDec.b = 43
	vDec.D = 45
	require.Equal(t, v, vDec)
}

type StructWithManualCodec struct {
	A int
	B *string `bcs:"optional"`
	C BasicStructEnum
}

func (s *StructWithManualCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Encode(s.A)
	e.EncodeOptional(s.B)

	switch {
	case s.C.A != nil:
		e.EncodeEnumVariantIdx(0)
		e.Encode(*s.C.A)
	case s.C.B != nil:
		e.EncodeEnumVariantIdx(1)
		e.Encode(*s.C.B)
	case s.C.C != nil:
		e.EncodeEnumVariantIdx(2)
		e.Encode(*s.C.C)
	default:
		panic("enum variant not set")
	}

	// Just to mark that this is a manual codec.
	return e.Encode(byte(1))
}

func (s *StructWithManualCodec) UnmarshalBCS(d *bcs.Decoder) error {
	d.Decode(&s.A)
	d.DecodeOptional(&s.B)

	variantIdx, _ := d.DecodeEnumVariantIdx()

	switch variantIdx {
	case 0:
		s.C.A = new(int32)
		d.Decode(s.C.A)
	case 1:
		s.C.B = new(string)
		d.Decode(s.C.B)
	case 2:
		s.C.C = new([]byte)
		d.Decode(s.C.C)
	default:
		panic("invalid enum variant")
	}

	// Just to ensure that this is a manual codec.
	var b byte
	d.Decode(&b)
	if b != 1 {
		panic("invalid manual codec marker")
	}

	return d.Err()
}

type StructWithAutoCodec StructWithManualCodec

func TestHighLevelCodecFuncs(t *testing.T) {
	v := StructWithManualCodec{
		A: 42,
		B: lo.ToPtr("aaa"),
		C: BasicStructEnum{B: lo.ToPtr("bbb")},
	}
	vEnc := bcs.TestCodec(t, v)
	require.Equal(t, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x3, 0x61, 0x61, 0x61, 0x1, 0x3, 0x62, 0x62, 0x62, 0x1}, vEnc)

	vAuto := StructWithAutoCodec(v)
	vAutoEnc := bcs.TestCodec(t, vAuto)
	vEncWithoutMarker := vEnc[:len(vEnc)-1]
	require.Equal(t, vEncWithoutMarker, vAutoEnc)
}
