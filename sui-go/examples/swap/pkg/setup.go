package pkg

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/move"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func Publish(
	client *suiclient.Client,
	signer suisigner.Signer,
	bytecode move.PackageBytecode,
) *sui.PackageID {
	txnBytes, err := client.Publish(
		context.Background(),
		suiclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: bytecode.Modules,
			Dependencies:    bytecode.Dependencies,
			GasBudget:       suijsonrpc.NewBigInt(10 * suiclient.DefaultGasBudget),
		},
	)
	if err != nil {
		panic(err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}
	packageID, err := txnResponse.GetPublishedPackageID()
	if err != nil {
		panic(err)
	}
	return packageID
}

func PublishMintTestcoin(client *suiclient.Client, signer suisigner.Signer) (
	*sui.PackageID,
	*sui.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		suiclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       suijsonrpc.NewBigInt(10 * suiclient.DefaultGasBudget),
		},
	)
	if err != nil {
		panic(err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}

	packageID, err := txnResponse.GetPublishedPackageID()
	if err != nil {
		panic(err)
	}

	treasuryCap, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	if err != nil {
		panic(err)
	}

	mintAmount := uint64(1000000)
	txnResponse, err = client.MintToken(
		context.Background(),
		signer,
		packageID,
		"testcoin",
		treasuryCap.ObjectID,
		mintAmount,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}

	return packageID, treasuryCap.ObjectID
}
