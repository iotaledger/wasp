package iscmoveclient

import (
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
)

func PTBOptionSome(
	ptb *sui2.ProgrammableTransactionBuilder,
	objTypeTag sui2.TypeTag,
	objRef *sui2.ObjectRef, // must be ImmOrOwnedObject
) *sui2.ProgrammableTransactionBuilder {
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       sui2.SuiPackageIdMoveStdlib,
				Module:        "option",
				Function:      "some",
				TypeArguments: []sui2.TypeTag{objTypeTag},
				Arguments: []sui2.Argument{
					ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: objRef}),
				},
			},
		},
	)
	return ptb
}

func PTBOptionSomeSuiCoin(
	ptb *sui2.ProgrammableTransactionBuilder,
	objRef *sui2.ObjectRef, // must be ImmOrOwnedObject
) *sui2.ProgrammableTransactionBuilder {
	return PTBOptionSome(ptb, *sui2.MustTypeTagFromString("0x2::coin::Coin<0x2::sui::SUI>"), objRef)
}

func PTBOptionNoneSuiCoin(
	ptb *sui2.ProgrammableTransactionBuilder,
) *sui2.ProgrammableTransactionBuilder {
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       sui2.SuiPackageIdMoveStdlib,
				Module:        "option",
				Function:      "none",
				TypeArguments: []sui2.TypeTag{*sui2.MustTypeTagFromString("0x2::coin::Coin<0x2::sui::SUI>")},
				Arguments:     []sui2.Argument{},
			},
		},
	)
	return ptb
}
