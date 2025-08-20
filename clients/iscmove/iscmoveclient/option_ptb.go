package iscmoveclient

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

func PTBOptionSome(
	ptb *iotago.ProgrammableTransactionBuilder,
	objTypeTag iotago.TypeTag,
	objRef *iotago.ObjectRef, // must be ImmOrOwnedObject
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.IotaPackageIDMoveStdlib,
				Module:        "option",
				Function:      "some",
				TypeArguments: []iotago.TypeTag{objTypeTag},
				Arguments: []iotago.Argument{
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: objRef}),
				},
			},
		},
	)
	return ptb
}

func PTBOptionSomeIotaCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
	objRef *iotago.ObjectRef, // must be ImmOrOwnedObject
) *iotago.ProgrammableTransactionBuilder {
	return PTBOptionSome(ptb, *iotago.MustTypeTagFromString("0x2::coin::Coin<0x2::iota::IOTA>"), objRef)
}

func PTBOptionNoneIotaCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.IotaPackageIDMoveStdlib,
				Module:        "option",
				Function:      "none",
				TypeArguments: []iotago.TypeTag{*iotago.MustTypeTagFromString("0x2::coin::Coin<0x2::iota::IOTA>")},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	return ptb
}
