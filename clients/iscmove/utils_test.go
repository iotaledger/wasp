package iscmove_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

func buildAndDeployISCContracts(t *testing.T, client *iscmove.Client, signer cryptolib.Signer) *sui_types.PackageID {
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address().AsSuiAddress(),
		iscBytecode.Modules,
		iscBytecode.Dependencies,
		nil,
		models.NewBigInt(sui.DefaultGasBudget*10),
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), cryptolib.SignerToSuiSigner(signer), txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
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

func buildDeployMintTestcoin(t *testing.T, client *iscmove.Client, signer cryptolib.Signer) (
	*sui_types.PackageID,
	*sui_types.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()
	suiSigner := cryptolib.SignerToSuiSigner(signer)

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address().AsSuiAddress(),
		testcoinBytecode.Modules,
		testcoinBytecode.Dependencies,
		nil,
		models.NewBigInt(sui.DefaultGasBudget*10),
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), suiSigner, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	treasuryCap, _, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	require.NoError(t, err)

	mintAmount := uint64(1000000)
	txnRes, err := client.MintToken(
		context.Background(),
		suiSigner,
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
