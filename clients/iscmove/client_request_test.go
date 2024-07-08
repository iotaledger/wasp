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

func TestCreateAndSendRequest(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	anchor, err := client.StartNewChain(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)

	assetsBagRef, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)

	createAndSendRequestTxnBytes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		anchor.ID,
		assetsBagRef,
		"test_isc_contract",
		"test_isc_func",
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	createAndSendRequestRes, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(cryptolibSigner),
		createAndSendRequestTxnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, createAndSendRequestRes.Effects.Data.IsSuccess())

	_, err = createAndSendRequestRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)
}
