package iscmoveclient

import (
	iotago "github.com/iotaledger/wasp/clients/iota-go/iotago"
)

func PTBOptionSome(
	ptb *iotago.ProgrammableTransactionBuilder,
	objTypeTag iotago.TypeTag,
	objRef *iotago.ObjectRef, // must be ImmOrOwnedObject
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.SuiPackageIdMoveStdlib,
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

func PTBOptionSomeSuiCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
	objRef *iotago.ObjectRef, // must be ImmOrOwnedObject
) *iotago.ProgrammableTransactionBuilder {
	return PTBOptionSome(ptb, *iotago.MustTypeTagFromString("0x2::coin::Coin<0x2::iotago::SUI>"), objRef)
}

func PTBOptionNoneSuiCoin(
	ptb *iotago.ProgrammableTransactionBuilder,
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       iotago.SuiPackageIdMoveStdlib,
				Module:        "option",
				Function:      "none",
				TypeArguments: []iotago.TypeTag{*iotago.MustTypeTagFromString("0x2::coin::Coin<0x2::iotago::SUI>")},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	return ptb
}
