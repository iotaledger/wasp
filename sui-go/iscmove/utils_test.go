package iscmove_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/iscmove"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

func buildAndDeployISCContracts(t *testing.T, client *iscmove.Client, signer *sui_signer.Signer) *sui_types.PackageID {
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		iscBytecode.Modules,
		iscBytecode.Dependencies,
		nil,
		models.NewSafeSuiBigInt(sui.DefaultGasBudget*10),
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	return packageID
}

func buildDeployMintTestcoin(t *testing.T, client *iscmove.Client, signer *sui_signer.Signer) (
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
		models.NewSafeSuiBigInt(sui.DefaultGasBudget*10),
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	treasuryCap, _, err := sui.GetCreatedObjectIdAndType(txnResponse, "coin", "TreasuryCap")
	require.NoError(t, err)

	mintAmount := uint64(1000000)
	txnRes, err := client.MintToken(
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
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())

	return packageID, treasuryCap
}
