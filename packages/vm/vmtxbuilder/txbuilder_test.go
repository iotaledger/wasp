package vmtxbuilder_test

import (
	"context"
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func TestTxBuilderBasic(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)
	iscPackage := buildAndDeployISCContracts(t, client, signer)

	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackage,
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	txb := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, anchor, signer.Address())

	req1 := createIscmoveReq(t, client, signer, iscPackage, anchor)
	txb.ConsumeRequest(req1)
	req2 := createIscmoveReq(t, client, signer, iscPackage, anchor)
	txb.ConsumeRequest(req2)
	// stateMetadata := transaction.NewStateMetadata(isc.SchemaVersion(1), commitment, &gas.FeePolicy{}, isc.CallArguments{}, "http://dummy")
	// ptb := txb.BuildTransactionEssence(stateMetadata.Bytes())
	stateMetadata := []byte("dummy stateMetadata")
	pt := txb.BuildTransactionEssence(stateMetadata)

	getCoinsRes, err := client.GetCoins(context.Background(), suiclient.GetCoinsRequest{Owner: signer.Address().AsSuiAddress()})
	require.NoError(t, err)

	tx := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt,
		[]*sui.ObjectRef{getCoinsRes.Data[len(getCoinsRes.Data)-2].Ref()},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: req1.RequestRef().ObjectID, Options: &suijsonrpc.SuiObjectDataOptions{ShowContent: true}})
	require.NotNil(t, getObjReq1.Error.Data.Deleted)
	getObjReq2, _ := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: req2.RequestRef().ObjectID})
	require.NotNil(t, getObjReq2.Error.Data.Deleted)
}

func TestTxBuilderSendAssetsAndRequest(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)
	recipient := newSignerWithFunds(t, testSeed, 1)
	iscPackage := buildAndDeployISCContracts(t, client, signer)

	anchor, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackage,
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	txb1 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, anchor, signer.Address())

	req1 := createIscmoveReq(t, client, signer, iscPackage, anchor)
	txb1.ConsumeRequest(req1)

	// stateMetadata := transaction.NewStateMetadata(isc.SchemaVersion(1), commitment, &gas.FeePolicy{}, isc.CallArguments{}, "http://dummy")
	// ptb := txb.BuildTransactionEssence(stateMetadata.Bytes())
	stateMetadata1 := []byte("dummy stateMetadata1")
	ptb1 := txb1.BuildTransactionEssence(stateMetadata1)

	coins, err := client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsSuiAddress(), suiclient.DefaultGasBudget)
	require.NoError(t, err)

	tx1 := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		ptb1,
		coins.CoinRefs(),
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txnBytes1, err := bcs.Marshal(&tx1)
	require.NoError(t, err)

	txnResponse1, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes1,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, txnResponse1.Effects.Data.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: req1.RequestRef().ObjectID, Options: &suijsonrpc.SuiObjectDataOptions{ShowContent: true}})
	require.NotNil(t, getObjReq1.Error.Data.Deleted)

	// reset
	tmp, err := client.UpdateObjectRef(context.Background(), &anchor.ObjectRef)
	require.NoError(t, err)
	anchor.ObjectRef = *tmp
	txb2 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, anchor, signer.Address())

	txb2.SendAssets(recipient.Address().AsSuiAddress(), isc.NewAssets(1))

	req2 := createIscmoveReq(t, client, signer, iscPackage, anchor)
	txb2.ConsumeRequest(req2)
	stateMetadata2 := []byte("dummy stateMetadata2")
	pt2 := txb2.BuildTransactionEssence(stateMetadata2)

	coins, err = client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsSuiAddress(), suiclient.DefaultGasBudget)
	require.NoError(t, err)

	tx2 := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt2,
		coins.CoinRefs(),
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txnBytes2, err := bcs.Marshal(&tx2)
	require.NoError(t, err)

	txnResponse2, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes2,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, txnResponse2.Effects.Data.IsSuccess())

	getObjReq2, _ := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: req2.RequestRef().ObjectID})
	require.NotNil(t, getObjReq2.Error.Data.Deleted)
}

