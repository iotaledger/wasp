package vmtxbuilder_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/vm/vmtxbuilder"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestTxBuilderBasic(t *testing.T) {
	client := l1starter.Instance().L1Client()
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	senderSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)
	iscPackage, err := client.L2().DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(chainSigner))
	require.NoError(t, err)

	anchor, err := client.L2().StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        chainSigner,
			AnchorOwner:   chainSigner.Address(),
			PackageID:     iscPackage,
			StateMetadata: []byte{1, 2, 3, 4},
			GasPrice:      iotagraphql.DefaultGasPrice,
			GasBudget:     iotagraphql.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotagraphql.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)
	selectedGasCoin, err := getCoinsRes.Address.Coins.Nodes[0].ObjectRef()
	require.NoError(t, err)

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
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes,
		cryptolib.SignerToIotaSigner(chainSigner),
	)
	require.NoError(t, err)
	require.True(t, txnResponse.ExecuteTransactionBlock.Effects.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), *req1.RequestRef().ObjectID)
	require.NotNil(t, getObjReq1)
	require.Equal(t, graphqltypes.ObjectKindWrappedOrDeleted, getObjReq1.Object.Status)
	getObjReq2, _ := client.GetObject(context.Background(), *req2.RequestRef().ObjectID)
	require.NotNil(t, getObjReq2)
	require.Equal(t, graphqltypes.ObjectKindWrappedOrDeleted, getObjReq2.Object.Status)
}

func TestTxBuilderSendAssetsAndRequest(t *testing.T) {
	client := l1starter.Instance().L1Client()
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	senderSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 1)
	recipientSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 2)
	iscPackage, err := client.L2().DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(chainSigner))
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotagraphql.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	anchor, err := client.L2().StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        chainSigner,
			AnchorOwner:   chainSigner.Address(),
			PackageID:     iscPackage,
			StateMetadata: []byte{1, 2, 3, 4},
			InitCoinRef:   lo.Must(getCoinsRes.Address.Coins.Nodes[1].ObjectRef()),
			GasPayments:   []*iotago.ObjectRef{lo.Must(getCoinsRes.Address.Coins.Nodes[0].ObjectRef())},
			GasPrice:      iotagraphql.DefaultGasPrice,
			GasBudget:     iotagraphql.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	selectedGasCoin := lo.Must(getCoinsRes.Address.Coins.Nodes[2].ObjectRef())
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
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txnBytes1, err := bcs.Marshal(&tx1)
	require.NoError(t, err)

	txnResponse1, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes1,
		cryptolib.SignerToIotaSigner(chainSigner),
	)
	require.NoError(t, err)
	require.True(t, txnResponse1.ExecuteTransactionBlock.Effects.IsSuccess())

	getObjReq1, _ := client.GetObject(context.Background(), *req1.RequestRef().ObjectID)
	require.NotNil(t, getObjReq1)
	require.Equal(t, graphqltypes.ObjectKindWrappedOrDeleted, getObjReq1.Object.Status)

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

	getCoinsRes, err = client.GetCoins(context.Background(), iotagraphql.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	tx2 := iotago.NewProgrammable(
		chainSigner.Address().AsIotaAddress(),
		pt2,
		[]*iotago.ObjectRef{lo.Must(getCoinsRes.Address.Coins.Nodes[0].ObjectRef())},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txnBytes2, err := bcs.Marshal(&tx2)
	require.NoError(t, err)

	txnResponse2, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes2,
		cryptolib.SignerToIotaSigner(chainSigner),
	)
	require.NoError(t, err)
	require.True(t, txnResponse2.ExecuteTransactionBlock.Effects.IsSuccess())

	getObjReq2, _ := client.GetObject(context.Background(), *req2.RequestRef().ObjectID)
	require.NotNil(t, getObjReq2)
	require.Equal(t, graphqltypes.ObjectKindWrappedOrDeleted, getObjReq2.Object.Status)
}

func TestRotateAndBuildTx(t *testing.T) {
	client := l1starter.Instance().L1Client()
	chainSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	rotateRecipientSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 2)
	iscPackage, err := client.L2().DeployISCContracts(context.Background(), cryptolib.SignerToIotaSigner(chainSigner))
	require.NoError(t, err)

	anchor, err := client.L2().StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:        chainSigner,
			AnchorOwner:   chainSigner.Address(),
			PackageID:     iscPackage,
			StateMetadata: []byte{1, 2, 3, 4},
			GasPrice:      iotagraphql.DefaultGasPrice,
			GasBudget:     iotagraphql.DefaultGasBudget,
		},
	)
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotagraphql.GetCoinsRequest{Owner: chainSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	selectedGasCoin := lo.Must(getCoinsRes.Address.Coins.Nodes[0].ObjectRef())

	stateAnchor := isc.NewStateAnchor(anchor, iscPackage)
	txb := vmtxbuilder.NewAnchorTransactionBuilder(iscPackage, &stateAnchor, chainSigner.Address())

	txb.RotationTransaction(rotateRecipientSigner.Address().AsIotaAddress())
	stateMetadata := []byte("dummy stateMetadata")
	pt := txb.BuildTransactionEssence(stateMetadata, 123)

	tx := iotago.NewProgrammable(
		chainSigner.Address().AsIotaAddress(),
		pt,
		[]*iotago.ObjectRef{selectedGasCoin},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txnBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes,
		cryptolib.SignerToIotaSigner(chainSigner),
	)
	require.NoError(t, err)
	require.True(t, txnResponse.ExecuteTransactionBlock.Effects.IsSuccess())
	getObjRes, err := client.GetObject(context.Background(), *anchor.ObjectID)
	require.NoError(t, err)
	require.Equal(t, rotateRecipientSigner.Address().AsIotaAddress(), getObjRes.Object.OwnerAddress())

	gasCoinGetObjRes, err := client.GetObject(context.Background(), *selectedGasCoin.ObjectID)
	require.NoError(t, err)
	require.Equal(t, rotateRecipientSigner.Address().AsIotaAddress(), gasCoinGetObjRes.Object.OwnerAddress())
}

func createIscmoveReq(
	t *testing.T,
	client clients.L1Client,
	signer cryptolib.Signer,
	iscPackage iotago.Address,
	anchor *iscmove.AnchorWithRef,
) isc.OnLedgerRequest {
	err := l1starter.Instance().L1Client().RequestFundsFromFaucet(context.Background(), *signer.Address().AsIotaAddress())
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
			GasPrice:         iotagraphql.DefaultGasPrice,
			GasBudget:        iotagraphql.DefaultGasBudget,
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
	err := l1starter.Instance().L1Client().RequestFundsFromFaucet(context.Background(), *signer.Address().AsIotaAddress())
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
			GasPrice:         iotagraphql.DefaultGasPrice,
			GasBudget:        iotagraphql.DefaultGasBudget,
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
