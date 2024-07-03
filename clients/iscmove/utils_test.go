package iscmove_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func buildAndDeployISCContracts(t *testing.T, client *iscmove.Client, signer cryptolib.Signer) *sui.PackageID {
	suiSigner := cryptolib.SignerToSuiSigner(signer)
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          suiSigner.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		suiSigner,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
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

func buildDeployMintTestcoin(t *testing.T, client *iscmove.Client, signer cryptolib.Signer) (*sui.ObjectRef, *sui.ObjectRef) {
	testcoinBytecode := contracts.Testcoin()
	suiSigner := cryptolib.SignerToSuiSigner(signer)

	txnBytes, err := client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          signer.Address().AsSuiAddress(),
		CompiledModules: testcoinBytecode.Modules,
		Dependencies:    testcoinBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		suiSigner,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: packageID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
	})
	require.NoError(t, err)
	packageRef := getObjectRes.Data.Ref()

	treasuryCapRef, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	require.NoError(t, err)

	mintAmount := uint64(1000000)
	txnRes, err := client.MintToken(
		context.Background(),
		suiSigner,
		packageID,
		"testcoin",
		treasuryCapRef.ObjectID,
		mintAmount,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())

	coinRef, err := txnRes.GetCreatedObjectInfo("testcoin", "TESTCOIN")
	require.NoError(t, err)

	return &packageRef, coinRef
}