func TestTxBuilderSendCrossChainRequest(t *testing.T) {
	client := newLocalnetClient()
	signer := newSignerWithFunds(t, testSeed, 0)
	iscPackage1 := buildAndDeployISCContracts(t, client, signer)

	anchor1, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackage1,
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	anchor2, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackage1,
		[]byte{1, 2, 3, 4},
		nil,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	txb1 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage1, anchor1, signer.Address())

	req1 := createIscmoveReq(t, client, signer, iscPackage1, anchor1)
	txb1.ConsumeRequest(req1)

	stateMetadata1 := []byte("dummy stateMetadata1")
	pt1 := txb1.BuildTransactionEssence(stateMetadata1)

	coins, err := client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsSuiAddress(), suiclient.DefaultGasBudget)
	require.NoError(t, err)

	tx1 := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt1,
		[]*sui.ObjectRef{coins.CoinRefs()[2]},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txnBytes1, err := bcs.Marshal(&tx1)
	require.NoError(t, err)

	txnResponse1, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes1,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, txnResponse1.Effects.Data.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: req1.RequestRef().ObjectID, Options: &suijsonrpc.SuiObjectDataOptions{ShowContent: true}})
	require.NotNil(t, getObjReq1.Error.Data.Deleted)

	// reset
	tmp, err := client.UpdateObjectRef(context.Background(), &anchor1.ObjectRef)
	require.NoError(t, err)
	anchor1.ObjectRef = *tmp
	txb2 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage1, anchor1, signer.Address())

	txb2.SendCrossChainRequest(&iscPackage1, anchor2.ObjectID, isc.NewAssets(1), &isc.SendMetadata{
		Message:   isc.NewMessage(isc.Hn("accounts"), isc.Hn("deposit")),
		Allowance: isc.NewAssets(1),
		GasBudget: 2,
	})

	stateMetadata2 := []byte("dummy stateMetadata2")
	pt2 := txb2.BuildTransactionEssence(stateMetadata2)

	coins, err = client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsSuiAddress(), suiclient.DefaultGasBudget)
	require.NoError(t, err)

	tx2 := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt2,
		[]*sui.ObjectRef{coins.CoinRefs()[2]},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txnBytes2, err := bcs.Marshal(&tx2)
	require.NoError(t, err)

	txnResponse2, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes2,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, txnResponse2.Effects.Data.IsSuccess())
	crossChainRequestRef, err := txnResponse2.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)

	txb3 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage1, anchor2, signer.Address())

	reqWithObj, err := client.GetRequestFromObjectID(context.Background(), crossChainRequestRef.ObjectID)
	require.NoError(t, err)
	req3, err := isc.OnLedgerFromRequest(reqWithObj, cryptolib.NewAddressFromSui(anchor2.ObjectID))
	require.NoError(t, err)
	txb3.ConsumeRequest(req3)

	coins, err = client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsSuiAddress(), suiclient.DefaultGasBudget)
	require.NoError(t, err)

	stateMetadata3 := []byte("dummy stateMetadata3")
	pt3 := txb3.BuildTransactionEssence(stateMetadata3)

	tx3 := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt3,
		[]*sui.ObjectRef{coins.CoinRefs()[2]},
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)

	txnBytes3, err := bcs.Marshal(&tx3)
	require.NoError(t, err)

	txnResponse3, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(signer),
		txnBytes3,
		&suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
	)
	require.NoError(t, err)
	require.True(t, txnResponse3.Effects.Data.IsSuccess())
}

var testSeed = []byte{50, 230, 119, 9, 86, 155, 106, 30, 245, 81, 234, 122, 116, 90, 172, 148, 59, 33, 88, 252, 134, 42, 231, 198, 208, 141, 209, 116, 78, 21, 216, 24}

func newSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	seed[0] = seed[0] + byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seed))
	err := suiclient.RequestFundsFromFaucet(context.TODO(), kp.Address().AsSuiAddress(), suiconn.LocalnetFaucetURL)
	require.NoError(t, err)
	return kp
}

func newLocalnetClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		suiconn.LocalnetEndpointURL,
		suiconn.LocalnetFaucetURL,
	)
}

func buildAndDeployISCContracts(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) sui.PackageID {
	suiSigner := cryptolib.SignerToSuiSigner(signer)
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          suiSigner.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		suiSigner,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	return *packageID
}

func createIscmoveReq(
	t *testing.T,
	client *iscmoveclient.Client,
	signer cryptolib.Signer,
	iscPackage sui.Address,
	anchor *iscmove.AnchorWithRef,
) isc.OnLedgerRequest {
	err := suiclient.RequestFundsFromFaucet(context.Background(), signer.Address().AsSuiAddress(), suiconn.LocalnetFaucetURL)
	require.NoError(t, err)
	getCoinsRes, err := client.GetCoins(
		context.Background(),
		suiclient.GetCoinsRequest{
			Owner: signer.Address().AsSuiAddress(),
		},
	)
	require.NoError(t, err)
	_ = getCoinsRes
	assetsBagNewRes, err := client.AssetsBagNew(
		context.Background(),
		signer,
		iscPackage,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagRef, err := assetsBagNewRes.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)
	_, err = client.AssetsBagPlaceCoinAmount(
		context.Background(),
		signer,
		iscPackage,
		assetsBagRef,
		getCoinsRes.Data[len(getCoinsRes.Data)-1].Ref(),
		suijsonrpc.SuiCoinType,
		111,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagRef, err = client.UpdateObjectRef(context.Background(), assetsBagRef)
	require.NoError(t, err)

	createAndSendRequestRes, err := client.CreateAndSendRequest(
		context.Background(),
		signer,
		iscPackage,
		anchor.ObjectID,
		assetsBagRef,
		uint32(isc.Hn("test_isc_contract")),
		uint32(isc.Hn("test_isc_func")),
		[][]byte{[]byte("one"), []byte("two"), []byte("three")},
		nil,
		10,
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
	req, err := isc.OnLedgerFromRequest(reqWithObj, cryptolib.NewAddressFromSui(anchor.ObjectID))
	require.NoError(t, err)

	return req
}
