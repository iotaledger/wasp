package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestStartNewChain(t *testing.T) {
	client := iscmoveclienttest.NewLocalnetClient()
	signer := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)

	anchor1, err := client.StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:            signer,
			ChainOwnerAddress: signer.Address(),
			PackageID:         l1starter.ISCPackageID(),
			StateMetadata:     []byte{1, 2, 3, 4},
			InitCoinRef:       getCoinsRes.Data[1].Ref(),
			GasPrice:          iotaclient.DefaultGasPrice,
			GasBudget:         iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	t.Log("anchor1: ", anchor1)
	anchor2, err := client.GetAnchorFromObjectID(context.Background(), &anchor1.Object.ID)
	require.NoError(t, err)
	require.Equal(t, anchor1, anchor2)
}

func TestReceiveRequestAndTransition(t *testing.T) {
	client := iscmoveclienttest.NewLocalnetClient()
	cryptolibSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)
	const topUpAmount = 123
	anchor := startNewChain(t, client, chainSigner)

	txnResponse, err := newAssetsBag(client, cryptolibSigner)
	require.NoError(t, err)
	sentAssetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	_, err = assetsBagPlaceCoinAmount(
		client,
		cryptolibSigner,
		sentAssetsBagRef,
		getCoinsRes.Data[2].Ref(),
		iotajsonrpc.IotaCoinType,
		10,
	)
	require.NoError(t, err)

	sentAssetsBagRef, err = client.UpdateObjectRef(context.Background(), sentAssetsBagRef)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		&iscmoveclient.CreateAndSendRequestRequest{
			Signer:        cryptolibSigner,
			PackageID:     l1starter.ISCPackageID(),
			AnchorAddress: anchor.ObjectID,
			AssetsBagRef:  sentAssetsBagRef,
			Message:       iscmovetest.RandomMessage(),
			Allowance:     iscmove.NewAssets(100),
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)

	require.NoError(t, err)

	requestRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	getCoinsRes, err = client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)
	gasCoin1 := getCoinsRes.Data[2]
	txnResponse, err = client.ReceiveRequestsAndTransition(
		context.Background(),
		&iscmoveclient.ReceiveRequestsAndTransitionRequest{
			Signer:        chainSigner,
			PackageID:     l1starter.ISCPackageID(),
			AnchorRef:     &anchor.ObjectRef,
			Reqs:          []iotago.ObjectRef{*requestRef},
			StateMetadata: []byte{1, 2, 3},
			TopUpAmount:   topUpAmount,
			GasPayment:    gasCoin1.Ref(),
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	getObjRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: gasCoin1.CoinObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
	})
	require.NoError(t, err)
	var gasCoin2 iscmoveclient.MoveCoin
	err = iotaclient.UnmarshalBCS(getObjRes.Data.Bcs.Data.MoveObject.BcsBytes, &gasCoin2)
	require.NoError(t, err)
	require.Equal(t, gasCoin1.Balance.Int64()+topUpAmount-txnResponse.Effects.Data.GasFee(), int64(gasCoin2.Balance))
}

func TestRotationTransaction(t *testing.T) {
	client := newLocalnetClient()
	recipientSigner := newSignerWithFunds(t, testSeed, 0)
	chainSigner := newSignerWithFunds(t, testSeed, 1)

	anchor := startNewChain(t, client, chainSigner)

	_, err := client.RotationTransaction(
		context.Background(),
		&iscmoveclient.RotationTransactionRequest{
			Signer:          chainSigner,
			PackageID:       l1starter.ISCPackageID(),
			AnchorRef:       &anchor.ObjectRef,
			RotationAddress: recipientSigner.Address().AsIotaAddress(),
			GasPrice:        iotaclient.DefaultGasPrice,
			GasBudget:       iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	getObjRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: anchor.ObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, recipientSigner.Address().AsIotaAddress(), getObjRes.Data.Owner.AddressOwner)
}

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) *iscmove.AnchorWithRef {
	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)
	anchor, err := client.StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:            signer,
			ChainOwnerAddress: signer.Address(),
			PackageID:         l1starter.ISCPackageID(),
			StateMetadata:     []byte{1, 2, 3, 4},
			InitCoinRef:       getCoinsRes.Data[1].Ref(),
			GasPrice:          iotaclient.DefaultGasPrice,
			GasBudget:         iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	return anchor
}
