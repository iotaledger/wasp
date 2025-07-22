package iotago_test

import (
	"testing"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
)

func TestBCS(t *testing.T) {
	bcs.TestCodecAndHash(
		t, iotago.TransferIota{
			Recipient: *iotago.AddressFromArray([32]byte{1, 2, 3}),
			Amount:    lo.ToPtr[uint64](123),
		},
		"9caad547c2aa",
	)

	pt := iotago.ProgrammableTransaction{
		Inputs: []iotago.CallArg{
			{
				// Pure: &[]byte{1, 2, 3},
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
	}

	bcs.TestCodec(t, pt)

	pt.Inputs[0].Object.SharedObject.Id = iotatest.TestAddress
	pt.Commands[0].MoveCall.Package = iotatest.TestAddress
	bcs.TestCodecAndHash(t, pt, "98d300b7eb13")
}
