package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
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
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	AllowanceNewRes, err := client.AllowanceNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	allowanceRef, err := AllowanceNewRes.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		anchor.ObjectID,
		assetsBagRef,
		uint32(isc.Hn("test_isc_contract")),
		uint32(isc.Hn("test_isc_func")),
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		allowanceRef,
		0,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	_, err = createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)
}
