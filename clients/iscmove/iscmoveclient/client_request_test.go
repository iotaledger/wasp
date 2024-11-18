package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestCreateAndSendRequest(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)

	anchor, _ := startNewChain(t, client, cryptolibSigner)

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		anchor.ObjectID,
		assetsBagRef,
		&iscmove.Message{
			Contract: uint32(isc.Hn("test_isc_contract")),
			Function: uint32(isc.Hn("test_isc_func")),
			Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		},
		&iscmove.Assets{
			Coins: iscmove.CoinBalances{
				"0x1::iota::IOTA":    11,
				"0xa::testa::TEST_A": 12,
			},
		},
		0,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	_, err = createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)
}

func TestCreateAndSendRequestWithAssets(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)

	anchor, _ := startNewChain(t, client, cryptolibSigner)

	assets := iscmove.NewAssets(100)

	createAndSendRequestRes, err := client.CreateAndSendRequestWithAssets(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		anchor.ObjectID,
		assets,
		&iscmove.Message{
			Contract: uint32(isc.Hn("test_isc_contract")),
			Function: uint32(isc.Hn("test_isc_func")),
			Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		},
		&iscmove.Assets{
			Coins: iscmove.CoinBalances{
				"0x1::iota::IOTA":    11,
				"0xa::testa::TEST_A": 12,
			},
		},
		0,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	_, err = createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)
}

func TestGetRequestFromObjectID(t *testing.T) {
	client := newLocalnetClient()
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)

	anchor, _ := startNewChain(t, client, cryptolibSigner)

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		anchor.ObjectID,
		assetsBagRef,
		&iscmove.Message{
			Contract: uint32(isc.Hn("test_isc_contract")),
			Function: uint32(isc.Hn("test_isc_func")),
			Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		},
		&iscmove.Assets{
			Coins: iscmove.CoinBalances{
				"0x1::iota::IOTA":    11,
				"0xa::testa::TEST_A": 12,
			},
		},
		0,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)

	reqInfo, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	req, err := client.GetRequestFromObjectID(context.Background(), reqInfo.ObjectID)
	require.NoError(t, err)
	require.Equal(t, iotajsonrpc.CoinValue(12), req.Object.Allowance.Coins["0xa::testa::TEST_A"])
	// require.Equal(t, iotajsonrpc.CoinValue(21), req.Object.Allowance.Coins["0x1::iota::IOTA"]) // FIXME
}
