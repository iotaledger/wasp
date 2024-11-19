package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

const TestTxFeePerReq = 100

func TestStartNewChain(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)
	gasObj := getCoinsRes.Data[2]

	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		getCoinsRes.Data[1].Ref(),
		gasObj.CoinObjectID,
		TestTxFeePerReq,
		[]*iotago.ObjectRef{getCoinsRes.Data[0].Ref()},
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	t.Log("anchor: ", anchor)
}

func TestGetAnchorFromObjectID(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)
	gasObjectAddr := getCoinsRes.Data[2].Ref().ObjectID

	anchor1, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		getCoinsRes.Data[1].Ref(),
		gasObjectAddr,
		100,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	t.Log("anchor1: ", anchor1)

	anchor2, err := client.GetAnchorFromObjectID(context.Background(), &anchor1.Object.ID)
	require.NoError(t, err)
	require.Equal(t, anchor1, anchor2)
}

func TestReceiveRequestAndTransition(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	chainSigner := newSignerWithFunds(t, testSeed, 1)

	anchor, gasObj := startNewChain(t, client, chainSigner)

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	sentAssetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	_, err = client.AssetsBagPlaceCoinAmount(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		sentAssetsBagRef,
		getCoinsRes.Data[len(getCoinsRes.Data)-1].Ref(),
		iotajsonrpc.IotaCoinType,
		1000,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	sentAssetsBagRef, err = client.UpdateObjectRef(context.Background(), sentAssetsBagRef)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		&anchor.Object.ID,
		sentAssetsBagRef,
		&iscmove.Message{
			Contract: uint32(isc.Hn("test_isc_contract")),
			Function: uint32(isc.Hn("test_isc_func")),
			Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		},
		iscmove.NewAssets(100),
		0,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	requestRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	getObjectRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: &gasObj, Options: &iotajsonrpc.IotaObjectDataOptions{ShowOwner: true, ShowContent: true, ShowBcs: true}})
	require.NoError(t, err)
	gasObjRef := getObjectRes.Data.Ref()
	var gasObj1, gasObj2 iscmoveclient.MoveCoin
	err = iotaclient.UnmarshalBCS(getObjectRes.Data.Bcs.Data.MoveObject.BcsBytes, &gasObj1)
	require.NoError(t, err)

	transitionRes, err := client.ReceiveRequestsAndTransition(
		context.Background(),
		chainSigner,
		l1starter.ISCPackageID(),
		&anchor.ObjectRef,
		[]iotago.ObjectRef{*requestRef},
		[]byte{1, 2, 3},
		[]*iotago.ObjectRef{&gasObjRef},
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	getObjectRes, err = client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: &gasObj, Options: &iotajsonrpc.IotaObjectDataOptions{ShowOwner: true, ShowContent: true, ShowBcs: true}})
	require.NoError(t, err)
	err = iotaclient.UnmarshalBCS(getObjectRes.Data.Bcs.Data.MoveObject.BcsBytes, &gasObj2)
	require.NoError(t, err)
	require.Equal(t, int64(gasObj1.Balance)-transitionRes.Effects.Data.GasFee()+TestTxFeePerReq, int64(gasObj2.Balance))
}

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) (*iscmove.RefWithObject[iscmove.Anchor], iotago.ObjectID) {
	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)
	gasObj := getCoinsRes.Data[2]
	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		getCoinsRes.Data[1].Ref(),
		gasObj.CoinObjectID,
		TestTxFeePerReq,
		[]*iotago.ObjectRef{getCoinsRes.Data[0].Ref()},
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	return anchor, *gasObj.CoinObjectID
}
