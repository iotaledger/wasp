package iscmoveclient

import (
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func PTBCreateAndSendRequest(
	ptb *sui.ProgrammableTransactionBuilder,
	packageID sui.PackageID,
	anchorID sui.ObjectID,
	argAssetsBag sui.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	argAllowance sui.Argument,
	onchainGasBudget uint64,
) *sui.ProgrammableTransactionBuilder {
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(anchorID),
					argAssetsBag,
					ptb.MustPure(iscContractHname),
					ptb.MustPure(iscFunctionHname),
					ptb.MustPure(args),
					argAllowance,
					ptb.MustPure(onchainGasBudget),
				},
			},
		},
	)
	return ptb
}
