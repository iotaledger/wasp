package pkg

import (
	"context"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func CreatePool(
	suiClient *suiclient.Client,
	signer suisigner.Signer,
	swapPackageID *sui.PackageID,
	testcoinID *sui.ObjectID,
	testCoin *suijsonrpc.Coin,
	suiCoins []*suijsonrpc.Coin,
) *sui.ObjectID {
	ptb := sui.NewProgrammableTransactionBuilder()

	arg0 := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: testCoin.Ref()})
	arg1 := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: suiCoins[0].Ref()})
	arg2 := ptb.MustPure(uint64(3))

	lspArg := ptb.Command(sui.Command{
		MoveCall: &sui.ProgrammableMoveCall{
			Package:  swapPackageID,
			Module:   "swap",
			Function: "create_pool",
			TypeArguments: []sui.TypeTag{{Struct: &sui.StructTag{
				Address: testcoinID,
				Module:  "testcoin",
				Name:    "TESTCOIN",
			}}},
			Arguments: []sui.Argument{arg0, arg1, arg2},
		}},
	)
	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{lspArg},
				Address: ptb.MustPure(signer.Address()),
			},
		},
	)
	pt := ptb.Finish()
	txData := sui.NewProgrammable(
		signer.Address(),
		pt,
		[]*sui.ObjectRef{suiCoins[1].Ref()},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(txData)
	if err != nil {
		panic(err)
	}

	txnResponse, err := suiClient.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}
	for _, change := range txnResponse.ObjectChanges {
		if change.Data.Created != nil {
			resource, err := sui.NewResourceType(change.Data.Created.ObjectType)
			if err != nil {
				panic(err)
			}
			if resource.Contains(nil, "swap", "Pool") {
				return &change.Data.Created.ObjectID
			}
		}
	}

	return nil
}
