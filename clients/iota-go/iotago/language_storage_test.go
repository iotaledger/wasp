package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
)

func TestTypeTagEncoding(t *testing.T) {
	typeTagU64 := iotago.TypeTag{U64: &serialization.EmptyEnum{}}
	bcs.TestCodecAndBytes(t, typeTagU64, []byte{2})

	typeTagStruct := iotago.TypeTag{
		Struct: &iotago.StructTag{
			Address:    iotago.MustObjectIDFromHex("0x2eeb551107032ae860d76661f3f4573dd0f8c701116137e6525dcd95d4f8e58"),
			Module:     "testcoin",
			Name:       "TESTCOIN",
			TypeParams: []iotago.TypeTag{},
		},
	}
	bcs.TestCodecAndHash(t, &typeTagStruct, "ae8b80292a4d")
}

func TestTypeTagString(t *testing.T) {
	testcases := []string{
		"u8",
		"u16",
		"u32",
		"u64",
		"u128",
		"u256",
		"bool",
		"address",
		"vector<u8>",
		"0x0000000000000000000000000000000000000000000000000000000000000002::object::UID",
		"0x0000000000000000000000000000000000000000000000000000000000000002::coin::Coin" +
			"<0x0000000000000000000000000000000000000000000000000000000000000002::iota::IOTA>",
	}
	for _, testcase := range testcases {
		typetag, err := iotago.TypeTagFromString(testcase)
		require.NoError(t, err)
		require.Equal(t, testcase, typetag.String())
	}
}

func TestStructTagEncoding(t *testing.T) {
	{
		s1 := "0x2::foo::bar<0x3::baz::qux<0x4::nested::result, 0x5::funny::other>, bool>"
		structTag, err := iotago.StructTagFromString(s1)
		require.NoError(t, err)

		require.Equal(t, iotago.MustObjectIDFromHex("0x2"), structTag.Address)
		require.Equal(t, iotago.Identifier("foo"), structTag.Module)
		require.Equal(t, iotago.Identifier("bar"), structTag.Name)

		typeParam0 := structTag.TypeParams[0].Struct
		require.Equal(t, iotago.MustObjectIDFromHex("0x3"), typeParam0.Address)
		require.Equal(t, iotago.Identifier("baz"), typeParam0.Module)
		require.Equal(t, iotago.Identifier("qux"), typeParam0.Name)
		typeParam00 := structTag.TypeParams[0].Struct.TypeParams[0].Struct
		require.Equal(t, iotago.MustObjectIDFromHex("0x4"), typeParam00.Address)
		require.Equal(t, iotago.Identifier("nested"), typeParam00.Module)
		require.Equal(t, iotago.Identifier("result"), typeParam00.Name)

		typeParam01 := structTag.TypeParams[0].Struct.TypeParams[1].Struct
		require.Equal(t, iotago.MustObjectIDFromHex("0x5"), typeParam01.Address)
		require.Equal(t, iotago.Identifier("funny"), typeParam01.Module)
		require.Equal(t, iotago.Identifier("other"), typeParam01.Name)

		require.NotNil(t, structTag.TypeParams[1].Bool)
	}

	{
		s2 := "0x2::coin::Coin<0x2e1df076b986a33cc40a809c44c96e35b48d0ab36da48e23c26ec776e6be3c4b::testcoin::TESTCOIN>"
		structTag, err := iotago.StructTagFromString(s2)
		require.NoError(t, err)

		require.Equal(t, iotago.MustObjectIDFromHex("0x2"), structTag.Address)
		require.Equal(t, iotago.Identifier("coin"), structTag.Module)
		require.Equal(t, iotago.Identifier("Coin"), structTag.Name)

		typeParam0 := structTag.TypeParams[0].Struct
		require.Equal(
			t,
			iotago.MustObjectIDFromHex("0x2e1df076b986a33cc40a809c44c96e35b48d0ab36da48e23c26ec776e6be3c4b"),
			typeParam0.Address,
		)
		require.Equal(t, iotago.Identifier("testcoin"), typeParam0.Module)
		require.Equal(t, iotago.Identifier("TESTCOIN"), typeParam0.Name)
	}
}
