package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/util/bcs"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestAssetsBagNewAndDestroyEmpty(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	assetsDestroyEmptyRes, err := client.AssetsDestroyEmpty(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagRef,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	_, err = assetsDestroyEmptyRes.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.Error(t, err, "not found")
}

func TestAssetsBagPlaceCoin(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	coinRef, _ := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		iotaclient.GetObjectRequest{
			ObjectID: coinRef.ObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := iotago.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := iotajsonrpc.CoinType(coinResource.SubType1.String())
	_, err = client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagMainRef,
		coinRef,
		testCointype,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
}

func TestAssetsBagPlaceCoinAmount(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	coinRef, _ := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		iotaclient.GetObjectRequest{
			ObjectID: coinRef.ObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := iotago.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := iotajsonrpc.CoinType(coinResource.SubType1.String())

	_, err = client.AssetsBagPlaceCoinAmount(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagMainRef,
		coinRef,
		testCointype,
		10,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
}

func TestGetAssetsBagFromAssetsBagID(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo("assets_bag", "AssetsBag")
	require.NoError(t, err)

	coinRef, _ := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		iotaclient.GetObjectRequest{
			ObjectID: coinRef.ObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := iotago.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := iotajsonrpc.CoinType(coinResource.SubType1.String())
	_, err = client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagMainRef,
		coinRef,
		testCointype,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	assetsBag, err := client.GetAssetsBagWithBalances(context.Background(), assetsBagMainRef.ObjectID)
	require.NoError(t, err)
	require.Equal(t, *assetsBagMainRef.ObjectID, assetsBag.ID)
	require.Equal(t, uint64(1), assetsBag.Size)
	bal, ok := assetsBag.Balances[testCointype]
	require.True(t, ok)
	require.Equal(t, testCointype, bal.CoinType)
	require.Equal(t, uint64(1000000), bal.TotalBalance)
}

func TestGetAssetsBagFromAnchorID(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	anchor := startNewChain(t, client, cryptolibSigner)

	coinRef, coinType := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		iotaclient.GetObjectRequest{
			ObjectID: coinRef.ObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := iotago.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := iotajsonrpc.CoinType(coinResource.SubType1.String())

	borrowAnchorAssetsAndPlaceCoin(
		t,
		context.Background(),
		client,
		cryptolibSigner,
		&anchor.ObjectRef,
		coinRef,
		coinType,
	)

	assetsBag, err := client.GetAssetsBagWithBalances(context.Background(), &anchor.Object.Assets.ID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), assetsBag.Size)
	bal, ok := assetsBag.Balances[testCointype]
	require.True(t, ok)
	require.Equal(t, testCointype, bal.CoinType)
	require.Equal(t, uint64(1000000), bal.TotalBalance)
}

func borrowAnchorAssetsAndPlaceCoin(
	t *testing.T, ctx context.Context,
	client *iscmoveclient.Client,
	cryptolibSigner cryptolib.Signer,
	anchorRef *iotago.ObjectRef,
	coinRef *iotago.ObjectRef,
	coinType *iotago.ResourceType,
) {
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)
	packageID := l1starter.ISCPackageID()

	ptb := iotago.NewProgrammableTransactionBuilder()
	typeTag, err := iotago.TypeTagFromString(coinType.String())
	require.NoError(t, err)
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        "anchor",
				Function:      "borrow_assets",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: anchorRef}),
				},
			},
		},
	)
	argAssetsBag := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: 0, Result: 0}}
	argBorrow := iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: 0, Result: 1}}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        "assets_bag",
				Function:      "place_coin",
				TypeArguments: []iotago.TypeTag{*typeTag},
				Arguments: []iotago.Argument{
					argAssetsBag,
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coinRef}),
				},
			},
		},
	)
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        "anchor",
				Function:      "return_assets_from_borrow",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: anchorRef}),
					argAssetsBag,
					argBorrow,
				},
			},
		},
	)
	pt := ptb.Finish()
	coins, err := client.GetCoinObjsForTargetAmount(ctx, signer.Address(), iotaclient.DefaultGasBudget)
	require.NoError(t, err)
	gasPayments := coins.CoinRefs()

	tx := iotago.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	execRes, err := client.SignAndExecuteTransaction(ctx, signer, txnBytes, &iotajsonrpc.IotaTransactionBlockResponseOptions{
		ShowEffects:       true,
		ShowObjectChanges: true,
	})
	require.NoError(t, err)
	require.True(t, execRes.Effects.Data.IsSuccess())
}

func TestGetAssetsBagFromRequestID(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	anchor := startNewChain(t, client, cryptolibSigner)

	coinRef, _ := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		iotaclient.GetObjectRequest{
			ObjectID: coinRef.ObjectID,
			Options:  &iotajsonrpc.IotaObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := iotago.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := iotajsonrpc.CoinType(coinResource.SubType1.String())

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	_, err = client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagRef,
		coinRef,
		testCointype,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	assetsBagGetObjectRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: assetsBagRef.ObjectID})
	require.NoError(t, err)
	tmpAssetsBagRef := assetsBagGetObjectRes.Data.Ref()

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		anchor.ObjectID,
		&tmpAssetsBagRef,
		&iscmove.Message{
			Contract: uint32(isc.Hn("test_isc_contract")),
			Function: uint32(isc.Hn("test_isc_func")),
			Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
		},
		nil,
		0,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	reqRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	reqWithObj, err := client.GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
	require.NoError(t, err)

	assetsBag, err := client.GetAssetsBagWithBalances(context.Background(), &reqWithObj.Object.AssetsBag.ID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), assetsBag.Size)
	bal, ok := assetsBag.Balances[testCointype]
	require.True(t, ok)
	require.Equal(t, testCointype, bal.CoinType)
	require.Equal(t, uint64(1000000), bal.TotalBalance)
}
