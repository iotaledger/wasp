package bcs_test

import (
	"bytes"
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

type InfEnum1 interface{}

type InfEnum2 interface{}

type InfEnumWithMethods interface {
	Dummy()
}

type EnumVariant1 struct {
	A int
}

func (e EnumVariant1) Dummy() {}

type EnumVariant2 struct {
	A string
}

func (e EnumVariant2) Dummy() {}

func TestEnumRegister(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	require.Panics(t, func() {
		bcs.RegisterEnumType2[struct{}, int, string]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumType2[int, int, string]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumType2[InfEnum1, int, InfEnum2]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumType2[InfEnumWithMethods, int, string]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumType2[InfEnumWithMethods, EnumVariant1, string]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumTypeWithIDs[InfEnumWithMethods](map[bcs.EnumVariantID]any{
			1:  EnumVariant1{},
			-2: EnumVariant2{},
		})
	})

	bcs.RegisterEnumType2[InfEnum1, int, string]()
	bcs.RegisterEnumType3[InfEnum2, int64, int32, int8]()
	bcs.RegisterEnumType2[InfEnumWithMethods, EnumVariant1, EnumVariant2]()
}

type RefEnum1 struct {
	A *int
	B *string
}

func (e RefEnum1) IsBcsEnum() {}

type RefEnum2 struct {
	A *int64
	B *int32
	C *int8
}

func (e RefEnum2) IsBcsEnum() {}

type structWithField[FieldType any] struct {
	A FieldType
}

func declareStructWithField[FieldType any](v FieldType) structWithField[FieldType] {
	return structWithField[FieldType]{A: v}
}

type structWithOptionalField[FieldType any] struct {
	A FieldType `bcs:"optional"`
}

type structWithNotEnumField[FieldType any] struct {
	A FieldType `bcs:"not_enum"`
}

func TestBasicInterfaceEnumCodec(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	bcs.RegisterEnumType2[InfEnum1, int, string]()
	bcs.RegisterEnumType3[InfEnum2, int64, int32, int8]()
	bcs.RegisterEnumType2[InfEnumWithMethods, EnumVariant1, EnumVariant2]()

	vS := "foo"
	refEnumEnc := ref_bcs.MustMarshal(RefEnum1{B: &vS})
	require.NotEmpty(t, refEnumEnc)
	bcs.TestCodecAndBytes(t, declareStructWithField(InfEnum1(vS)), refEnumEnc)

	vI := int32(42)
	refEnumEnc = ref_bcs.MustMarshal(RefEnum2{B: &vI})
	require.NotEmpty(t, refEnumEnc)
	bcs.TestCodecAndBytes(t, declareStructWithField(InfEnum2(vI)), refEnumEnc)

	var e InfEnumWithMethods = EnumVariant1{A: 42}
	bcs.TestCodecAndBytes(t, declareStructWithField(e), []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, &e, []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	e = EnumVariant2{A: "bar"}
	bcs.TestCodecAndBytes(t, declareStructWithField(e), []byte{0x1, 0x3, 0x62, 0x61, 0x72})
	bcs.TestCodecAndBytes(t, &e, []byte{0x1, 0x3, 0x62, 0x61, 0x72})

	bcs.TestEncodeErr(t, InfEnum1(int8(42)))
	bcs.TestEncodeErr(t, InfEnum2(int(42)))
	bcs.TestEncodeErr(t, InfEnum2(nil))
}

func TestInterfaceEnumVariantWithCustomCodec(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	bcs.RegisterEnumType2[InfEnum1, WithCustomCodec, string]()

	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1](WithCustomCodec{}), []byte{0x0, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1]("aaa"), []byte{0x1, 0x3, 0x61, 0x61, 0x61})
}

func TestIntEnumWithInfEnumVariant(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	bcs.RegisterEnumType2[InfEnum1, int, *InfEnum2]()
	bcs.RegisterEnumType1[InfEnum2, string]()

	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1](42), []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1](lo.ToPtr[InfEnum2]("aaa")), []byte{0x1, 0x0, 0x3, 0x61, 0x61, 0x61})
}

func TestIntEnumWithStructEnumVariant(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	bcs.RegisterEnumType3[InfEnum1, BasicStructEnum, WithCustomCodec, string]()

	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1](BasicStructEnum{A: lo.ToPtr[int32](10)}), []byte{0x0, 0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1](BasicStructEnum{B: lo.ToPtr("aaa")}), []byte{0x0, 0x1, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnum1](WithCustomCodec{}), []byte{0x1, 0x1, 0x2, 0x3})
}

