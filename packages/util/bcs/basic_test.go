package bcs_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

type BasicWithCustomCodec string

func (w BasicWithCustomCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Write([]byte{1, 2, 3})
	e.Encode(string(w))
	return nil
}

func (w *BasicWithCustomCodec) UnmarshalBCS(d *bcs.Decoder) error {
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

	*w = BasicWithCustomCodec(s)

	return nil
}

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

type BasicWithCustomPtrCodec string

func (w *BasicWithCustomPtrCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Write([]byte{1, 2, 3})
	e.Encode(string(*w))
	return nil
}

func (w *BasicWithCustomPtrCodec) UnmarshalBCS(d *bcs.Decoder) error {
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

	*w = BasicWithCustomPtrCodec(s)

	return nil
}

type BasicWithInit string

func (w *BasicWithInit) BCSInit() error {
	*w = *w + "!"
	return nil
}

type BasicWithCustomAndInit string

func (w *BasicWithCustomAndInit) MarshalBCS(e *bcs.Encoder) error {
	e.Write([]byte{1, 2, 3})
	e.Encode(string(*w))
	return nil
}

func (w *BasicWithCustomAndInit) UnmarshalBCS(d *bcs.Decoder) error {
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

	*w = BasicWithCustomAndInit(s)

	return nil
}

func (w *BasicWithCustomAndInit) BCSInit() error {
	*w = *w + "!"
	return nil
}

type InfWithCustomCodec interface{}

func TestMarshalAny(t *testing.T) {
	var v any = "hello"
	vEnc := bcs.MustMarshal(&v)
	vDec := bcs.MustUnmarshal[string](vEnc)
	require.Equal(t, v, vDec)

	type SameAsAny any
	var vSameAsAny SameAsAny = "hello"
	_, err := bcs.Marshal(&vSameAsAny)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not registered as enum")
}

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

	bcs.TestCodecAndBytesVsRef(t, true, []byte{0x01})
	bcs.TestCodecAndBytesVsRef(t, false, []byte{0x00})

	bcs.TestCodecAndBytesVsRef(t, int8(-1), []byte{0xFF})
	bcs.TestCodecAndBytesVsRef(t, int8(-128), []byte{0x80})
	bcs.TestCodecAndBytesVsRef(t, int8(127), []byte{0x7f})

	bcs.TestCodecAndBytesVsRef(t, uint8(0), []byte{0x00})
	bcs.TestCodecAndBytesVsRef(t, uint8(1), []byte{0x01})
	bcs.TestCodecAndBytesVsRef(t, uint8(255), []byte{0xFF})

	bcs.TestCodecAndBytesVsRef(t, int16(-4660), []byte{0xCC, 0xED})
	bcs.TestCodecAndBytesVsRef(t, int16(-32768), []byte{0x00, 0x80})
	bcs.TestCodecAndBytesVsRef(t, int16(32767), []byte{0xFF, 0x7F})

	bcs.TestCodecAndBytesVsRef(t, uint16(4660), []byte{0x34, 0x12})
	bcs.TestCodecAndBytesVsRef(t, uint16(0), []byte{0x00, 0x00})
	bcs.TestCodecAndBytesVsRef(t, uint16(65535), []byte{0xFF, 0xFF})

	bcs.TestCodecAndBytesVsRef(t, int32(-305419896), []byte{0x88, 0xA9, 0xCB, 0xED})
	bcs.TestCodecAndBytesVsRef(t, int32(-2147483648), []byte{0x0, 0x0, 0x0, 0x80})
	bcs.TestCodecAndBytesVsRef(t, int32(2147483647), []byte{0xFF, 0xFF, 0xFF, 0x7F})

	bcs.TestCodecAndBytesVsRef(t, uint32(305419896), []byte{0x78, 0x56, 0x34, 0x12})
	bcs.TestCodecAndBytesVsRef(t, uint32(0), []byte{0x00, 0x00, 0x00, 0x00})
	bcs.TestCodecAndBytesVsRef(t, uint32(4294967295), []byte{0xFF, 0xFF, 0xFF, 0xFF})

	bcs.TestCodecAndBytesVsRef(t, int64(-1311768467750121216), []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED})
	bcs.TestCodecAndBytesVsRef(t, int64(-9223372036854775808), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80})
	bcs.TestCodecAndBytesVsRef(t, int64(9223372036854775807), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F})

	bcs.TestCodecAndBytes(t, int(-1311768467750121216), []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xA9, 0xCB, 0xED})
	bcs.TestCodecAndBytes(t, int(-9223372036854775808), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80})
	bcs.TestCodecAndBytes(t, int(9223372036854775807), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F})

	bcs.TestCodecAndBytesVsRef(t, uint64(1311768467750121216), []byte{0x00, 0xEF, 0xCD, 0xAB, 0x78, 0x56, 0x34, 0x12})
	bcs.TestCodecAndBytesVsRef(t, uint64(0), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	bcs.TestCodecAndBytesVsRef(t, uint64(18446744073709551615), []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	bcs.TestCodecAndBytes(t, BasicWithCustomCodec("aaa"), []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, lo.ToPtr[BasicWithCustomPtrCodec]("aaa"), []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, BasicWithRwUtilCodec("aaa"), []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61})
}

