package iscmove

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/types"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func NewCreateAndSendRequestPTB(
	packageID sui.PackageID,
	anchorID sui.ObjectID,
	assetsBagRef *sui.ObjectRef,
	iscContractHname isc.Hname,
	iscFunctionHname isc.Hname,
	args [][]byte,
) sui.ProgrammableTransaction {
	ptb := sui.NewProgrammableTransactionBuilder()

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        types.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(anchorID),
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustPure(uint32(iscContractHname)),
					ptb.MustPure(uint32(iscFunctionHname)),
					ptb.MustPure(args),
				},
			},
		},
	)

	return ptb.Finish()
}
