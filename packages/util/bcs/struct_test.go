package bcs_test

import (
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

type BasicStruct struct {
	A int64
	B string
	C int64 `bcs:"-"`
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

type MapOptional struct {
	A map[byte]byte `bcs:"optional"`
}

type MapPtrOptional struct {
	A *map[byte]byte `bcs:"optional"`
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

func TestStructCodec(t *testing.T) {
	bcs.TestCodecAndBytesVsRef(t, BasicStruct{A: 42, B: "aaa"}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, NestedStruct{A: 42, B: BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, &OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesVsRef(t, OptionalNestedStruct{A: 42, B: nil}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, EmbeddedStruct{BasicStruct: BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, OptionalEmbeddedStruct{BasicStruct: &BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, OptionalEmbeddedStruct{BasicStruct: nil, C: 43}, []byte{0, 43, 0, 0, 0, 0, 0, 0, 0})
}

type WithUnexported struct {
	A int
	b int
	c int `bcs:"export"`
	D int `bcs:"-"`
}

type WithUnexportedWithoutMark struct {
	A int
	b int `bcs:"type=u16"`
}

type WithUnexportedWrongTag struct {
	A int `bcs:"export"`
}

func TestUnexportedFieldsCodec(t *testing.T) {
	v := WithUnexported{A: 42, b: 43, c: 44, D: 45}
	vEnc := lo.Must1(bcs.Marshal(&v))
	require.Equal(t, []byte{42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 44, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, vEnc)
	vDec := lo.Must1(bcs.Unmarshal[WithUnexported](vEnc))
	require.NotEqual(t, v, vDec)
	require.Equal(t, 0, vDec.b)
	require.Equal(t, 0, vDec.D)
	vDec.b = 43
	vDec.D = 45
	require.Equal(t, v, vDec)

	bcs.TestEncodeErr(t, WithUnexportedWithoutMark{A: 42, b: 43})
	bcs.TestEncodeErr(t, WithUnexportedWrongTag{A: 42})
}

type Embedded struct {
	A int
}

type WithEmbedded struct {
	Embedded
	B int
}

type embeddedPrivate struct {
	A int
}

type WithEmbeddedPrivate struct {
	embeddedPrivate
	B int
}

type WithEmbeddedPrivateWithTag struct {
	embeddedPrivate `bcs:"export"`
	B               int
}

func TestEmbedded(t *testing.T) {
	bcs.TestCodec(t, WithEmbedded{Embedded: Embedded{A: 5}, B: 10})
	bcs.TestCodecAsymmetric(t, WithEmbeddedPrivate{embeddedPrivate: embeddedPrivate{A: 5}, B: 10})
	bcs.TestCodec(t, WithEmbeddedPrivateWithTag{embeddedPrivate: embeddedPrivate{A: 5}, B: 10})
}

func TestStructWithPtr(t *testing.T) {
	vI := int64(42)
	bcs.TestCodecAndBytesVsRef(t, IntPtr{A: &vI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestEncodeErr(t, IntPtr{A: nil})
	pVI := &vI
	bcs.TestCodecAndBytesVsRef(t, IntMultiPtr{A: &pVI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestEncodeErr(t, IntMultiPtr{A: nil})
	bcs.TestCodecAndBytesVsRef(t, IntOptional{A: &vI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesVsRef(t, IntOptionalPtr{A: &pVI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
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

func TestStructWithContainers(t *testing.T) {
	bcs.TestCodecAndBytesVsRef(t, WithSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, WithSlice{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytesVsRef(t, WithOptionalSlice{A: &[]int32{42, 43}}, []byte{1, 0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, WithOptionalSlice{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, WithShortSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, WithArray{A: [3]int16{42, 43, 44}}, []byte{0x2a, 0x0, 0x2b, 0x0, 0x2c, 0x0})
	bcs.TestCodecAndBytes(t, WithMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	bcs.TestCodecAndBytes(t, WithMap{A: map[int16]bool{}}, []byte{0x0})
	bcs.TestEncodeErr(t, WithMap{A: nil})
	bcs.TestCodecAndBytes(t, WithOptionalMap{A: map[int16]bool{}}, []byte{0x1, 0x0})
	bcs.TestCodecAndBytes(t, WithOptionalMap{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, WithOptionalMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	bcs.TestCodecAndBytes(t, WithOptionalMapPtr{A: &map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	var m map[int16]bool
	bcs.TestEncodeErr(t, WithOptionalMapPtr{A: &m})
	bcs.TestCodecAndBytes(t, MapOptional{A: map[byte]byte{3: 4}}, []byte{0x1, 0x1, 0x3, 0x4})
	bcs.TestCodecAndBytes(t, MapOptional{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, MapPtrOptional{A: &map[byte]byte{3: 4}}, []byte{0x1, 0x1, 0x3, 0x4})
	bcs.TestCodecAndBytes(t, MapPtrOptional{A: nil}, []byte{0x0})
}

type WithCustomCodec struct{}

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

type WithNestedCustomCodec struct {
	A int `bcs:"type=i8"`
	B WithCustomCodec
}

type WithNestedRwUtilCodec struct {
	A int `bcs:"type=i8"`
	B WithRwUtilCodec
}

type WithNestedCustomPtrCodec struct {
	A int `bcs:"type=i8"`
	B BasicWithCustomPtrCodec
}

type WithNestedPtrCustomPtrCodec struct {
	A int `bcs:"type=i8"`
	B *BasicWithCustomPtrCodec
}

func TestStructCustomCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, &WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithRwUtilCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, &WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithNestedRwUtilCodec{A: 43, B: WithRwUtilCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, WithNestedCustomPtrCodec{A: 43, B: BasicWithCustomPtrCodec("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, &WithNestedCustomPtrCodec{A: 43, B: BasicWithCustomPtrCodec("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, WithNestedPtrCustomPtrCodec{A: 43, B: lo.ToPtr[BasicWithCustomPtrCodec]("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, &WithNestedPtrCustomPtrCodec{A: 43, B: lo.ToPtr[BasicWithCustomPtrCodec]("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
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

func TestStructWithThirdPartyTypes(t *testing.T) {
	bcs.TestCodecAndBytes(t, WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, &WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithBigIntVal{A: *big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithTime{A: time.Unix(12345, 6789)}, []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})
}

type WithByteArr struct {
	A string
	B int64 `bcs:"type=i32,bytearr"`
	C string
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

func TestStructWithByteArr(t *testing.T) {
	bcs.TestCodecAndBytes(t, WithByteArr{A: "aaa", B: 42, C: "ccc"}, []byte{0x3, 0x61, 0x61, 0x61, 0x4, 0x2a, 0x0, 0x0, 0x0, 0x3, 0x63, 0x63, 0x63})
	bcs.TestCodecAndBytes(t, WithByteArrElem{ByteArrayVal: []BasicStruct{{A: 42, B: "aaa"}, {A: 43, B: "bbb"}}}, []byte{0x2, 0xc, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x61, 0x61, 0x61, 0xc, 0x2b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x62, 0x62, 0x62})
	bcs.TestCodecAndBytes(t, WithByteArrInt{ByteArrVal: []int32{1, 2, 3}}, []byte{0x3, 0x4, 0x1, 0x0, 0x0, 0x0, 0x4, 0x2, 0x0, 0x0, 0x0, 0x4, 0x3, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithByteArrByte{A: []byte{1, 2, 3}}, []byte{0x3, 0x1, 0x1, 0x1, 0x2, 0x1, 0x3})
}

type InfOptional struct {
	A InfEnum1 `bcs:"optional,not_enum"`
}

type InfPtrOptional struct {
	A *InfEnum1 `bcs:"optional,not_enum"`
}

func TestStructWithNotEnum(t *testing.T) {
	bcs.TestCodecAndBytes(t, InfOptional{A: nil}, []byte{0x0})
	require.Equal(t, []byte{0x1, 0x3}, bcs.MustMarshal(&InfOptional{A: byte(3)}))
	bcs.TestDecodeErr[InfOptional](t, []byte{0x1, 0x3})
}

type ShortInt int64

func (v ShortInt) BCSOptions() bcs.TypeOptions {
	return bcs.TypeOptions{UnderlayingType: reflect.Int16}
}

type WithBCSOpts struct {
	A ShortInt
}

type WithBCSOptsOverride struct {
	A ShortInt `bcs:"type=i8"`
}

func TestSTructWithBCSOpts(t *testing.T) {
	bcs.TestCodecAndBytes(t, WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	bcs.TestCodecAndBytes(t, &WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	bcs.TestCodecAndBytes(t, WithBCSOptsOverride{A: 42}, []byte{0x2A})
	bcs.TestCodecAndBytes(t, &WithBCSOptsOverride{A: 42}, []byte{0x2A})
}

type CompactInt struct {
	A int64 `bcs:"compact"`
}

type CompactUint struct {
	A uint64 `bcs:"compact"`
}

type IntWithLessBytes struct {
	A int64 `bcs:"type=i16"`
}

type IntWithMoreBytes struct {
	A int16 `bcs:"type=i32"`
}

type IntWithLessBytesAndOtherUnsigned struct {
	A uint64 `bcs:"type=i16"`
	B int64  `bcs:"type=u16"`
}

type IntWithMoreBytesAndOtherUnsigned struct {
	A uint16 `bcs:"type=i32"`
	B int16  `bcs:"type=u32"`
}

func TestIntTypeConversion(t *testing.T) {
	bcs.TestCodecAndBytes(t, CompactInt{A: 1}, []byte{0x1})
	bcs.TestCodecAndBytes(t, CompactInt{A: 70000}, []byte{0xf0, 0xa2, 0x4})
	bcs.TestCodecAndBytes(t, CompactInt{A: math.MaxInt64}, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f})
	bcs.TestCodecAndBytes(t, CompactInt{A: math.MinInt64}, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x1})

	bcs.TestCodecAndBytes(t, CompactUint{A: 1}, []byte{0x1})
	bcs.TestCodecAndBytes(t, CompactUint{A: 70000}, []byte{0xf0, 0xa2, 0x4})
	bcs.TestCodecAndBytes(t, CompactUint{A: math.MaxUint64}, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x1})

	bcs.TestCodecAndBytes(t, IntWithLessBytes{A: 42}, []byte{42, 0})
	bcs.TestCodecAndBytes(t, IntWithMoreBytes{A: 42}, []byte{42, 0, 0, 0})
	bcs.TestCodecAndBytes(t, IntWithLessBytes{A: math.MinInt16}, []byte{0x0, 0x80})
	bcs.TestCodecAndBytes(t, IntWithLessBytes{A: math.MaxInt16}, []byte{0xff, 0x7f})
	bcs.TestCodecAndBytes(t, IntWithMoreBytes{A: math.MinInt16}, []byte{0x0, 0x80, 0xff, 0xff})
	bcs.TestCodecAndBytes(t, IntWithMoreBytes{A: math.MaxInt16}, []byte{0xff, 0x7f, 0x0, 0x0})

	bcs.TestEncodeErr(t, IntWithLessBytes{A: math.MaxUint16})
	bcs.TestEncodeErr(t, IntWithLessBytes{A: math.MinInt32})
	bcs.TestDecodeErr[IntWithMoreBytes](t, uint32(math.MaxInt32))

	bcs.TestCodecAndBytes(t, IntWithLessBytesAndOtherUnsigned{A: 42}, []byte{42, 0, 0, 0})
	bcs.TestCodecAndBytes(t, IntWithLessBytesAndOtherUnsigned{B: 42}, []byte{0, 0, 42, 0})
	bcs.TestCodecAndBytes(t, IntWithLessBytesAndOtherUnsigned{A: math.MaxInt16}, []byte{0xff, 0x7f, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, IntWithLessBytesAndOtherUnsigned{B: math.MaxUint16}, []byte{0x0, 0x0, 0xff, 0xff})

	bcs.TestEncodeErr(t, IntWithLessBytesAndOtherUnsigned{A: math.MaxUint16})
	bcs.TestEncodeErr(t, IntWithLessBytesAndOtherUnsigned{B: -1})
	bcs.TestEncodeErr(t, IntWithLessBytesAndOtherUnsigned{B: math.MaxUint16 + 1})

	bcs.TestCodecAndBytes(t, IntWithMoreBytesAndOtherUnsigned{A: 42}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, IntWithMoreBytesAndOtherUnsigned{B: 42}, []byte{0x0, 0x0, 0x0, 0x0, 0x2a, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, IntWithMoreBytesAndOtherUnsigned{A: math.MaxUint16}, []byte{0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, IntWithMoreBytesAndOtherUnsigned{B: math.MaxInt16}, []byte{0x0, 0x0, 0x0, 0x0, 0xff, 0x7f, 0x0, 0x0})

	bcs.TestEncodeErr(t, IntWithMoreBytesAndOtherUnsigned{B: -1})
}

type WithInit struct {
	A int
}

func (w *WithInit) BCSInit() error {
	w.A += 10
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
