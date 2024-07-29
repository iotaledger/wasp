package iscmove_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove/iscmove_types"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/sui-go/suiclient"
)

func TestCreateAndSendRequest(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	anchor := startNewChain(t, client, cryptolibSigner, iscPackageID)

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
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove_types.AssetsBagModuleName, iscmove_types.AssetsBagObjectName)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		anchor.ObjectID,
		assetsBagRef,
		isc.Hn("test_isc_contract"),
		isc.Hn("test_isc_func"),
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	_, err = createAndSendRequestRes.GetCreatedObjectInfo(iscmove_types.RequestModuleName, iscmove_types.RequestObjectName)
	require.NoError(t, err)
}
