package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestStartNewChain(t *testing.T) {
	client := iscmoveclienttest.NewHTTPClient()
	signer := iscmoveclienttest.GenSignerWithFundByCounter(t)

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
	client := iscmoveclienttest.NewHTTPClient()
	cryptolibSigner := iscmoveclienttest.GenSignerWithFundByCounter(t)
	chainSigner := iscmoveclienttest.GenSignerWithFundByCounter(t)
	const topUpAmount = 123
	anchor := startNewChain(t, client, chainSigner)

	txnResponse, err := newAssetsBag(client, cryptolibSigner)
	require.NoError(t, err)
	sentAssetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	_, err = assetsBagPlaceCoinAmountWithGasCoin(
		client,
		cryptolibSigner,
		sentAssetsBagRef,
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
			GasPayments: []*iotago.ObjectRef{
				getCoinsRes.Data[2].Ref(),
			},
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
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

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) *iscmove.AnchorWithRef {
	iotatest.EnsureCoinSplitWithBalance(t, cryptolib.SignerToIotaSigner(signer), l1starter.Instance().L1Client(), isc.GasCoinTargetValue)

	coinObjects, err := client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsIotaAddress(), isc.GasCoinTargetValue, iotaclient.DefaultGasBudget)
	require.NoError(t, err)

	chainGasCoins, gasCoin, err := coinObjects.PickIOTACoinsWithGas(iotajsonrpc.NewBigInt(isc.GasCoinTargetValue).Int, iotaclient.DefaultGasBudget, iotajsonrpc.PickMethodSmaller)
	require.NoError(t, err)

	selectedChainGasCoin, err := chainGasCoins.PickCoinNoLess(isc.GasCoinTargetValue)
	require.NoError(t, err)

	anchor, err := client.StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:            signer,
			ChainOwnerAddress: signer.Address(),
			PackageID:         l1starter.ISCPackageID(),
			StateMetadata:     []byte{1, 2, 3, 4},
			InitCoinRef:       selectedChainGasCoin.Ref(),
			GasPayments:       []*iotago.ObjectRef{gasCoin.Ref()},
			GasPrice:          iotaclient.DefaultGasPrice,
			GasBudget:         iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	return anchor
}
