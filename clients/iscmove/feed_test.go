package iscmove_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

func TestRequestsFeedOwnedRequests(t *testing.T) {
	client := iscmove.NewClient(iscmove.Config{
		APIURL: suiconn.LocalnetEndpointURL,
	})

	iscOwner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	chainOwner := newSignerWithFunds(t, suisigner.TestSeed, 1)

	iscPackageID := buildAndDeployISCContracts(t, client, iscOwner)
	anchor := startNewChain(t, client, chainOwner, iscPackageID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create AssetsBag owned by iscOwner
	txnResponse, err := client.AssetsBagNew(
		ctx,
		iscOwner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	log := testlogger.NewLogger(t)
	feed := iscmove.NewRequestsFeed(
		iscmove.Config{APIURL: suiconn.LocalnetEndpointURL},
		suiconn.LocalnetWebsocketEndpointURL,
		iscPackageID,
		*anchor.ObjectID,
		log,
	)

	anchorUpdates := make(chan *iscmove.RefWithObject[iscmove.Anchor])
	newRequests := make(chan *iscmove.Request, 1)
	feed.SubscribeToUpdates(ctx, anchorUpdates, newRequests)

	// create a Request and send to anchor
	txnResponse, err = client.CreateAndSendRequest(
		ctx,
		iscOwner,
		iscPackageID,
		anchor.ObjectID,
		assetsBagRef,
		isc.Hn("dummy_isc_contract"),
		isc.Hn("dummy_isc_func"),
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	requestRef, err := txnResponse.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	req := <-newRequests
	require.Equal(t, *requestRef.ObjectID, req.ID)

	updatedAnchor, ownedReqs, err := feed.FetchCurrentState(ctx)
	require.NoError(t, err)

	require.Equal(t, anchor.Version, updatedAnchor.Version)

	require.Len(t, ownedReqs, 1)
	require.Equal(t, *requestRef.ObjectID, ownedReqs[0].ID)
}
