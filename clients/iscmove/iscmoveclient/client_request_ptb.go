package iscmoveclient

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
)

func PTBCreateAndSendRequest(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	anchorID iotago.ObjectID,
	argAssetsBag iotago.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	// the order of the allowance will be reversed after processed by
	allowanceArray []iscmove.CoinAllowance,
	onchainGasBudget uint64,
) *iotago.ProgrammableTransactionBuilder {
	allowanceCoinTypes := make([]iotajsonrpc.CoinType, len(allowanceArray))
	allowanceBalances := make([]uint64, len(allowanceArray))
	for i, val := range allowanceArray {
		allowanceCoinTypes[i] = val.CoinType
		allowanceBalances[i] = val.Balance
	}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
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
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	anchorID iotago.ObjectID,
	argAssetsBag iotago.Argument,
	iscContractHname uint32,
	iscFunctionHname uint32,
	args [][]byte,
	allowanceCointypes iotago.Argument,
	allowanceBalances iotago.Argument,
	onchainGasBudget uint64,
) *iotago.ProgrammableTransactionBuilder {
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        iscmove.RequestModuleName,
				Function:      "create_and_send_request",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
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
