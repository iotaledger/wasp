package iscmove_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/sui-go/iscmove"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"github.com/stretchr/testify/require"
)

func TestAssetsBagNew(t *testing.T) {
	suiClient, signer := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 0)
	client := iscmove.NewClient(suiClient)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	assetsBagNewRes, err := client.AssetsBagNew(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, assetsBagNewRes.Effects.Data.IsSuccess())

	_, _, err = sui.GetCreatedObjectRefAndType(assetsBagNewRes, "assets_bag", "AssetsBag")
	require.NoError(t, err)
}

func TestAssetsBagAddItems(t *testing.T) {
	suiClient, signer := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 0)
	client := iscmove.NewClient(suiClient)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	assetsBagNewRes, err := client.AssetsBagNew(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, assetsBagNewRes.Effects.Data.IsSuccess())

	assetsBagMainRef, _, err := sui.GetCreatedObjectRefAndType(assetsBagNewRes, "assets_bag", "AssetsBag")
	require.NoError(t, err)

	assetsBagNewRes, err = client.AssetsBagNew(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, assetsBagNewRes.Effects.Data.IsSuccess())

	assetsBagChildRef, _, err := sui.GetCreatedObjectRefAndType(assetsBagNewRes, "assets_bag", "AssetsBag")
	require.NoError(t, err)

	_, coinRef1, _ := buildDeployMintTestcoin(t, client, signer)
	_, coinRef2, _ := buildDeployMintTestcoin(t, client, signer)
	fmt.Println("assetsBagMainRef: ", assetsBagMainRef)
	fmt.Println("assetsBagChildRef: ", assetsBagChildRef)
	fmt.Println("coinRef1: ", coinRef1)
	fmt.Println("coinRef2: ", coinRef2)
	assetsBagAddItemsRes, err := client.AssetsBagAddItems(
		context.Background(),
		signer,
		iscPackageID,
		assetsBagMainRef,
		[]*sui_types.ObjectRef{coinRef1, coinRef2},
		[]*sui_types.ObjectRef{},
		[]*sui_types.ObjectRef{assetsBagChildRef},
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, assetsBagAddItemsRes.Effects.Data.IsSuccess())
}
