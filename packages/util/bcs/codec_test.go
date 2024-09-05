package bcs_test

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
	"time"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestBasicTypesCodec(t *testing.T) {
	// Boolean	                         t/f    01/00
	// 8-bit signed                       -1    FF
	// 8-bit unsigned                      1    01
	// 16-bit signed                   -4660    CC ED
	// 16-bit unsigned                  4660    34 12
	// 32-bit signed              -305419896    88 A9 CB ED
	// 32-bit unsigned             305419896    78 56 34 12
	// 64-bit signed    -1311768467750121216	00 11 32 54 87 A9 CB ED
	// 64-bit unsigned   1311768467750121216	00 EF CD AB 78 56 34 12

	bcs.TestCodecAndBytes(t, true, []byte{0x01})
	bcs.TestCodecAndBytes(t, false, []byte{0x00})

	bcs.TestCodecAndBytes(t, int8(-1), []byte{0xFF})
	bcs.TestCodecAndBytes(t, int8(-128), []byte{0x80})
	bcs.TestCodecAndBytes(t, int8(127), []byte{0x7f})

	bcs.TestCodecAndBytes(t, uint8(0), []byte{0x00})
	bcs.TestCodecAndBytes(t, uint8(1), []byte{0x01})
	bcs.TestCodecAndBytes(t, uint8(255), []byte{0xFF})

	bcs.TestCodecAndBytes(t, int16(-4660), []byte{0xCC, 0xED})
	bcs.TestCodecAndBytes(t, int16(-32768), []byte{0x00, 0x80})
	bcs.TestCodecAndBytes(t, int16(32767), []byte{0xFF, 0x7F})

	bcs.TestCodecAndBytes(t, uint16(4660), []byte{0x34, 0x12})
	bcs.TestCodecAndBytes(t, uint16(0), []byte{0x00, 0x00})
	bcs.TestCodecAndBytes(t, uint16(65535), []byte{0xFF, 0xFF})

	bcs.TestCodecAndBytes(t, int32(-305419896), []byte{0x88, 0xA9, 0xCB, 0xED})
	bcs.TestCodecAndBytes(t, int32(-2147483648), []byte{0x0, 0x0, 0x0, 0x80})
	bcs.TestCodecAndBytes(t, int32(2147483647), []byte{0xFF, 0xFF, 0xFF, 0x7F})

	bcs.TestCodecAndBytes(t, uint32(305419896), []byte{0x78, 0x56, 0x34, 0x12})
	bcs.TestCodecAndBytes(t, uint32(0), []byte{0x00, 0x00, 0x00, 0x00})
	bcs.TestCodecAndBytes(t, uint32(4294967295), []byte{0xFF, 0xFF, 0xFF, 0xFF})

	bcs.TestCodecAndBytes(t, int64(-1311768467750121216), []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED})
	bcs.TestCodecAndBytes(t, int64(-9223372036854775808), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80})
	bcs.TestCodecAndBytes(t, int64(9223372036854775807), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F})

	bcs.TestCodecAndBytesNoRef(t, int(-1311768467750121216), []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED})
	bcs.TestCodecAndBytesNoRef(t, int(-9223372036854775808), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80})
	bcs.TestCodecAndBytesNoRef(t, int(9223372036854775807), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F})

	bcs.TestCodecAndBytes(t, uint64(1311768467750121216), []byte{0x00, 0xEF, 0xCD, 0xAB, 0x78, 0x56, 0x34, 0x12})
	bcs.TestCodecAndBytes(t, uint64(0), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	bcs.TestCodecAndBytes(t, uint64(18446744073709551615), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	bcs.TestCodecAndBytesNoRef(t, BaseWithCustomCodec("aaa"), []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, lo.ToPtr[BaseWithCustomPtrCodec]("aaa"), []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61})
}

