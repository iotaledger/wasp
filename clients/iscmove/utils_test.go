package iscmove_test

import (
	"context"
	"fmt"
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

	txnBytes, err := client.Publish(context.Background(), &models.PublishRequest{
		Sender:          signer.Address().AsSuiAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       models.NewBigInt(sui.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), cryptolib.SignerToSuiSigner(signer), txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageRef, err := txnResponse.GetPublishedPackageRef()
	require.NoError(t, err)

	return packageRef.ObjectID
}

func buildDeployMintTestcoin(t *testing.T, client *iscmove.Client, signer *sui_signer.Signer) (
	*sui_types.ObjectRef,
	*sui_types.ObjectRef,
	*sui_types.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()
	suiSigner := cryptolib.SignerToSuiSigner(signer)

	txnBytes, err := client.Publish(context.Background(), &models.PublishRequest{
		Sender:          signer.Address().AsSuiAddress(),
		CompiledModules: testcoinBytecode.Modules,
		Dependencies:    testcoinBytecode.Dependencies,
		GasBudget:       models.NewBigInt(sui.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
<<<<<<< HEAD:clients/iscmove/utils_test.go
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), suiSigner, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
=======
	deployTxnRes, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
>>>>>>> b796b4b5b (wip):sui-go/iscmove/utils_test.go
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, deployTxnRes.Effects.Data.IsSuccess())

	packageRef, err := deployTxnRes.GetPublishedPackageRef()
	require.NoError(t, err)

<<<<<<< HEAD:clients/iscmove/utils_test.go
	treasuryCap, _, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
=======
	treasuryCap, _, err := sui.GetCreatedObjectRefAndType(deployTxnRes, "coin", "TreasuryCap")
>>>>>>> b796b4b5b (wip):sui-go/iscmove/utils_test.go
	require.NoError(t, err)

	mintAmount := uint64(1000000)
	mintTxnRes, err := client.MintToken(
		context.Background(),
<<<<<<< HEAD:clients/iscmove/utils_test.go
		suiSigner,
		packageID,
=======
		signer,
		packageRef.ObjectID,
>>>>>>> b796b4b5b (wip):sui-go/iscmove/utils_test.go
		"testcoin",
		treasuryCap.ObjectID,
		mintAmount,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, mintTxnRes.Effects.Data.IsSuccess())
	fmt.Println("mintTxnRes")
	testcoinRef, _, err := sui.GetCreatedObjectRefAndType(mintTxnRes, "testcoin", "TESTCOIN")
	require.NoError(t, err)

	return packageRef, testcoinRef, treasuryCap.ObjectID
}
