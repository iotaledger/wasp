package iotago_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
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
		s1 := "0x2::foo::bar<0x3::baz::qux<0x4::nested::result, 0x5::funny::other, 0x6::more::another>, bool>"
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

		typeParam02 := structTag.TypeParams[0].Struct.TypeParams[2].Struct
		require.Equal(t, iotago.MustObjectIDFromHex("0x6"), typeParam02.Address)
		require.Equal(t, iotago.Identifier("more"), typeParam02.Module)
		require.Equal(t, iotago.Identifier("another"), typeParam02.Name)

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

func TestStructTagParsingEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Whitespace tests
		{
			name:  "whitespace after <",
			input: "0x1::module::name< bool>",
		},
		{
			name:  "whitespace before >",
			input: "0x1::module::name<bool >",
		},
		{
			name:  "whitespace around comma",
			input: "0x1::module::name<0x2::mod::test , bool>",
		},
		{
			name:  "multiple spaces",
			input: "0x1::module::name<  bool  ,   u64  >",
		},
		{
			name:  "tabs and spaces",
			input: "0x1::module::name<\tbool\t,\tu64\t>",
		},

		// Address format tests
		{
			name:  "short address",
			input: "0x1::module::name",
		},
		{
			name:  "long address",
			input: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef::module::name",
		},
		{
			name:  "uppercase hex",
			input: "0x1234ABCD::module::name",
		},
		{
			name:  "mixed case hex",
			input: "0x1234aBcD::module::name",
		},

		// Primitive type tests
		{
			name:  "all primitive types",
			input: "0x1::module::name<bool, u8, u16, u32, u64, u128, u256, address, signer>",
		},

		// Error cases
		{
			name:    "missing 0x prefix",
			input:   "1::module::name",
			wantErr: true,
			errMsg:  "unexpected character '1' at position 0",
		},
		{
			name:    "invalid hex",
			input:   "0xgg::module::name",
			wantErr: true,
			errMsg:  `invalid address`,
		},
		{
			name:    "too long address",
			input:   "0x" + strings.Repeat("1", 65) + "::module::name",
			wantErr: true,
			errMsg:  "invalid address",
		},
		{
			name:    "missing double colon",
			input:   "0x1:module::name",
			wantErr: true,
			errMsg:  `unexpected character ':' at position 3`,
		},
		{
			name:    "invalid identifier",
			input:   "0x1::123::name",
			wantErr: true,
			errMsg:  "unexpected character",
		},
		{
			name:    "empty type params",
			input:   "0x1::module::name<>",
			wantErr: true,
			errMsg:  `expected primitive type or address at position 18, got '>'`,
		},
		{
			name:    "unmatched <",
			input:   "0x1::module::name<bool",
			wantErr: true,
			errMsg:  `expected ',' or '>' at position 22, got ''`,
		},
		{
			name:    "unmatched >",
			input:   "0x1::module::name<bool>>",
			wantErr: true,
			errMsg:  "expected end of input",
		},
		{
			name:    "trailing comma",
			input:   "0x1::module::name<bool,>",
			wantErr: true,
			errMsg:  "expected primitive type or address at position 23, got '>'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := iotago.StructTagFromString(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

func TestStructTagRoundTrip(t *testing.T) {
	tests := []string{
		"0x1::module::name",
		"0x2::foo::bar<bool>",
		"0x3::test::deep<0x4::inner::struct<u64, bool>, u32>",
		"0x2::foo::bar<0x3::baz::qux<0x4::nested::result, 0x5::funny::other, 0x6::more::another>, bool>",
		"0xa::very_long_module_name::VeryLongStructName<0xabcd::another::Type>",
		"0x123::mod::name<bool, u8, u16, u32, u64, u128, u256, address, signer>",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Parse the input
			parsed, err := iotago.StructTagFromString(input)
			require.NoError(t, err)

			// Convert back to string
			strResult := parsed.String()

			// Parse again to ensure consistency
			reparsed, err := iotago.StructTagFromString(strResult)
			require.NoError(t, err)

			// Compare the structs field by field
			require.Equal(t, parsed.Address, reparsed.Address)
			require.Equal(t, parsed.Module, reparsed.Module)
			require.Equal(t, parsed.Name, reparsed.Name)
			require.Equal(t, len(parsed.TypeParams), len(reparsed.TypeParams))

			// Deep comparison of type params would be more complex,
			// but the String() comparison should be sufficient for the round-trip test
			require.Equal(t, parsed.String(), reparsed.String())
		})
	}
}

func TestStructTagWhitespaceVariations(t *testing.T) {
	// All these variations should parse to the same canonical form
	canonical := "0x1::module::name<bool, u64>"
	variations := []string{
		"0x1::module::name< bool, u64>",
		"0x1::module::name<bool , u64>",
		"0x1::module::name<bool, u64 >",
		"0x1::module::name< bool , u64 >",
		"0x1::module::name<  bool  ,  u64  >",
		"0x1::module::name<\tbool\t,\tu64\t>",
		"0x1::module::name<\tbool,u64\t>",
	}

	// Parse canonical form
	canonicalParsed, err := iotago.StructTagFromString(canonical)
	require.NoError(t, err)
	canonicalStr := canonicalParsed.String()

	for _, variation := range variations {
		t.Run("variation: "+variation, func(t *testing.T) {
			parsed, err := iotago.StructTagFromString(variation)
			require.NoError(t, err)

			// Should produce the same canonical string representation
			require.Equal(t, canonicalStr, parsed.String())

			// Should have the same structure
			require.Equal(t, canonicalParsed.Address, parsed.Address)
			require.Equal(t, canonicalParsed.Module, parsed.Module)
			require.Equal(t, canonicalParsed.Name, parsed.Name)
			require.Equal(t, len(canonicalParsed.TypeParams), len(parsed.TypeParams))
		})
	}
}
