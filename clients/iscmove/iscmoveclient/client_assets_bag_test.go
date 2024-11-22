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
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestAssetsBagNewAndDestroyEmpty(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagNewAndTransfer(ptb, l1starter.ISCPackageID(), cryptolibSigner.Address())
		},
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	assetsDestroyEmptyRes, err := iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsDestroyEmpty(ptb, l1starter.ISCPackageID(), ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}))
		},
	)
	require.NoError(t, err)

	_, err = assetsDestroyEmptyRes.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.Error(t, err, "not found")
}

func TestAssetsBagPlaceCoin(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := newAssetsBag(client, cryptolibSigner)
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
	_, err = iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagPlaceCoin(
				ptb,
				l1starter.ISCPackageID(),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagMainRef}),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coinRef}),
				string(testCointype),
			)
		},
	)
	require.NoError(t, err)
}

func TestAssetsBagPlaceCoinAmount(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := newAssetsBag(client, cryptolibSigner)
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

	_, err = iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
				ptb,
				l1starter.ISCPackageID(),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagMainRef}),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coinRef}),
				10,
				testCointype,
			)
		},
	)
	require.NoError(t, err)
}

func TestAssetsBagTakeCoinBalanceMergeTo(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()
	const topUpAmount = 123
	txnResponse, err := newAssetsBag(client, cryptolibSigner)
	require.NoError(t, err)
	assetsBagMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)
	mergeToCoin1 := getCoinsRes.Data[2]

	_, err = assetsBagPlaceCoinAmount(
		client,
		cryptolibSigner,
		assetsBagMainRef,
		getCoinsRes.Data[1].Ref(),
		iotajsonrpc.IotaCoinType,
		1000,
	)
	require.NoError(t, err)

	assetsBagMainRef, err = client.UpdateObjectRef(context.Background(), assetsBagMainRef)
	require.NoError(t, err)

	txnResponse, err = iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:      client,
			Signer:      cryptolibSigner,
			PackageID:   l1starter.ISCPackageID(),
			GasPayments: []*iotago.ObjectRef{mergeToCoin1.Ref()},
			GasPrice:    iotaclient.DefaultGasPrice,
			GasBudget:   iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagTakeCoinBalanceMergeTo(
				ptb,
				l1starter.ISCPackageID(),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagMainRef}),
				topUpAmount,
				iotajsonrpc.IotaCoinType,
			)
		},
	)
	require.NoError(t, err)

	getObjRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: mergeToCoin1.CoinObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowBcs: true},
	})
	require.NoError(t, err)
	var mergeToCoin2 iscmoveclient.MoveCoin
	err = iotaclient.UnmarshalBCS(getObjRes.Data.Bcs.Data.MoveObject.BcsBytes, &mergeToCoin2)
	require.NoError(t, err)
	require.Equal(t, mergeToCoin1.Balance.Int64()-txnResponse.Effects.Data.GasFee()+topUpAmount, int64(mergeToCoin2.Balance))
}

func TestGetAssetsBagFromAssetsBagID(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	txnResponse, err := iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagNewAndTransfer(ptb, l1starter.ISCPackageID(), cryptolibSigner.Address())
		},
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
	_, err = iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagPlaceCoin(
				ptb,
				l1starter.ISCPackageID(),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagMainRef}),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coinRef}),
				string(testCointype),
			)
		},
	)
	require.NoError(t, err)

	assetsBag, err := client.GetAssetsBagWithBalances(context.Background(), assetsBagMainRef.ObjectID)
	require.NoError(t, err)
	require.Equal(t, *assetsBagMainRef.ObjectID, assetsBag.ID)
	require.Equal(t, uint64(1), assetsBag.Size)
	bal, ok := assetsBag.Balances[testCointype]
	require.True(t, ok)
	require.Equal(t, testCointype, bal.CoinType)
	require.Equal(t, uint64(1000000), bal.TotalBalance.Uint64())
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
	require.Equal(t, uint64(2), assetsBag.Size)
	bal, ok := assetsBag.Balances[testCointype]
	require.True(t, ok)
	require.Equal(t, testCointype, bal.CoinType)
	require.Equal(t, uint64(1000000), bal.TotalBalance.Uint64())
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

	execRes, err := client.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects: true,
			},
		},
	)
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

	txnResponse, err := newAssetsBag(client, cryptolibSigner)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	_, err = iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    cryptolibSigner,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagPlaceCoin(
				ptb,
				l1starter.ISCPackageID(),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coinRef}),
				string(testCointype),
			)
		},
	)
	require.NoError(t, err)

	assetsBagGetObjectRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: assetsBagRef.ObjectID})
	require.NoError(t, err)
	tmpAssetsBagRef := assetsBagGetObjectRes.Data.Ref()

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		&iscmoveclient.CreateAndSendRequestRequest{
			Signer:        cryptolibSigner,
			PackageID:     l1starter.ISCPackageID(),
			AnchorAddress: anchor.ObjectID,
			AssetsBagRef:  &tmpAssetsBagRef,
			Message:       iscmovetest.RandomMessage(),
			Allowance: &iscmove.Assets{
				Coins: iscmove.CoinBalances{
					"0x1::iota::IOTA":    11,
					"0xa::testa::TEST_A": 12,
				},
			},
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
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
	require.Equal(t, uint64(1000000), bal.TotalBalance.Uint64())
}

func newAssetsBag(
	client *iscmoveclient.Client,
	signer cryptolib.Signer,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    signer,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagNewAndTransfer(ptb, l1starter.ISCPackageID(), signer.Address())
		},
	)
}

func assetsBagPlaceCoinAmount(
	client *iscmoveclient.Client,
	signer cryptolib.Signer,
	assetsBagRef *iotago.ObjectRef,
	coin *iotago.ObjectRef,
	coinType iotajsonrpc.CoinType,
	amount uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return iscmoveclienttest.PTBTestWrapper(
		&iscmoveclienttest.PTBTestWrapperRequest{
			Client:    client,
			Signer:    signer,
			PackageID: l1starter.ISCPackageID(),
			GasPrice:  iotaclient.DefaultGasPrice,
			GasBudget: iotaclient.DefaultGasBudget,
		},
		func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder {
			return iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
				ptb,
				l1starter.ISCPackageID(),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coin}),
				amount,
				string(coinType),
			)
		},
	)
}
