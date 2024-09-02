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
	bcs.TestCodecAndBytes(t, BasicStructEnum{A: lo.ToPtr[int32](10)}, []byte{0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, BasicStructEnum{B: lo.ToPtr("aaa")}, []byte{0x1, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, BasicStructEnum{C: lo.ToPtr([]byte{1, 2, 3})}, []byte{0x2, 0x3, 0x1, 0x2, 0x3})

	bcs.TestEncodeErr(t, BasicStructEnum{A: lo.ToPtr[int32](10), B: lo.ToPtr("aaa")})
	bcs.TestEncodeErr(t, BasicStructEnum{})
}

type StructEnumWithVariantWithCustomCodec struct {
	A *int32
	B *WithCustomCodec
}

func (StructEnumWithVariantWithCustomCodec) IsBcsEnum() {}

func TestStructEnumWithVariantWithCustomCodec(t *testing.T) {
	bcs.TestCodecAndBytesNoRef(t, StructEnumWithVariantWithCustomCodec{B: &WithCustomCodec{}}, []byte{0x1, 0x1, 0x2, 0x3})
}

type StructEnumWithEnumVariant struct {
	A *int32
	B *BasicStructEnum
}

func (StructEnumWithEnumVariant) IsBcsEnum() {}

func TestStructEnumWithEnumVariant(t *testing.T) {
	bcs.TestCodecAndBytes(t, StructEnumWithEnumVariant{B: &BasicStructEnum{A: lo.ToPtr[int32](10)}}, []byte{0x1, 0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, StructEnumWithEnumVariant{B: &BasicStructEnum{C: lo.ToPtr([]byte{1, 2, 3})}}, []byte{0x1, 0x2, 0x3, 0x1, 0x2, 0x3})
}

type StructEnumWithNullableVariants struct {
	A *int32
	B InfEnum1
	C []byte
	D map[string]int32
}

func (StructEnumWithNullableVariants) IsBcsEnum() {}

func TestStructEnumWithInfEnumVariant(t *testing.T) {
	t.Cleanup(func() {
		bcs.EnumTypes = make(map[reflect.Type][]reflect.Type)
	})

	bcs.RegisterEnumType2[InfEnum1, WithCustomCodec, string]()

	bcs.TestCodecAndBytesNoRef(t, StructEnumWithNullableVariants{B: WithCustomCodec{}}, []byte{0x1, 0x0, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytesNoRef(t, StructEnumWithNullableVariants{B: "aaa"}, []byte{0x1, 0x1, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytesNoRef(t, StructEnumWithNullableVariants{A: lo.ToPtr[int32](10)}, []byte{0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesNoRef(t, StructEnumWithNullableVariants{C: []byte{4, 5, 6}}, []byte{0x2, 0x3, 0x4, 0x5, 0x6})
	bcs.TestCodecAndBytesNoRef(t, StructEnumWithNullableVariants{D: map[string]int32{"a": 1, "b": 2}}, []byte{0x3, 0x2, 0x1, 0x61, 0x1, 0x0, 0x0, 0x0, 0x1, 0x62, 0x2, 0x0, 0x0, 0x0})
}