func TestMultiPtrCodec(t *testing.T) {
	var vI int16 = 4660
	var pVI *int16 = &vI
	var ppVI **int16 = &pVI
	bcs.TestCodecAndBytes(t, ppVI, []byte{0x34, 0x12})

	pVI = nil
	bcs.TestEncodeErr(t, ppVI)

	var vM map[int16]bool = map[int16]bool{1: true, 2: false, 3: true}
	var pVM *map[int16]bool = &vM
	var ppVM **map[int16]bool = &pVM
	bcs.TestCodecAndBytesNoRef(t, ppVM, []byte{0x3, 0x1, 0x0, 0x1, 0x2, 0x0, 0x0, 0x3, 0x0, 0x1})
}

func TestStringCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, "", []byte{0x0})
	bcs.TestCodecAndBytes(t, "qwerty", []byte{0x6, 0x71, 0x77, 0x65, 0x72, 0x74, 0x79})
	bcs.TestCodecAndBytes(t, "çå∞≠¢õß∂ƒ∫", []byte{24, 0xc3, 0xa7, 0xc3, 0xa5, 0xe2, 0x88, 0x9e, 0xe2, 0x89, 0xa0, 0xc2, 0xa2, 0xc3, 0xb5, 0xc3, 0x9f, 0xe2, 0x88, 0x82, 0xc6, 0x92, 0xe2, 0x88, 0xab})
	bcs.TestCodecAndBytes(t, strings.Repeat("a", 127), append([]byte{0x7f}, bytes.Repeat([]byte{0x61}, 127)...))
	bcs.TestCodecAndBytes(t, strings.Repeat("a", 128), append([]byte{0x80, 0x1}, bytes.Repeat([]byte{0x61}, 128)...))
	bcs.TestCodecAndBytes(t, strings.Repeat("a", 16383), append([]byte{0xff, 0x7f}, bytes.Repeat([]byte{0x61}, 16383)...))
	bcs.TestCodecAndBytes(t, strings.Repeat("a", 16384), append([]byte{0x80, 0x80, 0x1}, bytes.Repeat([]byte{0x61}, 16384)...))
}

func TestArrayCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, []int64{42, 43}, []byte{0x2, 0x2A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2B, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, []int8{42, 43}, []byte{0x2, 0x2A, 0x2B})
	bcs.TestCodecAndBytes(t, []int8(nil), []byte{0x0})
	bcs.TestCodecAndBytes(t, []uint8{42, 43}, []byte{0x2, 0x2A, 0x2B})
	bcs.TestCodecAndBytes(t, []int64(nil), []byte{0x0})

	bcs.TestCodecAndBytes(t, []*int16{lo.ToPtr[int16](1), lo.ToPtr[int16](2), lo.ToPtr[int16](3)}, []byte{0x3, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0})
	bcs.TestCodecAndBytes(t, []*byte{lo.ToPtr[byte](42), lo.ToPtr[byte](43)}, []byte{0x2, 0x2A, 0x2B})
	bcs.TestCodecAndBytes(t, []*int8{lo.ToPtr[int8](42), lo.ToPtr[int8](43)}, []byte{0x2, 0x2A, 0x2B})

	bcs.TestCodecAndBytesNoRef(t, []BaseWithCustomCodec{"a", "b"}, []byte{0x2, 0x1, 0x2, 0x3, 0x1, 0x61, 0x1, 0x2, 0x3, 0x1, 0x62})
	bcs.TestCodecAndBytesNoRef(t, []*BaseWithCustomCodec{lo.ToPtr[BaseWithCustomCodec]("a"), lo.ToPtr[BaseWithCustomCodec]("b")}, []byte{0x2, 0x1, 0x2, 0x3, 0x1, 0x61, 0x1, 0x2, 0x3, 0x1, 0x62})
	bcs.TestCodecAndBytesNoRef(t, []BaseWithCustomPtrCodec{"a", "b"}, []byte{0x2, 0x1, 0x2, 0x3, 0x1, 0x61, 0x1, 0x2, 0x3, 0x1, 0x62})

	bcs.TestCodecAndBytes(t, [3]int64{42, 43, 44}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, [3]byte{42, 43, 44}, []byte{0x2a, 0x2b, 0x2c})

	bcs.TestCodecAndBytes(t, []string{"aaa", "bbb"}, []byte{0x2, 0x3, 0x61, 0x61, 0x61, 0x3, 0x62, 0x62, 0x62})
	bcs.TestCodecAndBytes(t, [][]int16{{1, 2}, {3, 4, 5}}, []byte{0x2, 0x2, 0x1, 0x0, 0x2, 0x0, 0x3, 0x3, 0x0, 0x4, 0x0, 0x5, 0x0})
}