func TestBasicInit(t *testing.T) {
	vInit := BasicWithInit("aaa")
	vEnc := bcs.MustMarshal(&vInit)
	require.Equal(t, []byte{0x3, 0x61, 0x61, 0x61}, vEnc)
	vInitDec := bcs.MustUnmarshal[BasicWithInit](vEnc)
	require.NotEqual(t, vInit, vInitDec)
	require.Equal(t, vInit+"!", vInitDec)

	vCustomAndInit := BasicWithCustomAndInit("aaa")
	vEnc = bcs.MustMarshal(&vCustomAndInit)
	require.Equal(t, []byte{0x1, 0x2, 0x3, 0x3, 0x61, 0x61, 0x61}, vEnc)
	vCustomAndInitDec := bcs.MustUnmarshal[BasicWithCustomAndInit](vEnc)
	require.NotEqual(t, vCustomAndInit, vCustomAndInitDec)
	require.Equal(t, vCustomAndInit+"!", vCustomAndInitDec)
}

func TestBasicCodedErr(t *testing.T) {
	_, err := bcs.Marshal((*int)(nil))
	require.Error(t, err)
	_, err = bcs.Unmarshal[int](nil)
	require.Error(t, err)

	err = bcs.NewDecoder(bytes.NewReader(nil)).Decode((*int)(nil))
	require.Error(t, err)

	var v int
	err = bcs.NewDecoder(bytes.NewReader(nil)).Decode(&v)
	require.Error(t, err)
}

func TestMultiPtrCodec(t *testing.T) {
	var vI int16 = 4660
	var pVI *int16 = &vI
	var ppVI **int16 = &pVI
	bcs.TestCodecAndBytesVsRef(t, ppVI, []byte{0x34, 0x12})

	pVI = nil
	bcs.TestEncodeErr(t, ppVI)

	var vM map[int16]bool = map[int16]bool{1: true, 2: false, 3: true}
	var pVM *map[int16]bool = &vM
	var ppVM **map[int16]bool = &pVM
	bcs.TestCodecAndBytes(t, ppVM, []byte{0x3, 0x1, 0x0, 0x1, 0x2, 0x0, 0x0, 0x3, 0x0, 0x1})

	type testStruct struct {
		A **bool
	}

	a := true
	pa := &a
	v := &testStruct{A: &pa}
	pv := &v

	vEnc := bcs.MustMarshal(&pv)
	vDec := bcs.MustUnmarshal[****testStruct](vEnc)
	require.Equal(t, v, ***vDec)
}

