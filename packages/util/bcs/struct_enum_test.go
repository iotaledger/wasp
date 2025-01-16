package bcs_test

import (
	"testing"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

type BasicStructEnum struct {
	A *int32
	B *string
	C *[]byte
}

func (BasicStructEnum) IsBcsEnum() {}

func TestBasicStructEnum(t *testing.T) {
	bcs.TestCodecAndBytesVsRef(t, BasicStructEnum{A: lo.ToPtr[int32](10)}, []byte{0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, BasicStructEnum{B: lo.ToPtr("aaa")}, []byte{0x1, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytesVsRef(t, BasicStructEnum{C: lo.ToPtr([]byte{1, 2, 3})}, []byte{0x2, 0x3, 0x1, 0x2, 0x3})

	bcs.TestEncodeErr(t, BasicStructEnum{A: lo.ToPtr[int32](10), B: lo.ToPtr("aaa")})
	bcs.TestEncodeErr(t, BasicStructEnum{})
}

type StructEnumWithVariantWithCustomCodec struct {
	A *int32
	B *WithCustomCodec
}

func (StructEnumWithVariantWithCustomCodec) IsBcsEnum() {}

func TestStructEnumWithVariantWithCustomCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, StructEnumWithVariantWithCustomCodec{B: &WithCustomCodec{}}, []byte{0x1, 0x1, 0x2, 0x3})
}

type StructEnumWithEnumVariant struct {
	A *int32
	B *BasicStructEnum
}

func (StructEnumWithEnumVariant) IsBcsEnum() {}

func TestStructEnumWithEnumVariant(t *testing.T) {
	bcs.TestCodecAndBytesVsRef(t, StructEnumWithEnumVariant{B: &BasicStructEnum{A: lo.ToPtr[int32](10)}}, []byte{0x1, 0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytesVsRef(t, StructEnumWithEnumVariant{B: &BasicStructEnum{C: lo.ToPtr([]byte{1, 2, 3})}}, []byte{0x1, 0x2, 0x3, 0x1, 0x2, 0x3})
}

type StructEnumWithNullableVariants struct {
	A *int32
	B InfEnum1
	C []byte
	D map[string]int32
}

func (StructEnumWithNullableVariants) IsBcsEnum() {}

func TestStructEnumWithInfEnumVariant(t *testing.T) {
	t.Cleanup(func() { maps.Clear(bcs.EnumTypes) })

	bcs.RegisterEnumType2[InfEnum1, WithCustomCodec, string]()

	bcs.TestCodecAndBytes(t, StructEnumWithNullableVariants{B: WithCustomCodec{}}, []byte{0x1, 0x0, 0x1, 0x2, 0x3})
	bcs.TestCodecAndBytes(t, StructEnumWithNullableVariants{B: "aaa"}, []byte{0x1, 0x1, 0x3, 0x61, 0x61, 0x61})
	bcs.TestCodecAndBytes(t, StructEnumWithNullableVariants{A: lo.ToPtr[int32](10)}, []byte{0x0, 0xa, 0x0, 0x0, 0x0})
	bcs.TestCodecAndBytes(t, StructEnumWithNullableVariants{C: []byte{4, 5, 6}}, []byte{0x2, 0x3, 0x4, 0x5, 0x6})
	bcs.TestCodecAndBytes(t, StructEnumWithNullableVariants{D: map[string]int32{"a": 1, "b": 2}}, []byte{0x3, 0x2, 0x1, 0x61, 0x1, 0x0, 0x0, 0x0, 0x1, 0x62, 0x2, 0x0, 0x0, 0x0})

	bcs.TestEncodeErr(t, StructEnumWithNullableVariants{B: 123})
}