func TestMapCodec(t *testing.T) {
	intMapEnc := []byte{0x3, 0x0, 0x0, 0x0, 0x3, 0x0, 0x1, 0xfd, 0xff, 0x1}
	bcs.TestCodecAndBytesNoRef(t, map[int16]bool{-3: true, 0: false, 3: true}, intMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[int16]bool{3: true, 0: false, -3: true}, intMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[int16]bool{}, []byte{0x0})

	uintMapEnc := []byte{0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1}
	bcs.TestCodecAndBytesNoRef(t, map[uint16]bool{3: true, 1: false, 2: true}, uintMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[uint16]bool{2: true, 1: false, 3: true}, uintMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[uint16]bool{}, []byte{0x0})

	strMapEnc := []byte{0x3, 0x2, 0x61, 0x61, 0x0, 0x2, 0x62, 0x62, 0x1, 0x2, 0x63, 0x63, 0x1}
	bcs.TestCodecAndBytesNoRef(t, map[string]bool{"cc": true, "aa": false, "bb": true}, strMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[string]bool{"bb": true, "aa": false, "cc": true}, strMapEnc)

	intMapOfMapsEnc := []byte{0x2, 0x1, 0x0, 0x2, 0x2, 0x0, 0x1, 0x3, 0x0, 0x0, 0x2, 0x0, 0x1, 0x1, 0x0, 0x1}
	bcs.TestCodecAndBytesNoRef(t, map[int16]map[int16]bool{1: {2: true, 3: false}, 2: {1: true}}, intMapOfMapsEnc)

	customMapEnc := []byte{0x2, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61, 0x1, 0x2, 0x3, 0x1, 0x2, 0x3, 0x2, 0x62, 0x62, 0x1, 0x2, 0x3}
	bcs.TestCodecAndBytesNoRef(t, map[BaseWithCustomCodec]WithCustomCodec{"bb": {}, "aa": {}}, customMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[BaseWithCustomCodec]WithCustomCodec{"aa": {}, "bb": {}}, customMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[BaseWithCustomCodec]*WithCustomCodec{"bb": {}, "aa": {}}, customMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[BaseWithCustomCodec]*WithCustomCodec{"aa": {}, "bb": {}}, customMapEnc)
	bcs.TestCodecAndBytesNoRef(t, map[BaseWithCustomPtrCodec]*WithCustomCodec{"aa": {}, "bb": {}}, customMapEnc)

	customMapEnc = []byte{0x2, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61, 0x1, 0x2, 0x3, 0x2, 0x63, 0x63, 0x1, 0x2, 0x3, 0x2, 0x62, 0x62, 0x1, 0x2, 0x3, 0x2, 0x64, 0x64}
	bcs.TestCodecAndBytesNoRef(t, map[BaseWithCustomPtrCodec]BaseWithCustomPtrCodec{"aa": "cc", "bb": "dd"}, customMapEnc)
}

func TestCollectionSizeCodec(t *testing.T) {
	testSizeCodec(t, 0, []byte{0x0})
	testSizeCodec(t, 1, []byte{0x1})
	testSizeCodec(t, 127, []byte{0x7F})
	testSizeCodec(t, 128, []byte{0x80, 0x1})
	testSizeCodec(t, 16383, []byte{0xFF, 0x7F})
	testSizeCodec(t, 16384, []byte{0x80, 0x80, 0x1})
	testSizeCodec(t, 2097151, []byte{0xFF, 0xFF, 0x7F})
	testSizeCodec(t, 2097152, []byte{0x80, 0x80, 0x80, 0x1})
	testSizeCodec(t, 268435455, []byte{0xFF, 0xFF, 0xFF, 0x7F})
	testSizeCodec(t, 268435456, []byte{0x80, 0x80, 0x80, 0x80, 0x1})
	testSizeCodec(t, 2147483647, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x7})
	testSizeCodec(t, 4294967295, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xF})
	testSizeCodec(t, math.MaxUint32, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xF})
}

