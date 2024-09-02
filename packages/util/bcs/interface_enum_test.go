package bcs_test

import (
	"reflect"
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
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
	t.Cleanup(func() {
		bcs.EnumTypes = make(map[reflect.Type][]reflect.Type)
	})

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

type structWithEnum[EnumType any] struct {
	A EnumType
}

func StructWithEnum[EnumType any](v EnumType) structWithEnum[EnumType] {
	return structWithEnum[EnumType]{A: v}
}

func TestBasicInterfaceEnumCodec(t *testing.T) {
	t.Cleanup(func() {
		bcs.EnumTypes = make(map[reflect.Type][]reflect.Type)
	})

	bcs.RegisterEnumType2[InfEnum1, int, string]()
	bcs.RegisterEnumType3[InfEnum2, int64, int32, int8]()
	bcs.RegisterEnumType2[InfEnumWithMethods, EnumVariant1, EnumVariant2]()

	vS := "foo"
	refEnumEnc := ref_bcs.MustMarshal(RefEnum1{B: &vS})
	require.NotEmpty(t, refEnumEnc)
	bcs.TestCodecAndBytesNoRef(t, StructWithEnum(InfEnum1(vS)), refEnumEnc)

	vI := int32(42)
	refEnumEnc = ref_bcs.MustMarshal(RefEnum2{B: &vI})
	require.NotEmpty(t, refEnumEnc)
	bcs.TestCodecAndBytesNoRef(t, StructWithEnum(InfEnum2(vI)), refEnumEnc)

	var e InfEnumWithMethods = EnumVariant1{A: 42}
	bcs.TestCodecAndBytesNoRef(t, StructWithEnum(e), []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, &e, []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	e = EnumVariant2{A: "bar"}
	bcs.TestCodecAndBytesNoRef(t, StructWithEnum(e), []byte{0x1, 0x3, 0x62, 0x61, 0x72})
	bcs.TestCodecAndBytesNoRef(t, &e, []byte{0x1, 0x3, 0x62, 0x61, 0x72})
}

func TestInterfaceEnumVariantWithCustomCodec(t *testing.T) {
	t.Cleanup(func() {
		bcs.EnumTypes = make(map[reflect.Type][]reflect.Type)
	})

	bcs.RegisterEnumType2[InfEnum1, WithCustomCodec, string]()

	bcs.TestCodecAndBytesNoRef(t, lo.ToPtr[InfEnum1](WithCustomCodec{}), []byte{0x0, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytesNoRef(t, lo.ToPtr[InfEnum1]("aaa"), []byte{0x1, 0x3, 0x61, 0x61, 0x61})
}

func TestIntEnumWithStructEnumVariant(t *testing.T) {
	t.Cleanup(func() {
		bcs.EnumTypes = make(map[reflect.Type][]reflect.Type)
	})

	bcs.RegisterEnumType3[InfEnum1, BasicStructEnum, WithCustomCodec, string]()

	bcs.TestCodecAndBytesNoRef(t, lo.ToPtr[InfEnum1](BasicStructEnum{A: lo.ToPtr[int32](10)}), []byte{0x0, 0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, lo.ToPtr[InfEnum1](BasicStructEnum{B: lo.ToPtr("aaa")}), []byte{0x0, 0x1, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, lo.ToPtr[InfEnum1](WithCustomCodec{}), []byte{0x1, 0x1, 0x2, 0x3})
}
