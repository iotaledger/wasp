package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestStartNewChain(t *testing.T) {
	client := iscmoveclienttest.NewHTTPClient()
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
			ChainGasCoin:      getCoinsRes.Data[1].Ref(),
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

func ensureCoinSplit(t *testing.T, cryptolibSigner cryptolib.Signer, client clients.L1Client) {
	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	if len(getCoinsRes.Data) > 1 {
		return
	}

	coins, err := client.GetCoinObjsForTargetAmount(context.Background(), cryptolibSigner.Address().AsIotaAddress(), isc.GasCoinMaxValue, iotaclient.DefaultGasBudget)
	require.NoError(t, err)

	referenceGasPrice, err := client.GetReferenceGasPrice(context.TODO())
	require.NoError(t, err)

	txb := iotago.NewProgrammableTransactionBuilder()

	splitCmd := txb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.GetArgumentGasCoin(),
				Amounts: []iotago.Argument{txb.MustPure(isc.GasCoinMaxValue)},
			},
		},
	)
	txb.TransferArg(cryptolibSigner.Address().AsIotaAddress(), splitCmd)

	txData := iotago.NewProgrammable(
		cryptolibSigner.Address().AsIotaAddress(),
		txb.Finish(),
		[]*iotago.ObjectRef{coins[0].Ref()},
		iotaclient.DefaultGasBudget,
		referenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	result, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			Signer:      cryptolib.SignerToIotaSigner(cryptolibSigner),
			TxDataBytes: txnBytes,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func startNewChain(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) *iscmove.AnchorWithRef {
	ensureCoinSplit(t, signer, l1starter.Instance().L1Client())

	coinObjects, err := client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsIotaAddress(), isc.GasCoinMaxValue, iotaclient.DefaultGasBudget)
	require.NoError(t, err)

	chainGasCoins, gasCoin, err := coinObjects.PickIOTACoinsWithGas(iotajsonrpc.NewBigInt(isc.GasCoinMaxValue).Int, iotaclient.DefaultGasBudget, iotajsonrpc.PickMethodSmaller)
	require.NoError(t, err)

	selectedChainGasCoin, err := chainGasCoins.PickCoinNoLess(isc.GasCoinMaxValue)
	require.NoError(t, err)

	anchor, err := client.StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:            signer,
			ChainOwnerAddress: signer.Address(),
			PackageID:         l1starter.ISCPackageID(),
			StateMetadata:     []byte{1, 2, 3, 4},
			ChainGasCoin:      selectedChainGasCoin.Ref(),
			GasPayments:       []*iotago.ObjectRef{gasCoin.Ref()},
			GasPrice:          iotaclient.DefaultGasPrice,
			GasBudget:         iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	return anchor
}
