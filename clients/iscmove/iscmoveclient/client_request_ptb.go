package iscmoveclient

import (
	"github.com/iotaledger/wasp/clients/iscmove"
	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
)

func PTBCreateAndSendRequest(
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	anchorID sui2.ObjectID,
	argAssetsBag sui2.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	// the order of the allowance will be reversed after processed by
	allowanceArray []iscmove.CoinAllowance,
	onchainGasBudget uint64,
) *sui2.ProgrammableTransactionBuilder {
	allowanceCoinTypes := make([]suijsonrpc.CoinType, len(allowanceArray))
	allowanceBalances := make([]uint64, len(allowanceArray))
	for i, val := range allowanceArray {
		allowanceCoinTypes[i] = val.CoinType
		allowanceBalances[i] = val.Balance
	}
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
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
	ptb *sui2.ProgrammableTransactionBuilder,
	packageID sui2.PackageID,
	anchorID sui2.ObjectID,
	argAssetsBag sui2.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	allowanceCointypes sui2.Argument,
	allowanceBalances sui2.Argument,
	onchainGasBudget uint64,
) *sui2.ProgrammableTransactionBuilder {
	ptb.Command(
		sui2.Command{
			MoveCall: &sui2.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []sui2.TypeTag{},
				Arguments: []sui2.Argument{
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
