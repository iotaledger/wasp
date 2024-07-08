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

func TestStartNewChain(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, suisigner.TestSeed, 0)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
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
	t.Log("anchor: ", anchor)
}

func TestReceiveRequest(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	chainSigner := newSignerWithFunds(t, suisigner.TestSeed, 1)
	suiSigner := cryptolib.SignerToSuiSigner(cryptolibSigner)
	suiChainSigner := cryptolib.SignerToSuiSigner(chainSigner)

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	anchor, err := client.StartNewChain(
		context.Background(),
		chainSigner,
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
		suiSigner,
		createAndSendRequestTxnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, createAndSendRequestRes.Effects.Data.IsSuccess())
	requestRef, err := createAndSendRequestRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)

	resGetObject, err := client.GetObject(context.Background(),
		suiclient.GetObjectRequest{ObjectID: anchor.ID, Options: &suijsonrpc.SuiObjectDataOptions{ShowType: true}})
	require.NoError(t, err)
	anchorRef := resGetObject.Data.Ref()

	receiveAndUpdateStateRootRequestTxnBytes, err := client.ReceiveAndUpdateStateRootRequest(
		context.Background(),
		chainSigner,
		iscPackageID,
		&anchorRef,
		requestRef,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	receiveAndUpdateStateRootRequestRes, err := client.SignAndExecuteTransaction(
		context.Background(),
		suiChainSigner,
		receiveAndUpdateStateRootRequestTxnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, receiveAndUpdateStateRootRequestRes.Effects.Data.IsSuccess())
}
