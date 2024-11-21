package iscmoveclient

import (
	"slices"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
)

func PTBCreateAndSendRequest(
	ptb *iotago.ProgrammableTransactionBuilder,
	packageID iotago.PackageID,
	anchorID iotago.ObjectID,
	argAssetsBag iotago.Argument,
	msg *iscmove.Message,
	// the order of the allowance will be reversed after processed by move
	allowance *iscmove.Assets,
	onchainGasBudget uint64,
) *iotago.ProgrammableTransactionBuilder {
	allowanceCoinTypes := lo.Keys(allowance.Coins)
	slices.Sort(allowanceCoinTypes)
	allowanceBalances := lo.Map(allowanceCoinTypes, func(t string, i int) uint64 {
		return allowance.Coins[t].Uint64()
	})

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
					ptb.MustForceSeparatePure(msg.Contract),
					ptb.MustForceSeparatePure(msg.Function),
					ptb.MustForceSeparatePure(msg.Args),
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
