package pkg

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func SwapSui(
	suiClient *suiclient.Client,
	swapper suisigner.Signer,
	swapPackageID *sui.PackageID,
	testcoinID *sui.ObjectID,
	poolObjectID *sui.ObjectID,
	suiCoins []*suijsonrpc.Coin,
) {
	poolGetObjectRes, err := suiClient.GetObject(
		context.Background(), suiclient.GetObjectRequest{
			ObjectID: poolObjectID,
			Options: &suijsonrpc.SuiObjectDataOptions{
				ShowType:    true,
				ShowContent: true,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	// swap sui to testcoin
	ptb := sui.NewProgrammableTransactionBuilder()

	arg0 := ptb.MustObj(
		sui.ObjectArg{
			SharedObject: &sui.SharedObjectArg{
				Id:                   poolObjectID,
				InitialSharedVersion: poolGetObjectRes.Data.Ref().Version,
				Mutable:              true,
			},
		},
	)
	arg1 := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: suiCoins[0].Ref()})

	retCoinArg := ptb.Command(sui.Command{
		MoveCall: &sui.ProgrammableMoveCall{
			Package:  swapPackageID,
			Module:   "swap",
			Function: "swap_sui",
			TypeArguments: []sui.TypeTag{{Struct: &sui.StructTag{
				Address: testcoinID,
				Module:  "testcoin",
				Name:    "TESTCOIN",
			}}},
			Arguments: []sui.Argument{arg0, arg1},
		}},
	)
	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{retCoinArg},
				Address: ptb.MustPure(swapper.Address()),
			},
		},
	)
	pt := ptb.Finish()
	txData := sui.NewProgrammable(
		swapper.Address(),
		pt,
		[]*sui.ObjectRef{suiCoins[1].Ref()},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(txData)
	if err != nil {
		panic(err)
	}

	resp, err := suiClient.SignAndExecuteTransaction(
		context.Background(),
		swapper,
		txBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
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