func TestStringCodec(t *testing.T) {
	bcs.TestCodecAndBytesVsRef(t, "", []byte{0x0})
	bcs.TestCodecAndBytesVsRef(t, "qwerty", []byte{0x6, 0x71, 0x77, 0x65, 0x72, 0x74, 0x79})
	bcs.TestCodecAndBytesVsRef(t, "çå∞≠¢õß∂ƒ∫", []byte{24, 0xc3, 0xa7, 0xc3, 0xa5, 0xe2, 0x88, 0x9e, 0xe2, 0x89, 0xa0, 0xc2, 0xa2, 0xc3, 0xb5, 0xc3, 0x9f, 0xe2, 0x88, 0x82, 0xc6, 0x92, 0xe2, 0x88, 0xab})
	bcs.TestCodecAndBytesVsRef(t, strings.Repeat("a", 127), append([]byte{0x7f}, bytes.Repeat([]byte{0x61}, 127)...))
	bcs.TestCodecAndBytesVsRef(t, strings.Repeat("a", 128), append([]byte{0x80, 0x1}, bytes.Repeat([]byte{0x61}, 128)...))
	bcs.TestCodecAndBytesVsRef(t, strings.Repeat("a", 16383), append([]byte{0xff, 0x7f}, bytes.Repeat([]byte{0x61}, 16383)...))
	bcs.TestCodecAndBytesVsRef(t, strings.Repeat("a", 16384), append([]byte{0x80, 0x80, 0x1}, bytes.Repeat([]byte{0x61}, 16384)...))
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

type StructWithManualCodec struct {
	A int
	B *string `bcs:"optional"`
	C BasicStructEnum
	D []int8
	E int16
	F string
}

func (s *StructWithManualCodec) MarshalBCS(e *bcs.Encoder) error {
	e.Encode(s.A)
	e.EncodeOptional(s.B)

	switch {
	case s.C.A != nil:
		e.WriteEnumIdx(0)
		e.Encode(*s.C.A)
	case s.C.B != nil:
		e.WriteEnumIdx(1)
		e.Encode(*s.C.B)
	case s.C.C != nil:
		e.WriteEnumIdx(2)
		e.Encode(*s.C.C)
	default:
		panic("enum variant not set")
	}

	e.WriteLen(len(s.D))
	for _, v := range s.D {
		e.Encode(v)
	}

	e.WriteInt16(s.E)
	e.WriteString(s.F)

	// Just to mark that this is a manual codec.
	return e.Encode(byte(1))
}

func (s *StructWithManualCodec) UnmarshalBCS(d *bcs.Decoder) error {
	d.Decode(&s.A)
	d.DecodeOptional(&s.B)

	variantIdx := d.ReadEnumIdx()

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

	l := d.ReadLen()
	for i := 0; i < l; i++ {
		s.D = append(s.D, bcs.MustDecode[int8](d))
	}

	s.E = d.ReadInt16()
	s.F = d.ReadString()

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
		D: []int8{1, 2, 3},
		E: 10,
		F: "ccc",
	}
	vEnc := bcs.TestCodec(t, v)
	require.Equal(t, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x3, 0x61, 0x61, 0x61, 0x1, 0x3, 0x62, 0x62, 0x62, 0x3, 0x1, 0x2, 0x3, 0xa, 0x0, 0x3, 0x63, 0x63, 0x63, 0x1}, vEnc)

	vAuto := StructWithAutoCodec(v)
	vAutoEnc := bcs.TestCodec(t, vAuto)
	vEncWithoutMarker := vEnc[:len(vEnc)-1]
	require.Equal(t, vEncWithoutMarker, vAutoEnc)
}

func TestInfWithCustomCodec(t *testing.T) {
	t.Cleanup(func() {
		bcs.RemoveCustomEncoder[InfWithCustomCodec]()
		bcs.RemoveCustomDecoder[InfWithCustomCodec]()
	})

	bcs.AddCustomEncoder(func(e *bcs.Encoder, v InfWithCustomCodec) error {
		switch v := v.(type) {
		case string:
			e.WriteEnumIdx(0)
			e.WriteString(v)
		case int:
			e.WriteEnumIdx(1)
			e.WriteInt32(int32(v))
		default:
			return fmt.Errorf("unsupported type: %T", v)
		}

		return nil
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, v *InfWithCustomCodec) error {
		switch d.ReadEnumIdx() {
		case 0:
			*v = d.ReadString()
		case 1:
			*v = d.ReadInt()
		default:
			return fmt.Errorf("invalid enum variant")
		}

		return nil
	})

	bcs.TestCodec(t, InfWithCustomCodec("aaa"))
}
