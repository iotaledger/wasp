package bcs_test

import (
	"reflect"
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
)

type BasicStructEnum struct {
	A *int32
	B *string
	C *[]byte
}

func (BasicStructEnum) IsBcsEnum() {}

func TestBasicStructEnum(t *testing.T) {
	testCodec(t, BasicStructEnum{A: lo.ToPtr[int32](10)}, []byte{0x0, 0xa, 0x0, 0x0, 0x0})
	testCodec(t, BasicStructEnum{B: lo.ToPtr("aaa")}, []byte{0x1, 0x3, 0x61, 0x61, 0x61})
	testCodec(t, BasicStructEnum{C: lo.ToPtr([]byte{1, 2, 3})}, []byte{0x2, 0x3, 0x1, 0x2, 0x3})

	testCodecErr(t, BasicStructEnum{A: lo.ToPtr[int32](10), B: lo.ToPtr("aaa")})
	testCodecErr(t, BasicStructEnum{})
}

type StructEnumWithVariantWithCustomCodec struct {
	A *int32
	B *WithCustomCodec
}

func (StructEnumWithVariantWithCustomCodec) IsBcsEnum() {}

func TestStructEnumWithVariantWithCustomCodec(t *testing.T) {
	testCodecNoRef(t, StructEnumWithVariantWithCustomCodec{B: &WithCustomCodec{}}, []byte{0x1, 0x1, 0x2, 0x3})
}

type StructEnumWithEnumVariant struct {
	A *int32
	B *BasicStructEnum
}

func (StructEnumWithEnumVariant) IsBcsEnum() {}

func TestStructEnumWithEnumVariant(t *testing.T) {
	testCodec(t, StructEnumWithEnumVariant{B: &BasicStructEnum{A: lo.ToPtr[int32](10)}}, []byte{0x1, 0x0, 0xa, 0x0, 0x0, 0x0})
	testCodec(t, StructEnumWithEnumVariant{B: &BasicStructEnum{C: lo.ToPtr([]byte{1, 2, 3})}}, []byte{0x1, 0x2, 0x3, 0x1, 0x2, 0x3})
}

type StructEnumWithInfEnumVariant struct {
	A *int32
	B InfEnum1
}

func (StructEnumWithInfEnumVariant) IsBcsEnum() {}

func TestStructEnumWithInfEnumVariant(t *testing.T) {
	t.Cleanup(func() {
		bcs.EnumTypes = make(map[reflect.Type][]reflect.Type)
	})

	bcs.RegisterEnumType2[InfEnum1, WithCustomCodec, string]()

	testCodecNoRef(t, StructEnumWithInfEnumVariant{B: WithCustomCodec{}}, []byte{0x1, 0x0, 0x1, 0x2, 0x3})
	testCodecNoRef(t, StructEnumWithInfEnumVariant{B: "aaa"}, []byte{0x1, 0x1, 0x3, 0x61, 0x61, 0x61})
	testCodecNoRef(t, StructEnumWithInfEnumVariant{A: lo.ToPtr[int32](10)}, []byte{0x0, 0xa, 0x0, 0x0, 0x0})
}
