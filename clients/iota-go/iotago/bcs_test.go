package iotago_test

import (
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestBCS(t *testing.T) {
	testBCS(
		t, iotago.TransferIota{
			Recipient: *iotago.AddressFromArray([32]byte{1, 2, 3}),
			Amount:    lo.ToPtr[uint64](123),
		},
	)

	// fardream is crashing on this...
	// testBCS(t, iotago.TransferIota{
	// 	Recipient: *iotago.AddressFromArray([32]byte{1, 2, 3}),
	// })

	// fardream is unable to decode it, so just testing encoding
	testBCSEnc(
		t, iotago.ProgrammableTransaction{
			Inputs: []iotago.CallArg{
				{
					//Pure: &[]byte{1, 2, 3},
					Object: &iotago.ObjectArg{
						SharedObject: &iotago.SharedObjectArg{
							Id:                   iotatest.RandomAddress(),
							InitialSharedVersion: 13,
							Mutable:              true,
						},
					},
				},
			},
			Commands: []iotago.Command{
				{
					MoveCall: &iotago.ProgrammableMoveCall{
						Package:  iotatest.RandomAddress(),
						Module:   "aaa",
						Function: "bbb",
						TypeArguments: []iotago.TypeTag{
							{U32: &serialization.EmptyEnum{}},
							{Address: &serialization.EmptyEnum{}},
						},
						Arguments: []iotago.Argument{
							{
								Input: lo.ToPtr[uint16](10),
							},
							{
								NestedResult: &iotago.NestedResult{
									Cmd:    11,
									Result: 12,
								},
							},
						},
					},
				},
			},
		},
	)
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
