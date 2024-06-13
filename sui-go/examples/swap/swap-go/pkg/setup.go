package pkg

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"github.com/iotaledger/wasp/sui-go/utils"
)

func BuildAndPublish(client *sui.ImplSuiAPI, signer *sui_signer.Signer, path string) *sui_types.PackageID {
	modules, err := utils.MoveBuild(path)
	if err != nil {
		panic(err)
	}
	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		modules.Modules,
		modules.Dependencies,
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

func BuildDeployMintTestcoin(client *sui.ImplSuiAPI, signer *sui_signer.Signer) (
	*sui_types.PackageID,
	*sui_types.ObjectID,
) {
	modules, err := utils.MoveBuild(utils.GetGitRoot() + "/sui-go/contracts/testcoin/")
	if err != nil {
		panic(err)
	}

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		modules.Modules,
		modules.Dependencies,
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

	treasuryCap, _, err := sui.GetCreatedObjectIdAndType(txnResponse, "coin", "TreasuryCap")
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
