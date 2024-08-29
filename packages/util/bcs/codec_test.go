package bcs_test

import (
	"bytes"
	"fmt"
	"io"
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

type WithNestedCustomCodec struct {
	A int `bcs:"bytes=1"`
	B WithCustomCodec
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

func testCodecErr[V any](t *testing.T, v V) {
	_, err := bcs.Marshal(v)
	require.Error(t, err)
}

func testCodec[V any](t *testing.T, v V, expectedEnc []byte) {
	vEnc := lo.Must1(bcs.Marshal(v))
	vDec := lo.Must1(bcs.Unmarshal[V](vEnc))
	require.Equal(t, v, vDec)
	require.Equal(t, expectedEnc, vEnc)

	vEncExternal := lo.Must1(ref_bcs.Marshal(v))
	require.Equal(t, vEncExternal, vEnc)
}

// Does not use reference implementation for encoding
func testCodecNoRef[V any](t *testing.T, v V, expectedEnc []byte) {
	vEnc := lo.Must1(bcs.Marshal(v))
	vDec := lo.Must1(bcs.Unmarshal[V](vEnc))
	require.Equal(t, v, vDec)
	require.Equal(t, expectedEnc, vEnc)
}

func TestCodecBasic(t *testing.T) {
	// Boolean	                         t/f    01/00
	// 8-bit signed                       -1    FF
	// 8-bit unsigned                      1    01
	// 16-bit signed                   -4660    CC ED
	// 16-bit unsigned                  4660    34 12
	// 32-bit signed              -305419896    88 A9 CB ED
	// 32-bit unsigned             305419896    78 56 34 12
	// 64-bit signed    -1311768467750121216	00 11 32 54 87 A9 CB ED
	// 64-bit unsigned   1311768467750121216	00 EF CD AB 78 56 34 12

	testCodec(t, true, []byte{0x01})
	testCodec(t, false, []byte{0x00})

	testCodec(t, int8(-1), []byte{0xFF})
	testCodec(t, int8(-128), []byte{0x80})
	testCodec(t, int8(127), []byte{0x7f})

	testCodec(t, uint8(0), []byte{0x00})
	testCodec(t, uint8(1), []byte{0x01})
	testCodec(t, uint8(255), []byte{0xFF})

	testCodec(t, int16(-4660), []byte{0xCC, 0xED})
	testCodec(t, int16(-32768), []byte{0x00, 0x80})
	testCodec(t, int16(32767), []byte{0xFF, 0x7F})

	testCodec(t, uint16(4660), []byte{0x34, 0x12})
	testCodec(t, uint16(0), []byte{0x00, 0x00})
	testCodec(t, uint16(65535), []byte{0xFF, 0xFF})

	testCodec(t, int32(-305419896), []byte{0x88, 0xA9, 0xCB, 0xED})
	testCodec(t, int32(-2147483648), []byte{0x0, 0x0, 0x0, 0x80})
	testCodec(t, int32(2147483647), []byte{0xFF, 0xFF, 0xFF, 0x7F})

	testCodec(t, uint32(305419896), []byte{0x78, 0x56, 0x34, 0x12})
	testCodec(t, uint32(0), []byte{0x00, 0x00, 0x00, 0x00})
	testCodec(t, uint32(4294967295), []byte{0xFF, 0xFF, 0xFF, 0xFF})

	testCodec(t, int64(-1311768467750121216), []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED})
	testCodec(t, int64(-9223372036854775808), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80})
	testCodec(t, int64(9223372036854775807), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F})

	testCodecNoRef(t, int(-1311768467750121216), []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED})
	testCodecNoRef(t, int(-9223372036854775808), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80})
	testCodecNoRef(t, int(9223372036854775807), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F})

	testCodec(t, uint64(1311768467750121216), []byte{0x00, 0xEF, 0xCD, 0xAB, 0x78, 0x56, 0x34, 0x12})
	testCodec(t, uint64(0), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	testCodec(t, uint64(18446744073709551615), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
}

func TestCodecMultiRef(t *testing.T) {
	var vI int16 = 4660
	var pVI *int16 = &vI
	var ppVI **int16 = &pVI
	testCodec(t, ppVI, []byte{0x34, 0x12})

	pVI = nil
	testCodecErr(t, ppVI)

	var vM map[int16]bool = map[int16]bool{1: true, 2: false, 3: true}
	var pVM *map[int16]bool = &vM
	var ppVM **map[int16]bool = &pVM
	testCodecNoRef(t, ppVM, []byte{0x3, 0x1, 0x0, 0x1, 0x2, 0x0, 0x0, 0x3, 0x0, 0x1})
}

func TestCodecString(t *testing.T) {
	testCodec(t, "", []byte{0x0})
	testCodec(t, "qwerty", []byte{0x6, 0x71, 0x77, 0x65, 0x72, 0x74, 0x79})
	testCodec(t, "çå∞≠¢õß∂ƒ∫", []byte{24, 0xc3, 0xa7, 0xc3, 0xa5, 0xe2, 0x88, 0x9e, 0xe2, 0x89, 0xa0, 0xc2, 0xa2, 0xc3, 0xb5, 0xc3, 0x9f, 0xe2, 0x88, 0x82, 0xc6, 0x92, 0xe2, 0x88, 0xab})
	testCodec(t, strings.Repeat("a", 127), append([]byte{0x7f}, bytes.Repeat([]byte{0x61}, 127)...))
	testCodec(t, strings.Repeat("a", 128), append([]byte{0x80, 0x1}, bytes.Repeat([]byte{0x61}, 128)...))
	testCodec(t, strings.Repeat("a", 16383), append([]byte{0xff, 0x7f}, bytes.Repeat([]byte{0x61}, 16383)...))
	testCodec(t, strings.Repeat("a", 16384), append([]byte{0x80, 0x80, 0x1}, bytes.Repeat([]byte{0x61}, 16384)...))
}

func TestCodecCollections(t *testing.T) {
	testCodec(t, []int64{42, 43}, []byte{0x2, 0x2A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2B, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	testCodec(t, []int64(nil), []byte{0x0})

	testCodec(t, [3]int64{42, 43, 44}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	testCodec(t, []string{"aaa", "bbb"}, []byte{0x2, 0x3, 0x61, 0x61, 0x61, 0x3, 0x62, 0x62, 0x62})
	testCodec(t, [][]int16{{1, 2}, {3, 4, 5}}, []byte{0x2, 0x2, 0x1, 0x0, 0x2, 0x0, 0x3, 0x3, 0x0, 0x4, 0x0, 0x5, 0x0})

	mapEnc := []byte{0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1}
	testCodecNoRef(t, map[int16]bool{3: true, 1: false, 2: true}, mapEnc)
	testCodecNoRef(t, map[int16]bool{2: true, 1: false, 3: true}, mapEnc)
	testCodecNoRef(t, map[int16]bool{}, []byte{0x0})
}

func TestCodecArraySize(t *testing.T) {
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

func TestCustom(t *testing.T) {
	testUint128(t, "10", true)
	testUint128(t, "1770887431076116955186", true)
	testUint128(t, "999999999999999999999999999999999999999999999999999", false)

	testCodecNoRef(t, time.Unix(12345, 6789), []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})
}

func testUint128(t *testing.T, v string, expectSuccess bool) {
	var bi big.Int
	_, ok := bi.SetString(v, 10)
	require.True(t, ok)

	if expectSuccess {
		refBiEnc := ref_bcs.MustMarshal(lo.Must1(ref_bcs.NewUint128FromBigInt(&bi)))

		testCodecNoRef(t, bi, refBiEnc)
		testCodecNoRef(t, &bi, refBiEnc)
	} else {
		testCodecErr(t, bi)
		testCodecErr(t, &bi)

		_, err := ref_bcs.NewUint128FromBigInt(&bi)
		require.Error(t, err)
	}
}

func TestStruct(t *testing.T) {
	testCodec(t, BasicStruct{A: 42, B: "aaa"}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	testCodecNoRef(t, IntWithLessBytes{A: 42}, []byte{42, 0})
	testCodecNoRef(t, IntWithMoreBytes{A: 42}, []byte{42, 0, 0, 0})
	vI := int64(42)
	pVI := &vI
	testCodec(t, IntPtr{A: &vI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	testCodecErr(t, IntPtr{A: nil})
	testCodec(t, IntMultiPtr{A: &pVI}, []byte{42, 0, 0, 0, 0, 0, 0, 0})
	testCodecErr(t, IntMultiPtr{A: nil})
	testCodec(t, IntOptional{A: &vI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	testCodec(t, IntOptionalPtr{A: &pVI}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0})
	testCodec(t, NestedStruct{A: 42, B: BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	testCodec(t, OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 1, 43, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97})
	testCodec(t, OptionalNestedStruct{A: 42, B: nil}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 0})
	testCodec(t, EmbeddedStruct{BasicStruct: BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	testCodec(t, OptionalEmbeddedStruct{BasicStruct: &BasicStruct{A: 42, B: "aaa"}, C: 43}, []byte{1, 42, 0, 0, 0, 0, 0, 0, 0, 3, 97, 97, 97, 43, 0, 0, 0, 0, 0, 0, 0})
	testCodec(t, OptionalEmbeddedStruct{BasicStruct: nil, C: 43}, []byte{0, 43, 0, 0, 0, 0, 0, 0, 0})
	testCodec(t, WithSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	testCodec(t, WithSlice{A: nil}, []byte{0x0})
	testCodec(t, WithOptionalSlice{A: &[]int32{42, 43}}, []byte{1, 0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	testCodec(t, WithOptionalSlice{A: nil}, []byte{0x0})
	testCodecNoRef(t, WithShortSlice{A: []int32{42, 43}}, []byte{0x2, 0x2a, 0x0, 0x0, 0x0, 0x2b, 0x0, 0x0, 0x0})
	testCodec(t, WithArray{A: [3]int16{42, 43, 44}}, []byte{0x2a, 0x0, 0x2b, 0x0, 0x2c, 0x0})
	testCodecNoRef(t, WithMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	testCodecNoRef(t, WithMap{A: map[int16]bool{}}, []byte{0x0})
	testCodecErr(t, WithMap{A: nil})
	testCodecNoRef(t, WithOptionalMap{A: map[int16]bool{}}, []byte{0x1, 0x0})
	testCodecNoRef(t, WithOptionalMap{A: nil}, []byte{0x0})
	testCodecNoRef(t, WithOptionalMap{A: map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	testCodecNoRef(t, WithOptionalMapPtr{A: &map[int16]bool{3: true, 1: false, 2: true}}, []byte{0x1, 0x3, 0x1, 0x0, 0x0, 0x2, 0x0, 0x1, 0x3, 0x0, 0x1})
	var m map[int16]bool
	testCodecErr(t, WithOptionalMapPtr{A: &m})
	testCodecNoRef(t, WithCustomCodec{}, []byte{0x1, 0x2, 0x3})
	testCodecNoRef(t, WithNestedCustomCodec{A: 43, B: WithCustomCodec{}}, []byte{0x2b, 0x1, 0x2, 0x3})
	testCodecNoRef(t, WithBCSOpts{A: 42}, []byte{0x2A, 0x0})
	testCodecNoRef(t, WithBCSOptsOverride{A: 42}, []byte{0x2A})
	testCodecNoRef(t, WithBigIntPtr{A: big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	testCodecNoRef(t, WithBigIntVal{A: *big.NewInt(42)}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	testCodecNoRef(t, WithTime{A: time.Unix(12345, 6789)}, []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})
}

func TestCodecStructUnexported(t *testing.T) {
	v := WitUnexported{A: 42, b: 43, c: 44, D: 45}
	vEnc := lo.Must1(bcs.Marshal(v))
	vDec := lo.Must1(bcs.Unmarshal[WitUnexported](vEnc))
	require.NotEqual(t, v, vDec)
	require.Equal(t, 0, vDec.b)
	require.Equal(t, 0, vDec.D)
	vDec.b = 43
	vDec.D = 45
	require.Equal(t, v, vDec)
}

type ExampleStruct struct {
	A int
	B int `bcs:"bytes=2"`
	C ExampleNestedStruct
}

func (s *ExampleStruct) Write(dest io.Writer) error {
	w := rwutil.NewWriter(dest)

	w.WriteInt64(int64(s.A))
	w.WriteInt16(int16(s.B))
	w.Write(&s.C)

	return nil
}

func (s *ExampleStruct) Read(src io.Reader) error {
	r := rwutil.NewReader(src)

	s.A = int(r.ReadInt64())
	s.B = int(r.ReadInt16())
	r.Read(&s.C)

	return r.Err
}

type ExampleNestedStruct struct {
	C int
	D []string `bcs:"len_bytes=2"`
}

func (s *ExampleNestedStruct) Write(dest io.Writer) error {
	w := rwutil.NewWriter(dest)

	w.WriteInt64(int64(s.C))
	w.WriteSize16(len(s.D))
	for _, v := range s.D {
		w.WriteString(v)
	}

	return nil
}

func (s *ExampleNestedStruct) Read(src io.Reader) error {
	r := rwutil.NewReader(src)

	s.C = int(r.ReadInt64())
	size := r.ReadSize16()
	s.D = make([]string, size)
	for i := range s.D {
		s.D[i] = r.ReadString()
	}

	return r.Err
}

func TestVsRwutil(t *testing.T) {
	v := ExampleStruct{
		A: 42,
		B: 43,
		C: ExampleNestedStruct{
			C: 44,
			D: []string{"aaa", "bbb"},
		},
	}

	vEnc := lo.Must1(bcs.Marshal(v))
	vDec := lo.Must1(bcs.Unmarshal[ExampleStruct](vEnc))
	require.Equal(t, v, vDec)

	written := rwutil.WriteToBytes(&v)
	require.Equal(t, written, vEnc)

	var read ExampleStruct
	rwutil.ReadFromBytes(written, &read)
	require.Equal(t, v, read)

	var readFromEnc ExampleStruct
	rwutil.ReadFromBytes(vEnc, &readFromEnc)
	require.Equal(t, v, readFromEnc)
}
