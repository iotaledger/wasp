package iscmove

import (
	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/sui-go/sui"
)

func NewCreateAndSendRequestPTB(packageID sui.PackageID, anchorID sui.ObjectID, assetsBagRef *sui.ObjectRef, iscContractName string, iscFunctionName string, args [][]byte) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(anchorID),
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustPure(&bcs.Option[string]{Some: iscContractName}),
					ptb.MustPure(&bcs.Option[string]{Some: iscFunctionName}),
					ptb.MustPure(&bcs.Option[[][]byte]{Some: args}),
				},
			},
		},
	)

	return ptb.Finish()
}
