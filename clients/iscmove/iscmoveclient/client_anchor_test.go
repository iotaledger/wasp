package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	iotaclient2 "github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestStartNewChain(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient2.GetCoinsRequest{Owner: signer.Address().AsSuiAddress()})
	require.NoError(t, err)

	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		getCoinsRes.Data[1].Ref(),
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	t.Log("anchor: ", anchor)
}

func TestGetAnchorFromObjectID(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)

	anchor1, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
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

	anchor := startNewChain(t, client, chainSigner)

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	sentAssetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient2.GetCoinsRequest{Owner: cryptolibSigner.Address().AsSuiAddress()})
	require.NoError(t, err)

	_, err = client.AssetsBagPlaceCoinAmount(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		sentAssetsBagRef,
		getCoinsRes.Data[len(getCoinsRes.Data)-1].Ref(),
		iotajsonrpc.IotaCoinType,
		10,
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
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
		uint32(isc.Hn("test_isc_contract")),
		uint32(isc.Hn("test_isc_func")),
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		nil,
		0,
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
	)

	require.NoError(t, err)

	requestRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	_, err = client.ReceiveRequestAndTransition(
		context.Background(),
		chainSigner,
		l1starter.ISCPackageID(),
		&anchor.ObjectRef,
		[]iotago.ObjectRef{*requestRef},
		[]byte{1, 2, 3},
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
}

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) *iscmove.AnchorWithRef {
	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		iotaclient2.DefaultGasPrice,
		iotaclient2.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	return anchor
}
