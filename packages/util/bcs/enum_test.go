package bcs_test

import (
	"reflect"
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/stretchr/testify/require"
)

type EmptyEnum1 interface{}

type EmptyEnum2 interface{}

type EnumWithMethods interface {
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
		bcs.RegisterEnumType2[EmptyEnum1, int, EmptyEnum2]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumType2[EnumWithMethods, int, string]()
	})

	require.Panics(t, func() {
		bcs.RegisterEnumType2[EnumWithMethods, EnumVariant1, string]()
	})

	bcs.RegisterEnumType2[EmptyEnum1, int, string]()
	bcs.RegisterEnumType3[EmptyEnum2, int64, int32, int8]()
	bcs.RegisterEnumType2[EnumWithMethods, EnumVariant1, EnumVariant2]()
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

func TestBasicEnumCodec(t *testing.T) {
	bcs.RegisterEnumType2[EmptyEnum1, int, string]()
	bcs.RegisterEnumType3[EmptyEnum2, int64, int32, int8]()
	bcs.RegisterEnumType2[EnumWithMethods, EnumVariant1, EnumVariant2]()

	vS := "foo"
	refEnumEnc := ref_bcs.MustMarshal(RefEnum1{B: &vS})
	require.NotEmpty(t, refEnumEnc)
	testCodecNoRef(t, StructWithEnum(EmptyEnum1(vS)), refEnumEnc)

	vI := int32(42)
	refEnumEnc = ref_bcs.MustMarshal(RefEnum2{B: &vI})
	require.NotEmpty(t, refEnumEnc)
	testCodecNoRef(t, StructWithEnum(EmptyEnum2(vI)), refEnumEnc)

	var e EnumWithMethods = EnumVariant1{A: 42}
	testCodecNoRef(t, StructWithEnum(e), []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	testCodecNoRef(t, &e, []byte{0x0, 0x2a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	e = EnumVariant2{A: "bar"}
	testCodecNoRef(t, StructWithEnum(e), []byte{0x1, 0x3, 0x62, 0x61, 0x72})
	testCodecNoRef(t, &e, []byte{0x1, 0x3, 0x62, 0x61, 0x72})
}
