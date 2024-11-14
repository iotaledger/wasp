package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestRequestsFeed(t *testing.T) {
	client := newLocalnetClient()

	iscOwner := newSignerWithFunds(t, testSeed, 0)
	chainOwner := newSignerWithFunds(t, testSeed, 1)

	anchor := startNewChain(t, client, chainOwner)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create AssetsBag owned by iscOwner
	txnResponse, err := client.AssetsBagNew(
		ctx,
		iscOwner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	log := testlogger.NewLogger(t)

	wsClient, err := iscmoveclient.NewWebsocketClient(
		ctx,
		iotaconn.LocalnetWebsocketEndpointURL,
		iotaconn.LocalnetFaucetURL,
		log,
	)
	require.NoError(t, err)

	chainFeed := iscmoveclient.NewChainFeed(
		ctx,
		wsClient,
		l1starter.ISCPackageID(),
		*anchor.ObjectID,
		log,
	)
	defer func() {
		cancel()
		chainFeed.WaitUntilStopped()
	}()

	anchorUpdates := make(chan *iscmove.AnchorWithRef, 10)
	newRequests := make(chan *iscmove.RefWithObject[iscmove.Request], 10)
	chainFeed.SubscribeToUpdates(ctx, *anchor.ObjectID, anchorUpdates, newRequests)

	// create a Request and send to anchor
	txnResponse, err = client.CreateAndSendRequest(
		ctx,
		iscOwner,
		l1starter.ISCPackageID(),
		anchor.ObjectID,
		assetsBagRef,
		&iscmove.Message{
			Contract: uint32(isc.Hn("test_isc_contract")),
			Function: uint32(isc.Hn("test_isc_func")),
			Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		},
		nil,
		0,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	requestRef, err := txnResponse.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	req := <-newRequests
	require.Equal(t, *requestRef.ObjectID, req.Object.ID)

	updatedAnchor, ownedReqs, err := chainFeed.FetchCurrentState(ctx)
	require.NoError(t, err)

	require.Equal(t, anchor.Version, updatedAnchor.Version)

	require.Len(t, ownedReqs, 1)
	require.Equal(t, *requestRef.ObjectID, ownedReqs[0].Object.ID)

	_, err = client.ReceiveRequestsAndTransition(
		context.Background(),
		chainOwner,
		l1starter.ISCPackageID(),
		&anchor.ObjectRef,
		[]iotago.ObjectRef{*requestRef},
		[]byte{1, 2, 3},
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	upd := <-anchorUpdates
	require.EqualValues(t, []byte{1, 2, 3}, upd.Object.StateMetadata)
}
