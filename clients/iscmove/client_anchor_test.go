package iscmove_test

import (
	"context"
	"testing"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/stretchr/testify/require"
)

func TestStartNewChain(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, suisigner.TestSeed, 0)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	txnBytes, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		})
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	anchor, err := client.GetAnchorFromSuiTransactionBlockResponse(context.Background(), txnResponse)
	require.NoError(t, err)
	t.Log("anchor: ", anchor)
}

func TestReceiveRequest(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	chainSigner := newSignerWithFunds(t, suisigner.TestSeed, 1)

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	anchor := startNewChain(t, client, chainSigner, iscPackageID)

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
	sentAssetsBagRef, err := signAndExecuteTransactionGetObjectRef(client, cryptolibSigner, txnBytes, "assets_bag", "AssetsBag")
	require.NoError(t, err)

	createAndSendRequestTxnBytes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		anchor.Ref.ObjectID,
		sentAssetsBagRef,
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
	requestRef, err := createAndSendRequestRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)

	resGetObject, err := client.GetObject(context.Background(),
		suiclient.GetObjectRequest{ObjectID: anchor.Ref.ObjectID, Options: &suijsonrpc.SuiObjectDataOptions{ShowType: true}})
	require.NoError(t, err)
	anchorRef := resGetObject.Data.Ref()

	receiveAndUpdateStateRootRequestTxnBytes, err := client.ReceiveAndUpdateStateRootRequest(
		context.Background(),
		chainSigner,
		iscPackageID,
		&anchorRef,
		[]*sui.ObjectRef{requestRef},
		[]byte{1, 2, 3},
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	receiveAndUpdateStateRootRequestRes, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(chainSigner),
		receiveAndUpdateStateRootRequestTxnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, receiveAndUpdateStateRootRequestRes.Effects.Data.IsSuccess())
}

func startNewChain(t *testing.T, client *iscmove.Client, signer cryptolib.Signer, iscPackageID *sui.PackageID) *iscmove.Anchor {
	txnBytes, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		})
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	anchor, err := client.GetAnchorFromSuiTransactionBlockResponse(context.Background(), txnResponse)
	require.NoError(t, err)
	return anchor
}
