package pkg

import (
	"context"

	iotaclient2 "github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	iotago "github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func CreatePool(
	suiClient *iotaclient2.Client,
	signer iotasigner.Signer,
	swapPackageID *iotago.PackageID,
	testcoinID *iotago.ObjectID,
	testCoin *iotajsonrpc.Coin,
	suiCoins []*iotajsonrpc.Coin,
) *iotago.ObjectID {
	ptb := iotago.NewProgrammableTransactionBuilder()

	arg0 := ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: testCoin.Ref()})
	arg1 := ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: suiCoins[0].Ref()})
	arg2 := ptb.MustPure(uint64(3))

	lspArg := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:  swapPackageID,
				Module:   "swap",
				Function: "create_pool",
				TypeArguments: []iotago.TypeTag{
					{
						Struct: &iotago.StructTag{
							Address: testcoinID,
							Module:  "testcoin",
							Name:    "TESTCOIN",
						},
					},
				},
				Arguments: []iotago.Argument{arg0, arg1, arg2},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{lspArg},
				Address: ptb.MustPure(signer.Address()),
			},
		},
	)
	pt := ptb.Finish()
	txData := iotago.NewProgrammable(
		signer.Address(),
		pt,
		[]*iotago.ObjectRef{suiCoins[1].Ref()},
		iotaclient2.DefaultGasBudget,
		iotaclient2.DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&txData)
	if err != nil {
		panic(err)
	}

	txnResponse, err := suiClient.SignAndExecuteTransaction(
		context.Background(),
		signer,
		txBytes,
		&iotajsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}
	for _, change := range txnResponse.ObjectChanges {
		if change.Data.Created != nil {
			resource, err := iotago.NewResourceType(change.Data.Created.ObjectType)
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
