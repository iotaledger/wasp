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
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	// commitment, err := transaction.L1CommitmentFromAnchor(anchor.Object)
	// require.NoError(t, err)

	txb := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, anchor)

	req1 := createIscmoveReq(t, client, signer, iscPackage, anchor)
	txb.ConsumeRequest(req1)
	req2 := createIscmoveReq(t, client, signer, iscPackage, anchor)
	txb.ConsumeRequest(req2)
	// stateMetadata := transaction.NewStateMetadata(isc.SchemaVersion(1), commitment, &gas.FeePolicy{}, isc.CallArguments{}, "http://dummy")
	// ptb := txb.BuildTransactionEssence(stateMetadata.Bytes())
	stateMetadata := []byte("dummy stateMetadata")
	ptb := txb.BuildTransactionEssence(stateMetadata)

	coins, err := client.GetCoinObjsForTargetAmount(context.Background(), signer.Address().AsSuiAddress(), suiclient.DefaultGasBudget)
	require.NoError(t, err)

	tx := sui.NewProgrammable(
		signer.Address().AsSuiAddress(),
		ptb,
		coins.CoinRefs(),
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
	txnResponse, err := client.AssetsBagNew(
		context.Background(),
		signer,
		iscPackage,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	assetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
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
