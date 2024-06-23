package iscmove_test

import (
	"context"
	"testing"

	"github.com/iotaledger/wasp/sui-go/iscmove"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
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

	_, _, err = sui.GetCreatedObjectIdAndType(assetsBagNewRes, "assets_bag", "AssetsBag")
	require.NoError(t, err)
}
