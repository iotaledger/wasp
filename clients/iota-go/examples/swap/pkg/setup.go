package pkg

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/move"
	"github.com/iotaledger/wasp/clients/iota-go/sui"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
)

func Publish(
	client *suiclient2.Client,
	signer suisigner.Signer,
	bytecode move.PackageBytecode,
) *sui.PackageID {
	txnBytes, err := client.Publish(
		context.Background(),
		suiclient2.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: bytecode.Modules,
			Dependencies:    bytecode.Dependencies,
			GasBudget:       suijsonrpc2.NewBigInt(10 * suiclient2.DefaultGasBudget),
		},
	)
	if err != nil {
		panic(err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &suijsonrpc2.SuiTransactionBlockResponseOptions{
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

func PublishMintTestcoin(client *suiclient2.Client, signer suisigner.Signer) (
	*sui.PackageID,
	*sui.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		suiclient2.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       suijsonrpc2.NewBigInt(10 * suiclient2.DefaultGasBudget),
		},
	)
	if err != nil {
		panic(err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &suijsonrpc2.SuiTransactionBlockResponseOptions{
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
		&suijsonrpc2.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}

	return packageID, treasuryCap.ObjectID
}
