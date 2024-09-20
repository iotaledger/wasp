package iscmoveclient

import (
	"github.com/iotaledger/wasp/sui-go/sui"
)

func PTBOptionSome(
	ptb *sui.ProgrammableTransactionBuilder,
	objTypeTag sui.TypeTag,
	objRef *sui.ObjectRef, // must be ImmOrOwnedObject
) *sui.ProgrammableTransactionBuilder {
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       sui.SuiPackageIdMoveStdlib,
				Module:        "option",
				Function:      "some",
				TypeArguments: []sui.TypeTag{objTypeTag},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: objRef}),
				},
			},
		},
	)
	return ptb
}

func PTBOptionSomeSuiCoin(
	ptb *sui.ProgrammableTransactionBuilder,
	objRef *sui.ObjectRef, // must be ImmOrOwnedObject
) *sui.ProgrammableTransactionBuilder {
	return PTBOptionSome(ptb, *sui.MustTypeTagFromString("0x2::coin::Coin<0x2::sui::SUI>"), objRef)
}

func PTBOptionNoneSuiCoin(
	ptb *sui.ProgrammableTransactionBuilder,
) *sui.ProgrammableTransactionBuilder {
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       sui.SuiPackageIdMoveStdlib,
				Module:        "option",
				Function:      "none",
				TypeArguments: []sui.TypeTag{*sui.MustTypeTagFromString("0x2::coin::Coin<0x2::sui::SUI>")},
				Arguments:     []sui.Argument{},
			},
		},
	)
	return ptb
}
