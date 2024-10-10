package pkg

import (
	"context"
	"fmt"

	sui2 "github.com/iotaledger/wasp/clients/iota-go/sui"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func SwapSui(
	suiClient *suiclient2.Client,
	swapper suisigner.Signer,
	swapPackageID *sui2.PackageID,
	testcoinID *sui2.ObjectID,
	poolObjectID *sui2.ObjectID,
	suiCoins []*suijsonrpc2.Coin,
) {
	poolGetObjectRes, err := suiClient.GetObject(
		context.Background(), suiclient2.GetObjectRequest{
			ObjectID: poolObjectID,
			Options: &suijsonrpc2.SuiObjectDataOptions{
				ShowType:    true,
				ShowContent: true,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	// swap sui to testcoin
	ptb := sui2.NewProgrammableTransactionBuilder()

	arg0 := ptb.MustObj(
		sui2.ObjectArg{
			SharedObject: &sui2.SharedObjectArg{
				Id:                   poolObjectID,
				InitialSharedVersion: poolGetObjectRes.Data.Ref().Version,
				Mutable:              true,
			},
		},
	)
	arg1 := ptb.MustObj(sui2.ObjectArg{ImmOrOwnedObject: suiCoins[0].Ref()})

	retCoinArg := ptb.Command(sui2.Command{
		MoveCall: &sui2.ProgrammableMoveCall{
			Package:  swapPackageID,
			Module:   "swap",
			Function: "swap_sui",
			TypeArguments: []sui2.TypeTag{{Struct: &sui2.StructTag{
				Address: testcoinID,
				Module:  "testcoin",
				Name:    "TESTCOIN",
			}}},
			Arguments: []sui2.Argument{arg0, arg1},
		}},
	)
	ptb.Command(
		sui2.Command{
			TransferObjects: &sui2.ProgrammableTransferObjects{
				Objects: []sui2.Argument{retCoinArg},
				Address: ptb.MustPure(swapper.Address()),
			},
		},
	)
	pt := ptb.Finish()
	txData := sui2.NewProgrammable(
		swapper.Address(),
		pt,
		[]*sui2.ObjectRef{suiCoins[1].Ref()},
		suiclient2.DefaultGasBudget,
		suiclient2.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&txData)
	if err != nil {
		panic(err)
	}

	resp, err := suiClient.SignAndExecuteTransaction(
		context.Background(),
		swapper,
		txBytes,
		&suijsonrpc2.SuiTransactionBlockResponseOptions{
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
