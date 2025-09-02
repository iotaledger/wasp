package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

// TestRequestsFeed relies of the alphanet, so can't use global l1starter
func TestRequestsFeed(t *testing.T) {
	client := iscmoveclienttest.NewAlphanetHTTPClient()

	iscOwner := iscmoveclienttest.NewAlphanetSignerWithFunds(t, testcommon.TestSeed, 0)
	anchorOwner := iscmoveclienttest.NewAlphanetSignerWithFunds(t, testcommon.TestSeed, 1)

	remoteNode := l1starter.NewRemoteIotaNode(iotaconn.AlphanetEndpointURL, iotaconn.AlphanetFaucetURL, cryptolib.SignerToIotaSigner(iscOwner))
	remoteNode.Start(context.Background())

	anchor := StartNewChainWithPackageIDAndL1Client(t, client, anchorOwner, remoteNode.ISCPackageID(), remoteNode.L1Client())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create AssetsBag owned by iscOwner
	txnResponse, err := NewAssetsBagWithPackageID(client, iscOwner, remoteNode.ISCPackageID())
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectByName(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	log := testlogger.NewLogger(t)

	chainFeed, err := iscmoveclient.NewChainFeed(
		ctx,
		remoteNode.ISCPackageID(),
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
			PackageID:     remoteNode.ISCPackageID(),
			AnchorAddress: anchor.ObjectID,
			AssetsBagRef:  assetsBagRef,
			Message:       iscmovetest.RandomMessage(),
			AllowanceBCS:  nil,
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	requestRef, err := txnResponse.GetCreatedObjectByName(iscmove.RequestModuleName, iscmove.RequestObjectName)
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
			PackageID:        remoteNode.ISCPackageID(),
			AnchorRef:        &anchor.ObjectRef,
			ConsumedRequests: []iotago.ObjectRef{*requestRef},
			SentAssets:       []iscmoveclient.SentAssets{},
			StateMetadata:    []byte{1, 2, 3},
			TopUpAmount:      100,
			GasPayment: lo.MaxBy(getCoinsRes.Data, func(a, b *iotajsonrpc.Coin) bool {
				return a.Balance.Int.Cmp(b.Balance.Int) >= 0
			}).Ref(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	upd := <-anchorUpdates
	require.EqualValues(t, []byte{1, 2, 3}, upd.Object.StateMetadata)
}
