package iscmoveclient

import (
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func PTBCreateAndSendRequest(
	ptb *sui.ProgrammableTransactionBuilder,
	packageID sui.PackageID,
	anchorID sui.ObjectID,
	argAssetsBag sui.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	// the order of the allowance will be reversed after processed by
	allowanceArray []iscmove.CoinAllowance,
	onchainGasBudget uint64,
) *sui.ProgrammableTransactionBuilder {
	allowanceCoinTypes := make([]suijsonrpc.CoinType, len(allowanceArray))
	allowanceBalances := make([]uint64, len(allowanceArray))
	for i, val := range allowanceArray {
		allowanceCoinTypes[i] = val.CoinType
		allowanceBalances[i] = val.Balance
	}
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustForceSeparatePure(anchorID),
					argAssetsBag,
					ptb.MustForceSeparatePure(iscContractHname),
					ptb.MustForceSeparatePure(iscFunctionHname),
					ptb.MustForceSeparatePure(args),
					ptb.MustForceSeparatePure(allowanceCoinTypes),
					ptb.MustForceSeparatePure(allowanceBalances),
					ptb.MustForceSeparatePure(onchainGasBudget),
				},
			},
		},
	)
	return ptb
}

func PTBCreateAndSendCrossRequest(
	ptb *sui.ProgrammableTransactionBuilder,
	packageID sui.PackageID,
	anchorID sui.ObjectID,
	argAssetsBag sui.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	allowanceCointypes sui.Argument,
	allowanceBalances sui.Argument,
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
					ptb.MustForceSeparatePure(anchorID),
					argAssetsBag,
					ptb.MustForceSeparatePure(iscContractHname),
					ptb.MustForceSeparatePure(iscFunctionHname),
					ptb.MustForceSeparatePure(args),
					allowanceCointypes,
					allowanceBalances,
					ptb.MustForceSeparatePure(onchainGasBudget),
				},
			},
		},
	)
	return ptb
}
