package sui_test

import (
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/serialization"
	"github.com/iotaledger/wasp/sui-go/sui/suitest"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestBCS(t *testing.T) {
	testBCS(t, sui.TransferSui{
		Recipient: *sui.AddressFromArray([32]byte{1, 2, 3}),
		Amount:    lo.ToPtr[uint64](123),
	})

	// fardream is crashing on this...
	// testBCS(t, sui.TransferSui{
	// 	Recipient: *sui.AddressFromArray([32]byte{1, 2, 3}),
	// })

	// fardream is unable to decode it, so just testing encoding
	testBCSEnc(t, sui.ProgrammableTransaction{
		Inputs: []sui.CallArg{
			{
				//Pure: &[]byte{1, 2, 3},
				Object: &sui.ObjectArg{
					SharedObject: &sui.SharedObjectArg{
						Id:                   suitest.RandomAddress(),
						InitialSharedVersion: 13,
						Mutable:              true,
					},
				},
			},
		},
		Commands: []sui.Command{
			{
				MoveCall: &sui.ProgrammableMoveCall{
					Package:  suitest.RandomAddress(),
					Module:   "aaa",
					Function: "bbb",
					TypeArguments: []sui.TypeTag{
						{U32: &serialization.EmptyEnum{}},
						{Address: &serialization.EmptyEnum{}},
					},
					Arguments: []sui.Argument{
						{
							Input: lo.ToPtr[uint16](10),
						},
						{
							NestedResult: &sui.NestedResult{
								Cmd:    11,
								Result: 12,
							},
						},
					},
				},
			},
		},
	})
}

func testBCS[V any](t *testing.T, v V) {
	enc := testBCSEnc(t, v)

	var refDec V
	_, err := ref_bcs.Unmarshal(enc, &refDec)
	require.NoError(t, err)

	dec := bcs.MustUnmarshal[V](enc)

	require.Equal(t, refDec, dec)
	require.Equal(t, dec, v)
}

func testBCSEnc[V any](t *testing.T, v V) []byte {
	refEnc := ref_bcs.MustMarshal(v)
	enc := bcs.MustMarshal(&v)
	require.Equal(t, refEnc, enc)

	return enc
}
