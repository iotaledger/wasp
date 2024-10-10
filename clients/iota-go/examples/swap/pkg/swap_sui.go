package pkg

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func SwapSui(
	suiClient *iotaclient.Client,
	swapper iotasigner.Signer,
	swapPackageID *iotago.PackageID,
	testcoinID *iotago.ObjectID,
	poolObjectID *iotago.ObjectID,
	suiCoins []*iotajsonrpc.Coin,
) {
	poolGetObjectRes, err := suiClient.GetObject(
		context.Background(), iotaclient.GetObjectRequest{
			ObjectID: poolObjectID,
			Options: &iotajsonrpc.SuiObjectDataOptions{
				ShowType:    true,
				ShowContent: true,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	// swap iotago to testcoin
	ptb := iotago.NewProgrammableTransactionBuilder()

	arg0 := ptb.MustObj(
		iotago.ObjectArg{
			SharedObject: &iotago.SharedObjectArg{
				Id:                   poolObjectID,
				InitialSharedVersion: poolGetObjectRes.Data.Ref().Version,
				Mutable:              true,
			},
		},
	)
	arg1 := ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: suiCoins[0].Ref()})

	retCoinArg := ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:  swapPackageID,
				Module:   "swap",
				Function: "swap_sui",
				TypeArguments: []iotago.TypeTag{
					{
						Struct: &iotago.StructTag{
							Address: testcoinID,
							Module:  "testcoin",
							Name:    "TESTCOIN",
						},
					},
				},
				Arguments: []iotago.Argument{arg0, arg1},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{retCoinArg},
				Address: ptb.MustPure(swapper.Address()),
			},
		},
	)
	pt := ptb.Finish()
	txData := iotago.NewProgrammable(
		swapper.Address(),
		pt,
		[]*iotago.ObjectRef{suiCoins[1].Ref()},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&txData)
	if err != nil {
		panic(err)
	}

	resp, err := suiClient.SignAndExecuteTransaction(
		context.Background(),
		swapper,
		txBytes,
		&iotajsonrpc.SuiTransactionBlockResponseOptions{
			ShowObjectChanges: true,
			ShowEffects:       true,
		},
	)
	if err != nil || !resp.Effects.Data.IsSuccess() {
		panic(err)
	}

	for _, change := range resp.ObjectChanges {
		if change.Data.Created != nil {
			fmt.Println("change.Data.Created.ObjectID: ", change.Data.Created.ObjectID)
			fmt.Println("change.Data.Created.ObjectType: ", change.Data.Created.ObjectType)
			fmt.Println("change.Data.Created.Owner.AddressOwner: ", change.Data.Created.Owner.AddressOwner)
		}
		if change.Data.Mutated != nil {
			fmt.Println("change.Data.Mutated.ObjectID: ", change.Data.Mutated.ObjectID)
			fmt.Println("change.Data.Mutated.ObjectType: ", change.Data.Mutated.ObjectType)
			fmt.Println("change.Data.Mutated.Owner.AddressOwner: ", change.Data.Mutated.Owner.AddressOwner)
		}
	}
}
