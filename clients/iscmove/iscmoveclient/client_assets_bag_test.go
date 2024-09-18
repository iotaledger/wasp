package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func TestAssetsBagNewAndDestroyEmpty(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
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
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
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
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	_, coinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: coinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())
	_, err = client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagMainRef,
		coinInfo.Ref,
		testCointype,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
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
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	_, coinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: coinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())

	_, err = client.AssetsBagPlaceCoinAmount(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagMainRef,
		coinInfo.Ref,
		testCointype,
		10,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
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
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo("assets_bag", "AssetsBag")
	require.NoError(t, err)

	_, coinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: coinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())
	_, err = client.AssetsBagPlaceCoin(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		assetsBagMainRef,
		coinInfo.Ref,
		testCointype,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
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

	_, testcoinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: testcoinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())

	borrowAnchorAssetsAndPlaceCoin(t, context.Background(), client, cryptolibSigner, &anchor.ObjectRef, testcoinInfo)

	assetsBag, err := client.GetAssetsBagWithBalances(context.Background(), &anchor.Object.Assets.Value.ID)
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
	anchorRef *sui.ObjectRef,
	testcoinInfo *sui.ObjectInfo,
) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)
	packageID := l1starter.ISCPackageID()

	ptb := sui.NewProgrammableTransactionBuilder()
	typeTag, err := sui.TypeTagFromString(testcoinInfo.Type.String())
	require.NoError(t, err)
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        "anchor",
				Function:      "borrow_assets",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: anchorRef}),
				},
			},
		},
	)
	argAssetsBag := sui.Argument{NestedResult: &sui.NestedResult{Cmd: 0, Result: 0}}
	argBorrow := sui.Argument{NestedResult: &sui.NestedResult{Cmd: 0, Result: 1}}
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        "assets_bag",
				Function:      "place_coin",
				TypeArguments: []sui.TypeTag{*typeTag},
				Arguments: []sui.Argument{
					argAssetsBag,
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: testcoinInfo.Ref}),
				},
			},
		},
	)
	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       &packageID,
				Module:        "anchor",
				Function:      "return_assets_from_borrow",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: anchorRef}),
					argAssetsBag,
					argBorrow,
				},
			},
		},
	)
	pt := ptb.Finish()
	coins, err := client.GetCoinObjsForTargetAmount(ctx, signer.Address(), suiclient.DefaultGasBudget)
	require.NoError(t, err)
	gasPayments := coins.CoinRefs()

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	require.NoError(t, err)

	execRes, err := client.SignAndExecuteTransaction(ctx, signer, txnBytes, &suijsonrpc.SuiTransactionBlockResponseOptions{
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

	_, testcoinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: testcoinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())

	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
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
		testcoinInfo.Ref,
		testCointype,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	assetsBagGetObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: assetsBagRef.ObjectID})
	require.NoError(t, err)
	tmpAssetsBagRef := assetsBagGetObjectRes.Data.Ref()

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		cryptolibSigner,
		l1starter.ISCPackageID(),
		anchor.ObjectID,
		&tmpAssetsBagRef,
		uint32(isc.Hn("test_isc_contract")),
		uint32(isc.Hn("test_isc_func")),
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		nil,
		0,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	reqRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	reqWithObj, err := client.GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
	require.NoError(t, err)

	assetsBag, err := client.GetAssetsBagWithBalances(context.Background(), &reqWithObj.Object.AssetsBag.Value.ID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), assetsBag.Size)
	bal, ok := assetsBag.Balances[testCointype]
	require.True(t, ok)
	require.Equal(t, testCointype, bal.CoinType)
	require.Equal(t, uint64(1000000), bal.TotalBalance)
}
