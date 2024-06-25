package pkg

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/move"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

func Publish(client *sui.ImplSuiAPI, signer *sui_signer.Signer, bytecode move.PackageBytecode) *sui_types.PackageID {
	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		bytecode.Modules,
		bytecode.Dependencies,
		nil,
		models.NewSafeSuiBigInt(10*sui.DefaultGasBudget),
	)
	if err != nil {
		panic(err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
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

func PublishMintTestcoin(client *sui.ImplSuiAPI, signer *sui_signer.Signer) (
	*sui_types.PackageID,
	*sui_types.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		testcoinBytecode.Modules,
		testcoinBytecode.Dependencies,
		nil,
		models.NewSafeSuiBigInt(10*sui.DefaultGasBudget),
	)
	if err != nil {
		panic(err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
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

	treasuryCap, _, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	if err != nil {
		panic(err)
	}

	mintAmount := uint64(1000000)
	txnResponse, err = client.MintToken(
		context.Background(),
		signer,
		packageID,
		"testcoin",
		treasuryCap,
		mintAmount,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}

	return packageID, treasuryCap
}