func testSizeCodec(t *testing.T, v int, expectedEnc []byte) {
	var rwutilEncBuff bytes.Buffer
	r := rwutil.NewWriter(&rwutilEncBuff)
	r.WriteSizeWithLimit(v, math.MaxUint32)
	require.NoError(t, r.Err)
	rwutilEnc := rwutilEncBuff.Bytes()

	refBcsEnc := ref_bcs.ULEB128Encode(v)

	require.Equal(t, expectedEnc, rwutilEnc)
	require.Equal(t, refBcsEnc, rwutilEnc)
}

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
	ByteArrayVal []BasicStruct `bcs:"elem_bytearr"`
}

type WithByteArrEntry struct {
	ByteArrayVal map[int16]BasicStruct `bcs:"elem_bytearr"`
}

type WithByteArrByte struct {
	A []byte `bcs:"elem_bytearr"`
}

type WithByteArrInt struct {
	ByteArrVal []int32 `bcs:"elem_bytearr"`
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

type BaseWithCustomCodec string

func (w BaseWithCustomCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Write([]byte{1, 2, 3})
	e.Encode(string(w))
	return nil
}

func (w *BaseWithCustomCodec) UnmarshalBCS(d *bcs.Decoder) error {
	b := make([]byte, 3)
	if _, err := d.Read(b); err != nil {
		return err
	}

	if b[0] != 1 || b[1] != 2 || b[2] != 3 {
		return fmt.Errorf("invalid value: %v", b)
	}

	var s string
	if err := d.Decode(&s); err != nil {
		return err
	}

	*w = BaseWithCustomCodec(s)

	return nil
}

type BaseWithCustomPtrCodec string

func (w *BaseWithCustomPtrCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Write([]byte{1, 2, 3})
	e.Encode(string(*w))
	return nil
}

func (w *BaseWithCustomPtrCodec) UnmarshalBCS(d *bcs.Decoder) error {
	b := make([]byte, 3)
	if _, err := d.Read(b); err != nil {
		return err
	}

	if b[0] != 1 || b[1] != 2 || b[2] != 3 {
		return fmt.Errorf("invalid value: %v", b)
	}

	var s string
	if err := d.Decode(&s); err != nil {
		return err
	}

	*w = BaseWithCustomPtrCodec(s)

	return nil
}

type WithNestedCustomCodec struct {
	A int `bcs:"bytes=1"`
	B WithCustomCodec
}

type WithNestedCustomPtrCodec struct {
	A int `bcs:"bytes=1"`
	B BaseWithCustomPtrCodec
}

type WithNestedPtrCustomPtrCodec struct {
	A int `bcs:"bytes=1"`
	B *BaseWithCustomPtrCodec
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
	bcs.TestCodecAndBytes(t, BasicStruct{A: 42, B: "aaa"}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytesNoRef(t, IntWithLessBytes{A: 42}, []byte{42, 0})
	bcs.TestCodecAndBytesNoRef(t, IntWithMoreBytes{A: 42}, []byte{42, 0, 0, 0})
	vI := int64(42)
	pVI := &vI
	bcs.TestCodecAndBytes(t, IntPtr{A: &vI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestEncodeErr(t, IntPtr{A: nil})
	bcs.TestCodecAndBytes(t, IntMultiPtr{A: &pVI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestEncodeErr(t, IntMultiPtr{A: nil})
	bcs.TestCodecAndBytes(t, IntOptional{A: &vI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytes(t, IntOptionalPtr{A: &pVI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytes(t, NestedStruct{A: 42, B: BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytes(t, OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytes(t, &OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	bcs.TestCodecAndBytes(t, OptionalNestedStruct{A: 42, B: nil}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytes(t, EmbeddedStruct{BasicStruct: BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytes(t, OptionalEmbeddedStruct{BasicStruct: &BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytes(t, OptionalEmbeddedStruct{BasicStruct: nil, C: 43}, []byte{0, 43, 0, 0, 0, 0, 0, 0, 0})
	bcs.TestCodecAndBytesNoRef(t, WithByteArr{A: "aaa", B: 42, C: "ccc"}, []byte{0x3, 0x61, 0x61, 0x61, 0x4, 0x2a, 0x0, 0x0, 0x0, 0x3, 0x63, 0x63, 0x63})
	bcs.TestCodecAndBytes(t, WithSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithSlice{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, WithOptionalSlice{A: &[]int32{42, 43}}, []byte{1, 0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithOptionalSlice{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytesNoRef(t, WithShortSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, WithArray{A: [3]int16{42, 43, 44}}, []byte{0x2a, 0x0, 0x2b, 0x0, 0x2c, 0x0})
	bcs.TestCodecAndBytesNoRef(t, WithByteArrElem{ByteArrayVal: []BasicStruct{{A: 42, B: "aaa"}, {A: 43, B: "bbb"}}}, []byte{0x2, 0xc, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x61, 0x61, 0x61, 0xc, 0x2b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x62, 0x62, 0x62})
	bcs.TestCodecAndBytesNoRef(t, WithByteArrInt{ByteArrVal: []int32{1, 2, 3}}, []byte{0x3, 0x4, 0x1, 0x0, 0x0, 0x0, 0x4, 0x2, 0x0, 0x0, 0x0, 0x4, 0x3, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, WithByteArrByte{A: []byte{1, 2, 3}}, []byte{0x3, 0x1, 0x1, 0x1, 0x2, 0x1, 0x3})
	bcs.TestCodecAndBytesNoRef(t, WithByteArrEntry{ByteArrayVal: map[int16]BasicStruct{1: {A: 42, B: "aaa"}}}, []byte{0x1, 0xe, 0x1, 0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, WithMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	bcs.TestCodecAndBytesNoRef(t, WithMap{A: map[int16]bool{}}, []byte{0x0})
	bcs.TestEncodeErr(t, WithMap{A: nil})
	bcs.TestCodecAndBytesNoRef(t, WithOptionalMap{A: map[int16]bool{}}, []byte{0x1, 0x0})
	bcs.TestCodecAndBytesNoRef(t, WithOptionalMap{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytesNoRef(t, WithOptionalMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	bcs.TestCodecAndBytesNoRef(t, WithOptionalMapPtr{A: &map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	var m map[int16]bool
	bcs.TestEncodeErr(t, WithOptionalMapPtr{A: &m})
	bcs.TestCodecAndBytesNoRef(t, WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytesNoRef(t, &WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	bcs.TestCodecAndBytesNoRef(t, WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytesNoRef(t, &WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytesNoRef(t, WithNestedCustomPtrCodec{A: 43, B: BaseWithCustomPtrCodec("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, &WithNestedCustomPtrCodec{A: 43, B: BaseWithCustomPtrCodec("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, WithNestedPtrCustomPtrCodec{A: 43, B: lo.ToPtr[BaseWithCustomPtrCodec]("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, &WithNestedPtrCustomPtrCodec{A: 43, B: lo.ToPtr[BaseWithCustomPtrCodec]("aa")}, []byte{0x2b, 0x1, 0x2, 0x3, 0x2, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	bcs.TestCodecAndBytesNoRef(t, &WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	bcs.TestCodecAndBytesNoRef(t, WithBCSOptsOverride{A: 42}, []byte{0x2A})
	bcs.TestCodecAndBytesNoRef(t, &WithBCSOptsOverride{A: 42}, []byte{0x2A})
	bcs.TestCodecAndBytesNoRef(t, WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, &WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, WithBigIntVal{A: *big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, WithTime{A: time.Unix(12345, 6789)}, []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})
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

func TestDecodingWithPresetMap(t *testing.T) {
	vEnc := bcs.MustMarshal(&WithMap{
		A: map[int16]bool{1: true, 2: false},
	})

	vDecMap := map[int16]bool{3: true}
	vDec := WithMap{
		A: vDecMap,
	}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)

	// NOTE: Preset maps are overriden, preset value is ignored, preset collection is not altered.
	require.Equal(t, map[int16]bool{1: true, 2: false}, vDec.A)
	require.Equal(t, map[int16]bool{3: true}, vDecMap)
}

func TestDecodingWithPresetSlice(t *testing.T) {
	vEnc := bcs.MustMarshal(&WithSlice{
		A: []int32{1, 2},
	})

	vDecArr := []int32{3}
	vDec := WithSlice{
		A: vDecArr,
	}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)

	// NOTE: Preset slices are overriden, preset value is ignored, preset collection is not altered.
	require.Equal(t, []int32{1, 2}, vDec.A)
	require.Equal(t, []int32{3}, vDecArr)
}

func TestDecodingWithPresetPtr(t *testing.T) {
	vEnc := bcs.MustMarshal(&IntPtr{
		A: lo.ToPtr[int64](42),
	})

	pv := lo.ToPtr[int64](43)
	vDec := IntPtr{
		A: pv,
	}

	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)
	require.Equal(t, lo.ToPtr[int64](42), vDec.A)

	// NOTE: Preset field pointers are KEPT. Their values ARE altered upon decoding.
	*pv = 10
	require.Equal(t, lo.ToPtr[int64](10), vDec.A)
}

func TestDecodingWithPresetOptional(t *testing.T) {
	vEnc := bcs.MustMarshal(&IntOptional{
		A: lo.ToPtr[int64](42),
	})

	vDecA := lo.ToPtr[int64](43)
	vDec := IntOptional{
		A: vDecA,
	}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)
	require.Equal(t, lo.ToPtr[int64](42), vDec.A)
	require.Equal(t, lo.ToPtr[int64](42), vDecA)

	vEnc = bcs.MustMarshal(&IntOptional{
		A: nil,
	})

	vDec = IntOptional{
		A: lo.ToPtr[int64](43),
	}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)

	// NOTE: Preset field pointers are KEPT. Their values ARE altered upon decoding, but ONLY if present.
	require.Equal(t, lo.ToPtr[int64](43), vDec.A)
}

func TestDecodingWithPresetNestedPtr(t *testing.T) {
	vEnc := bcs.MustMarshal(&OptionalNestedStruct{
		A: 42,
		B: &BasicStruct{A: 43, B: "aaa"},
	})

	pv := lo.ToPtr(BasicStruct{C: 20})
	vDec := OptionalNestedStruct{
		A: 45,
		B: pv,
	}

	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)
	require.Equal(t, &BasicStruct{A: 43, B: "aaa", C: 20}, vDec.B)

	// NOTE: Preset field pointers are KEPT. Their values ARE altered upon decoding, but ONLY if present.
	// This might be useful to preset some fields in decoded structure and pass it to other place where is decoded.
	pv.A = 10
	require.Equal(t, &BasicStruct{A: 10, B: "aaa", C: 20}, vDec.B)

	vEnc = bcs.MustMarshal(&OptionalNestedStruct{
		A: 42,
		B: nil,
	})

	pv = lo.ToPtr(BasicStruct{C: 20})
	vDec = OptionalNestedStruct{
		A: 45,
		B: pv,
	}

	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)
	require.Equal(t, &BasicStruct{C: 20}, vDec.B)
	pv.A = 10
	require.Equal(t, &BasicStruct{A: 10, C: 20}, vDec.B)
}

func TestDecodingWithPresetStructEnumVariant(t *testing.T) {
	vEnc := bcs.MustMarshal(&BasicStructEnum{B: lo.ToPtr("aaa")})

	vDecB := lo.ToPtr("bbb")
	vDec := BasicStructEnum{A: lo.ToPtr[int32](42), B: vDecB}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&vDec)
	// NOTE: Preset struct enum variants are KEPT - for a performance reason.
	require.Equal(t, BasicStructEnum{A: lo.ToPtr[int32](42), B: lo.ToPtr("aaa")}, vDec)
}
