package iscmoveclient_test

import (
	"context"
	"fmt"
	"github.com/samber/lo"
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

func TestStartNewChain(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)

	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		getCoinsRes.Data[1].Ref(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
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

	anchor := startNewChain(t, client, chainSigner)

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
		5000,
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
		iscmove.NewAssets(5000),
		5000,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)

	require.NoError(t, err)

	requestRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	chainOwnerFunds, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	fmt.Printf("%v", chainOwnerFunds.Data)

	_, err = client.ReceiveRequestAndTransition(
		context.Background(),
		chainSigner,
		l1starter.ISCPackageID(),
		&anchor.ObjectRef,
		[]iotago.ObjectRef{*requestRef},
		[]byte{1, 2, 3},
		*chainOwnerFunds.Data[3].Ref(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	funds, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})

	gasCoinValue, found := lo.Find(funds.Data, func(item *iotajsonrpc.Coin) bool {
		return item.CoinObjectID.String() == chainOwnerFunds.Data[3].CoinObjectID.String()
	})

	require.True(t, found)
	require.Equal(t, gasCoinValue.Balance.Uint64(), uint64(chainOwnerFunds.Data[3].Balance.Uint64()+1337))
}

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) *iscmove.AnchorWithRef {
	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		l1starter.ISCPackageID(),
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	return anchor
}
