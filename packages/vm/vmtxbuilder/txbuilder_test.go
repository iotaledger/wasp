package vmtxbuilder_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
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
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestTxBuilderBasic(t *testing.T) {
	client := l1starter.Instance().L1Client()
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	senderSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)
	iscPackage, err := client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(chainSigner))
	require.NoError(t, err)

	anchor, err := client.L2().StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        chainSigner,
			AnchorOwner:   chainSigner.Address(),
			PackageID:     iscPackage,
			StateMetadata: []byte{1, 2, 3, 4},
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)
	selectedGasCoin := getCoinsRes.Data[0].Ref()

	stateAnchor := isc.NewStateAnchor(anchor, iscPackage)
	txb := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, &stateAnchor, chainSigner.Address())

	req1 := createIscmoveReq(t, client, senderSigner, iscPackage, anchor)
	txb.ConsumeRequest(req1)
	req2 := createIscmoveReq(t, client, senderSigner, iscPackage, anchor)
	txb.ConsumeRequest(req2)
	stateMetadata := []byte("dummy stateMetadata")
	pt := txb.BuildTransactionEssence(stateMetadata, 123)

	tx := iotago.NewProgrammable(
		chainSigner.Address().AsIotaAddress(),
		pt,
		[]*iotago.ObjectRef{selectedGasCoin},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      cryptolib.SignerToIotaSigner(chainSigner),
			Options:     &iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
		},
	)

	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: req1.RequestRef().ObjectID, Options: &iotajsonrpc.IotaObjectDataOptions{ShowContent: true}})
	require.NotNil(t, getObjReq1.Error.Data.Deleted)
	getObjReq2, _ := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: req2.RequestRef().ObjectID})
	require.NotNil(t, getObjReq2.Error.Data.Deleted)
}

func TestTxBuilderSendAssetsAndRequest(t *testing.T) {
	client := l1starter.Instance().L1Client()
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	senderSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)
	recipientSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 2)
	iscPackage, err := client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(chainSigner))
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	anchor, err := client.L2().StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        chainSigner,
			AnchorOwner:   chainSigner.Address(),
			PackageID:     iscPackage,
			StateMetadata: []byte{1, 2, 3, 4},
			InitCoinRef:   getCoinsRes.Data[1].Ref(),
			GasPayments:   []*iotago.ObjectRef{getCoinsRes.Data[0].Ref()},
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	selectedGasCoin := getCoinsRes.Data[2].Ref()
	stateAnchor := isc.NewStateAnchor(anchor, iscPackage)
	txb1 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, &stateAnchor, chainSigner.Address())

	req1 := createIscmoveReq(t, client, senderSigner, iscPackage, anchor)
	txb1.ConsumeRequest(req1)

	// stateMetadata := transaction.NewStateMetadata(isc.SchemaVersion(1), commitment, &gas.FeePolicy{}, isc.CallArguments{}, "http://dummy")
	// ptb := txb.BuildTransactionEssence(stateMetadata.Bytes())
	stateMetadata1 := []byte("dummy stateMetadata1")
	ptb1 := txb1.BuildTransactionEssence(stateMetadata1, 123)

	tx1 := iotago.NewProgrammable(
		chainSigner.Address().AsIotaAddress(),
		ptb1,
		[]*iotago.ObjectRef{selectedGasCoin},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txnBytes1, err := bcs.Marshal(&tx1)
	require.NoError(t, err)

	txnResponse1, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes1,
			Signer:      cryptolib.SignerToIotaSigner(chainSigner),
			Options:     &iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
		},
	)

	require.NoError(t, err)
	require.True(t, txnResponse1.Effects.Data.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: req1.RequestRef().ObjectID, Options: &iotajsonrpc.IotaObjectDataOptions{ShowContent: true}})
	require.NotNil(t, getObjReq1.Error.Data.Deleted)

	// reset
	tmp, err := client.UpdateObjectRef(context.Background(), &anchor.ObjectRef)
	require.NoError(t, err)
	anchor.ObjectRef = *tmp
	txb2 := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, &stateAnchor, chainSigner.Address())

	txb2.SendAssets(recipientSigner.Address().AsIotaAddress(), isc.NewAssets(1))

	req2 := createIscmoveReq(t, client, senderSigner, iscPackage, anchor)
	txb2.ConsumeRequest(req2)

	zeroAssets := iscmove.NewEmptyAssets()
	req3 := createIscmoveReqWithAssets(t, client, senderSigner, iscPackage, anchor, zeroAssets)
	txb2.ConsumeRequest(req3)
	stateMetadata2 := []byte("dummy stateMetadata2")
	pt2 := txb2.BuildTransactionEssence(stateMetadata2, 123)

	getCoinsRes, err = client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	tx2 := iotago.NewProgrammable(
		chainSigner.Address().AsIotaAddress(),
		pt2,
		[]*iotago.ObjectRef{getCoinsRes.Data[0].Ref()},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txnBytes2, err := bcs.Marshal(&tx2)
	require.NoError(t, err)

	txnResponse2, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes2,
			Signer:      cryptolib.SignerToIotaSigner(chainSigner),
			Options:     &iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
		},
	)

	require.NoError(t, err)
	require.True(t, txnResponse2.Effects.Data.IsSuccess())

	getObjReq2, _ := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: req2.RequestRef().ObjectID})
	require.NotNil(t, getObjReq2.Error.Data.Deleted)
}

