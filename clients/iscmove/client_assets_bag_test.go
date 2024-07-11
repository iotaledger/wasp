package iscmove_test

import (
	"context"
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/stretchr/testify/require"
)

func TestAssetsBagNewAndDestroyEmpty(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnBytes, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, txnBytes)
	assetsBagRef, err := signAndExecuteTransactionGetObjectRef(client, cryptolibSigner, txnBytes, "assets_bag", "AssetsBag")
	require.NoError(t, err)

	assetsDestroyEmptyTxnBytes, err := client.AssetsDestroyEmpty(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		assetsBagRef,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsDestroyEmptyRes, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(cryptolibSigner),
		assetsDestroyEmptyTxnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, assetsDestroyEmptyRes.Effects.Data.IsSuccess())
	_, err = assetsDestroyEmptyRes.GetCreatedObjectInfo("assets_bag", "AssetsBag")
	require.Error(t, err, "not found")
}

func TestAssetsBagAddItems(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnBytes, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, txnBytes)
	assetsBagMain, err := signAndExecuteTransactionGetObjectRef(client, cryptolibSigner, txnBytes, "assets_bag", "AssetsBag")
	require.NoError(t, err)

	_, coinRef := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: coinRef.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	assetsBagAddItemsTxnBytes, err := client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		assetsBagMain,
		coinRef,
		*getCoinRef.Data.Type,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagAddItemsRes, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(cryptolibSigner),
		assetsBagAddItemsTxnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, assetsBagAddItemsRes.Effects.Data.IsSuccess())
}