type NonEnumInf interface{}

type WithNonEnumInf struct {
	A NonEnumInf `bcs:"not_enum"`
}

func (s *WithNonEnumInf) UnmarshalBCS(d *bcs.Decoder) error {
	var a int
	d.MustDecode(&a)
	s.A = a
	return nil
}

type WithNonEnumInfNoDecode WithNonEnumInf

type WithNonEnumByteArrInf struct {
	A NonEnumInf `bcs:"not_enum,bytearr"`
	B int
}

func (s *WithNonEnumByteArrInf) UnmarshalBCS(d *bcs.Decoder) error {
	var b []byte
	d.MustDecode(&b)
	s.A = bcs.MustUnmarshal[int](b)

	d.MustDecode(&s.B)

	return nil
}

func TestInfNotEnum(t *testing.T) {
	bcs.TestCodecAndBytes(t, &WithNonEnumInf{A: 42}, []byte{0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	vEnc := bcs.MustMarshal(&WithNonEnumInfNoDecode{A: 42})
	_, err := bcs.Unmarshal[WithNonEnumInfNoDecode](vEnc)
	require.Error(t, err)

	vEnc = bcs.MustMarshal(&WithNonEnumInfNoDecode{A: 42})
	v := &WithNonEnumInfNoDecode{A: int(0)}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&v)
	require.Equal(t, 42, v.A)

	vEnc = bcs.MustMarshal(&WithNonEnumInfNoDecode{A: lo.ToPtr(43)})
	v = &WithNonEnumInfNoDecode{A: lo.ToPtr(0)}
	bcs.NewDecoder(bytes.NewReader(vEnc)).MustDecode(&v)
	require.Equal(t, lo.ToPtr(43), v.A)

	bcs.TestCodecAndBytes(t, &WithNonEnumByteArrInf{A: 42, B: 10}, []byte{0x8, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
}

func TestInfNotEnumIgnoresEnumRegistration(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	// Cannot decode because value of A is not set upon decoding and
	// it is not enum (because it is not yet registered)
	type WithEnum structWithField[InfEnum1]
	bcs.TestDecodeErr[WithEnum](t, &WithEnum{A: 42})

	// Can decode because although value of A is not enum (because it is not yet registered),
	// it is not set upon decoding
	vEnc := bcs.MustMarshal(&WithEnum{A: 42})
	vDec := bcs.MustUnmarshalInto(vEnc, &WithEnum{A: 0})
	require.Equal(t, &WithEnum{A: 42}, vDec)

	// Registering enum
	bcs.RegisterEnumType2[InfEnum1, int, string]()

	// Can encode/decode because value of A is now enum
	bcs.TestCodecAndBytes(t, &WithEnum{A: 42}, []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	// Cannot decode because although InfEnum1 is now registered as enum,
	// field A is marked as "not_enum".
	type WithNonEnum structWithNotEnumField[InfEnum1]
	bcs.TestDecodeErr[WithNonEnum](t, &WithNonEnum{A: 42})

	// But we can decode because if we preset value before decoding
	vEncNotEnum := bcs.MustMarshal(&WithNonEnum{A: 42})
	vDecNotEnum := bcs.MustUnmarshalInto(vEncNotEnum, &WithNonEnum{A: 0})
	require.Equal(t, &WithNonEnum{A: 42}, vDecNotEnum)
}

func TestInfEnumNone(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	require.Panics(t, func() {
		bcs.RegisterEnumType3[InfEnumWithMethods, bcs.None, int, EnumVariant1]()
	})

	bcs.RegisterEnumType2[InfEnumWithMethods, bcs.None, EnumVariant1]()

	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnumWithMethods](EnumVariant1{A: 42}), []byte{0x1, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, lo.ToPtr[InfEnumWithMethods](nil), []byte{0x0})

	type WithEnum structWithField[InfEnumWithMethods]
	bcs.RegisterEnumType3[InfEnum1, bcs.None, int, string]()
	bcs.TestCodecAndBytes(t, WithEnum{A: nil}, []byte{0x0})
	bcs.TestCodecAndBytes(t, WithEnum{A: EnumVariant1{A: 42}}, []byte{0x1, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	type WithOptionalInf structWithOptionalField[any]
	bcs.TestCodecAndBytes(t, WithOptionalInf{A: nil}, []byte{0x0})

	type WithInf structWithField[any]
	bcs.TestEncodeErr(t, WithInf{A: nil})
}
