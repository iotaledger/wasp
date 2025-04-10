package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestRequestsFeed(t *testing.T) {
	t.Skip()
	client := iscmoveclienttest.NewHTTPClient()

	iscOwner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	anchorOwner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)

	anchor := startNewChain(t, client, anchorOwner)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create AssetsBag owned by iscOwner
	txnResponse, err := newAssetsBag(client, iscOwner)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	log := testlogger.NewLogger(t)

	chainFeed, err := iscmoveclient.NewChainFeed(
		ctx,
		l1starter.ISCPackageID(),
		*anchor.ObjectID,
		log,
		iotaconn.AlphanetWebsocketEndpointURL,
		iotaconn.AlphanetEndpointURL,
	)
	require.NoError(t, err)
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
		&iscmoveclient.CreateAndSendRequestRequest{
			Signer:        iscOwner,
			PackageID:     l1starter.ISCPackageID(),
			AnchorAddress: anchor.ObjectID,
			AssetsBagRef:  assetsBagRef,
			Message:       iscmovetest.RandomMessage(),
			Allowance:     iscmove.NewAssets(100),
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	requestRef, err := txnResponse.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	req := <-newRequests
	require.Equal(t, *requestRef.ObjectID, req.Object.ID)

	ownedReqs := make([]*iscmove.RefWithObject[iscmove.Request], 0)
	updatedAnchor, err := chainFeed.FetchCurrentState(ctx, 1000, func(err error, i *iscmove.RefWithObject[iscmove.Request]) {
		require.NoError(t, err)
		ownedReqs = append(ownedReqs, i)
	})
	require.NoError(t, err)

	require.Equal(t, anchor.Version, updatedAnchor.Version)

	require.Len(t, ownedReqs, 1)
	require.Equal(t, *requestRef.ObjectID, ownedReqs[0].Object.ID)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: anchorOwner.Address().AsIotaAddress()})
	require.NoError(t, err)

	_, err = client.ReceiveRequestsAndTransition(
		context.Background(),
		&iscmoveclient.ReceiveRequestsAndTransitionRequest{
			Signer:           anchorOwner,
			PackageID:        l1starter.ISCPackageID(),
			AnchorRef:        &anchor.ObjectRef,
			ConsumedRequests: []iotago.ObjectRef{*requestRef},
			SentAssets:       []iscmoveclient.SentAssets{},
			StateMetadata:    []byte{1, 2, 3},
			TopUpAmount:      100,
			GasPayment:       getCoinsRes.Data[0].Ref(),
			GasPrice:         iotaclient.DefaultGasPrice,
			GasBudget:        iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	upd := <-anchorUpdates
	require.EqualValues(t, []byte{1, 2, 3}, upd.Object.StateMetadata)
}
