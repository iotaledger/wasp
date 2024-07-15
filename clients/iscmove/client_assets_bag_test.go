package iscmove_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

func TestAssetsBagNewAndDestroyEmpty(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	assetsDestroyEmptyRes, err := client.AssetsDestroyEmpty(
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

	_, err = assetsDestroyEmptyRes.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.Error(t, err, "not found")
}

func TestAssetsBagAddItems(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
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

	_, err = client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		assetsBagMainRef,
		coinRef,
		*getCoinRef.Data.Type,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
}
