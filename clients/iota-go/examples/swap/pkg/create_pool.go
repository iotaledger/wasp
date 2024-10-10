package pkg

import (
	"context"

	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func CreatePool(
	suiClient *suiclient2.Client,
	signer suisigner.Signer,
	swapPackageID *sui2.PackageID,
	testcoinID *sui2.ObjectID,
	testCoin *suijsonrpc2.Coin,
	suiCoins []*suijsonrpc2.Coin,
) *sui2.ObjectID {
	ptb := sui2.NewProgrammableTransactionBuilder()

	arg0 := ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: testCoin.Ref()})
	arg1 := ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: suiCoins[0].Ref()})
	arg2 := ptb.MustPure(uint64(3))

	lspArg := ptb.Command(sui2.Command{
		MoveCall: &sui2.ProgrammableMoveCall{
			Package:  swapPackageID,
			Module:   "swap",
			Function: "create_pool",
			TypeArguments: []sui2.TypeTag{{Struct: &sui2.StructTag{
				Address: testcoinID,
				Module:  "testcoin",
				Name:    "TESTCOIN",
			}}},
			Arguments: []sui2.Argument{arg0, arg1, arg2},
		}},
	)
	ptb.Command(
		sui2.Command{
			TransferObjects: &sui2.ProgrammableTransferObjects{
				Objects: []sui2.Argument{lspArg},
				Address: ptb.MustPure(signer.Address()),
			},
		},
	)
	pt := ptb.Finish()
	txData := sui2.NewProgrammable(
		signer.Address(),
		pt,
		[]*sui2.ObjectRef{suiCoins[1].Ref()},
		suiclient2.DefaultGasBudget,
		suiclient2.DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&txData)
	if err != nil {
		panic(err)
	}

	txnResponse, err := suiClient.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txBytes,
		&suijsonrpc2.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}
	for _, change := range txnResponse.ObjectChanges {
		if change.Data.Created != nil {
			resource, err := sui2.NewResourceType(change.Data.Created.ObjectType)
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
