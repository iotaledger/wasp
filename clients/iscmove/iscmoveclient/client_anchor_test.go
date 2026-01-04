package iscmoveclient_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestStartNewChain(t *testing.T) {
	client := iscmoveclienttest.NewHTTPClient()
	signer := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	time.Sleep(1 * time.Second) // FIXME tmp for graphql
	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)

	anchor1, err := client.StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        signer,
			AnchorOwner:   signer.Address(),
			PackageID:     l1starter.ISCPackageID(),
			StateMetadata: []byte{1, 2, 3, 4},
			InitCoinRef:   getCoinsRes.Data[1].Ref(),
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	t.Log("anchor1: ", anchor1)
	anchor2, err := client.GetAnchorFromObjectID(context.Background(), &anchor1.Object.ID)
	require.NoError(t, err)
	require.Equal(t, anchor1, anchor2)
}

func TestReceiveRequestAndTransition(t *testing.T) {
	client := iscmoveclienttest.NewHTTPClient()
	cryptolibSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)

	time.Sleep(1 * time.Second) // FIXME tmp for graphql
	const topUpAmount = 123
	anchor := startNewChain(t, client, chainSigner)

	txnResponse, err := newAssetsBag(client, cryptolibSigner)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)

	sentAssetsBagRef, err := txnResponse.GetCreatedObjectByName(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	// Sleep to allow indexer to catch up after newAssetsBag transaction

	_, err = assetsBagPlaceCoinAmountWithGasCoin(
		client,
		cryptolibSigner,
		sentAssetsBagRef,
		iotagraphql.IotaCoinType,
		10,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)

	sentAssetsBagRef, err = client.UpdateObjectRef(context.Background(), sentAssetsBagRef)
	require.NoError(t, err)

	// Fetch fresh coin references after assetsBagPlaceCoinAmountWithGasCoin modified the gas coin
	// Sleep to allow indexer to catch up
	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	var createAndSendRequestRes *iotagraphql.IotaTransactionBlockResponse
	client.MustWaitForNextVersionForTesting(context.Background(), 30*time.Second, nil, getCoinsRes.Data[1].Ref(), func() {
		createAndSendRequestRes, err = client.CreateAndSendRequest(
			context.Background(),
			&iscmoveclient.CreateAndSendRequestRequest{
				Signer:        cryptolibSigner,
				PackageID:     l1starter.ISCPackageID(),
				AnchorAddress: anchor.ObjectID,
				AssetsBagRef:  sentAssetsBagRef,
				Message:       iscmovetest.RandomMessage(),
				AllowanceBCS:  nil,
				GasPayments: []*iotago.ObjectRef{
					getCoinsRes.Data[1].Ref(),
				},
				GasPrice:  iotaclient.DefaultGasPrice,
				GasBudget: iotaclient.DefaultGasBudget,
			},
		)

		require.NoError(t, err)
	})
	time.Sleep(1 * time.Second)

	requestRef, err := createAndSendRequestRes.GetCreatedObjectByName(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	// Sleep to allow indexer to catch up after request creation
	getCoinsRes, err = client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)
	gasCoin1 := getCoinsRes.Data[1]

	client.MustWaitForNextVersionForTesting(context.Background(), 30*time.Second, nil, requestRef, func() {
		client.MustWaitForNextVersionForTesting(context.Background(), 30*time.Second, nil, gasCoin1.Ref(), func() {
			txnResponse, err = client.ReceiveRequestsAndTransition(
				context.Background(),
				&iscmoveclient.ReceiveRequestsAndTransitionRequest{
					Signer:           chainSigner,
					PackageID:        l1starter.ISCPackageID(),
					AnchorRef:        &anchor.ObjectRef,
					ConsumedRequests: []iotago.ObjectRef{*requestRef},
					SentAssets:       []iscmoveclient.SentAssets{},
					StateMetadata:    []byte{1, 2, 3},
					TopUpAmount:      topUpAmount,
					GasPayment:       gasCoin1.Ref(),
					GasPrice:         iotaclient.DefaultGasPrice,
					GasBudget:        iotaclient.DefaultGasBudget,
				},
			)
			require.NoError(t, err)
		})
	})
	time.Sleep(1 * time.Second)

	getObjRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: gasCoin1.CoinObjectID,
		Options:  &iotagraphql.IotaObjectDataOptions{ShowBcs: true},
	})
	require.NoError(t, err)
	var gasCoin2 iscmoveclient.MoveCoin
	err = iotaclient.UnmarshalBCS(getObjRes.Data.Bcs.Data.MoveObject.BcsBytes, &gasCoin2)
	require.NoError(t, err)
	require.Equal(t, gasCoin1.Balance.Int64()+topUpAmount-txnResponse.Effects.Data.GasFee(), int64(gasCoin2.Balance))
}

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) *iscmove.AnchorWithRef {
	return StartNewChainWithPackageIDAndL1Client(t, client, signer, l1starter.ISCPackageID(), l1starter.Instance().L1Client())
}

func StartNewChainWithPackageIDAndL1Client(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer, packageID iotago.PackageID, l1Client clients.L1Client) *iscmove.AnchorWithRef {
	iotatest.EnsureCoinSplitWithBalance(t, cryptolib.SignerToIotaSigner(signer), l1Client, isc.GasCoinTargetValue)

	coinObjects, err := client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsIotaAddress(), isc.GasCoinTargetValue, iotaclient.DefaultGasBudget)
	require.NoError(t, err)

	chainGasCoins, gasCoin, err := coinObjects.PickIOTACoinsWithGas(iotagraphql.NewBigInt(isc.GasCoinTargetValue).Int, iotaclient.DefaultGasBudget, iotagraphql.PickMethodSmaller)
	require.NoError(t, err)

	selectedChainGasCoin, err := chainGasCoins.PickCoinNoLess(isc.GasCoinTargetValue)
	require.NoError(t, err)

	var anchor *iscmove.AnchorWithRef
	client.MustWaitForNextVersionForTesting(context.Background(), 30*time.Second, nil, gasCoin.Ref(), func() {
		client.MustWaitForNextVersionForTesting(context.Background(), 30*time.Second, nil, selectedChainGasCoin.Ref(), func() {
			anchor, err = client.StartNewChain(
				context.Background(),
				&iscmoveclient.StartNewChainRequest{
					Signer:        signer,
					AnchorOwner:   signer.Address(),
					PackageID:     packageID,
					StateMetadata: []byte{1, 2, 3, 4},
					InitCoinRef:   selectedChainGasCoin.Ref(),
					GasPayments:   []*iotago.ObjectRef{gasCoin.Ref()},
					GasPrice:      iotaclient.DefaultGasPrice,
					GasBudget:     iotaclient.DefaultGasBudget,
				},
			)
			require.NoError(t, err)
		})
	})

	return anchor
}
