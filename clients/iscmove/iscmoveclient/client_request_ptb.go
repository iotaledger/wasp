package iscmoveclient

import (
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func NewCreateAndSendRequestPTB(
	packageID sui.PackageID,
	anchorID sui.ObjectID,
	assetsBagRef *sui.ObjectRef,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(anchorID),
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustPure(iscContractHname),
					ptb.MustPure(iscFunctionHname),
					ptb.MustPure(args),
				},
			},
		},
	)

	return ptb.Finish()
}
