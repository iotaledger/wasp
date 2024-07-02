package sui_test

import (
	"testing"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/serialization"
	"github.com/stretchr/testify/require"
)

func TestTypeTagEncoding(t *testing.T) {
	typeTagU64 := sui.TypeTag{U64: &serialization.EmptyEnum{}}
	typeTagMarshaled, err := bcs.Marshal(typeTagU64)
	require.NoError(t, err)
	require.Equal(t, []byte{2}, typeTagMarshaled)

	typeTagStruct := sui.TypeTag{Struct: &sui.StructTag{
		Address: sui.MustObjectIDFromHex("0x2eeb551107032ae860d76661f3f4573dd0f8c701116137e6525dcd95d4f8e58"),
		Module:  "testcoin",
		Name:    "TESTCOIN",
	}}
	typeTagStructMarshaled, err := bcs.Marshal(typeTagStruct)
	require.NoError(t, err)
	var structTag sui.TypeTag
	_, err = bcs.Unmarshal(typeTagStructMarshaled, &structTag)
	require.NoError(t, err)
}

func TestStructTagEncoding(t *testing.T) {
	s1 := `0x2::foo::bar<0x3::baz::qux<0x4::nested::result, 0x5::funny::other>, bool>`
	structTag, err := sui.StructTagFromString(s1)
	require.NoError(t, err)

	require.Equal(t, sui.MustObjectIDFromHex("0x2"), structTag.Address)
	require.Equal(t, sui.Identifier("foo"), structTag.Module)
	require.Equal(t, sui.Identifier("bar"), structTag.Name)

	typeParam0 := structTag.TypeParams[0].Struct
	require.Equal(t, sui.MustObjectIDFromHex("0x3"), typeParam0.Address)
	require.Equal(t, sui.Identifier("baz"), typeParam0.Module)
	require.Equal(t, sui.Identifier("qux"), typeParam0.Name)
	typeParam00 := structTag.TypeParams[0].Struct.TypeParams[0].Struct
	require.Equal(t, sui.MustObjectIDFromHex("0x4"), typeParam00.Address)
	require.Equal(t, sui.Identifier("nested"), typeParam00.Module)
	require.Equal(t, sui.Identifier("result"), typeParam00.Name)

	typeParam01 := structTag.TypeParams[0].Struct.TypeParams[1].Struct
	require.Equal(t, sui.MustObjectIDFromHex("0x5"), typeParam01.Address)
	require.Equal(t, sui.Identifier("funny"), typeParam01.Module)
	require.Equal(t, sui.Identifier("other"), typeParam01.Name)

	require.NotNil(t, structTag.TypeParams[1].Bool)
}