func TestRotateAndBuildTx(t *testing.T) {
	client := l1starter.Instance().L1Client()
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	senderSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)
	rotateRecipientSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 2)
	iscPackage, err := client.DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(chainSigner))
	require.NoError(t, err)

	anchor, err := client.L2().StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        chainSigner,
			AnchorOwner:   chainSigner.Address(),
			PackageID:     iscPackage,
			StateMetadata: []byte{1, 2, 3, 4},
			GasPrice:      iotaclient.DefaultGasPrice,
			GasBudget:     iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	selectedGasCoin := getCoinsRes.Data[0].Ref()

	stateAnchor := isc.NewStateAnchor(anchor, iscPackage)
	txb := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, &stateAnchor, chainSigner.Address())

	req1 := createIscmoveReq(t, client, senderSigner, iscPackage, anchor)
	txb.ConsumeRequest(req1)
	req2 := createIscmoveReq(t, client, senderSigner, iscPackage, anchor)
	txb.ConsumeRequest(req2)
	txb.RotationTransaction(rotateRecipientSigner.Address().AsIotaAddress())
	stateMetadata := []byte("dummy stateMetadata")
	pt := txb.BuildTransactionEssence(stateMetadata, 123)

	tx := iotago.NewProgrammable(
		chainSigner.Address().AsIotaAddress(),
		pt,
		[]*iotago.ObjectRef{selectedGasCoin},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes,
			Signer:      cryptolib.SignerToIotaSigner(chainSigner),
			Options:     &iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
		},
	)

	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: req1.RequestRef().ObjectID, Options: &iotajsonrpc.IotaObjectDataOptions{ShowContent: true}})
	require.NotNil(t, getObjReq1.Error.Data.Deleted)
	getObjReq2, _ := client.GetObject(context.Background(), iotaclient.GetObjectRequest{ObjectID: req2.RequestRef().ObjectID})
	require.NotNil(t, getObjReq2.Error.Data.Deleted)

	getObjRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: anchor.ObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, rotateRecipientSigner.Address().AsIotaAddress(), getObjRes.Data.Owner.AddressOwner)

	gasCoinGetObjRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: selectedGasCoin.ObjectID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, rotateRecipientSigner.Address().AsIotaAddress(), gasCoinGetObjRes.Data.Owner.AddressOwner)
}

func createIscmoveReq(
	t *testing.T,
	client clients.L1Client,
	signer cryptolib.Signer,
	iscPackage iotago.Address,
	anchor *iscmove.AnchorWithRef,
) isc.OnLedgerRequest {
	err := iotaclient.RequestFundsFromFaucet(context.Background(), signer.Address().AsIotaAddress(), l1starter.Instance().FaucetURL())
	require.NoError(t, err)

	createAndSendRequestRes, err := client.L2().CreateAndSendRequestWithAssets(
		context.Background(),
		&iscmoveclient.CreateAndSendRequestWithAssetsRequest{
			Signer:           signer,
			PackageID:        iscPackage,
			AnchorAddress:    anchor.ObjectID,
			Assets:           iscmove.NewAssets(111),
			Message:          iscmovetest.RandomMessage(),
			AllowanceBCS:     nil,
			OnchainGasBudget: 100,
			GasPrice:         iotaclient.DefaultGasPrice,
			GasBudget:        iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	reqRef, err := createAndSendRequestRes.GetCreatedObjectByName(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)
	reqWithObj, err := client.L2().GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
	require.NoError(t, err)
	req, err := isc.OnLedgerFromMoveRequest(reqWithObj, cryptolib.NewAddressFromIota(anchor.ObjectID))
	require.NoError(t, err)

	return req
}

func createIscmoveReqWithAssets(
	t *testing.T,
	client clients.L1Client,
	signer cryptolib.Signer,
	iscPackage iotago.Address,
	anchor *iscmove.AnchorWithRef,
	assets *iscmove.Assets,
) isc.OnLedgerRequest {
	err := iotaclient.RequestFundsFromFaucet(context.Background(), signer.Address().AsIotaAddress(), l1starter.Instance().FaucetURL())
	require.NoError(t, err)

	createAndSendRequestRes, err := client.L2().CreateAndSendRequestWithAssets(
		context.Background(),
		&iscmoveclient.CreateAndSendRequestWithAssetsRequest{
			Signer:           signer,
			PackageID:        iscPackage,
			AnchorAddress:    anchor.ObjectID,
			Assets:           assets,
			Message:          iscmovetest.RandomMessage(),
			AllowanceBCS:     lo.Must(bcs.Marshal(assets)),
			OnchainGasBudget: 100,
			GasPrice:         iotaclient.DefaultGasPrice,
			GasBudget:        iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(t, err)
	reqRef, err := createAndSendRequestRes.GetCreatedObjectByName(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(t, err)
	reqWithObj, err := client.L2().GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
	require.NoError(t, err)
	req, err := isc.OnLedgerFromMoveRequest(reqWithObj, cryptolib.NewAddressFromIota(anchor.ObjectID))
	require.NoError(t, err)

	return req
}
