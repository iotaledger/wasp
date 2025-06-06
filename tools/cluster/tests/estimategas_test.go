package tests

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func testEstimateGasOnLedger(t *testing.T, env *ChainEnv) {
	// Create tx sender with some funds
	sender := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	env.DepositFunds(100*isc.Million, sender.(*cryptolib.KeyPair))

	// Mind some TESTCOINs
	coinPackageID, treasuryCap := iotaclienttest.DeployCoinPackage(
		t,
		env.Clu.L1Client().IotaClient(),
		cryptolib.SignerToIotaSigner(sender),
		contracts.Testcoin(),
	)
	testcoinType := coin.MustTypeFromString(fmt.Sprintf(
		"%s::%s::%s",
		coinPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	))
	testcoinRef := iotaclienttest.MintCoins(
		t,
		env.Clu.L1Client().IotaClient(),
		cryptolib.SignerToIotaSigner(sender),
		coinPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap,
		1*isc.Million,
	)

	ptb := iotago.NewProgrammableTransactionBuilder()

	// Create asset bag for transaction
	ptb = iscmoveclient.PTBAssetsBagNew(ptb, l1starter.ISCPackageID(), sender.Address())
	argAssetsBag := ptb.LastCommandResultArg()

	// Place some IOTAs into new asset bag for gas payment
	ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
		ptb,
		l1starter.ISCPackageID(),
		argAssetsBag,
		iotago.GetArgumentGasCoin(),
		iotajsonrpc.CoinValue(iotaclient.DefaultGasBudget),
		iotajsonrpc.IotaCoinType,
	)

	// Place some TESTCOINs into new asset bag
	ptb = iscmoveclient.PTBAssetsBagPlaceCoinWithAmount(
		ptb,
		l1starter.ISCPackageID(),
		argAssetsBag,
		ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: testcoinRef}),
		iotajsonrpc.CoinValue(122),
		iotajsonrpc.CoinType(testcoinType.String()),
	)

	// Deposit funds from L1 asset bag into L2 account
	msg := &iscmove.Message{
		Contract: uint32(isc.Hn("accounts")),
		Function: uint32(isc.Hn("deposit")),
	}
	allowanceVal := iotajsonrpc.CoinValue(1 * isc.Million)
	allowance := iscmove.NewAssets(allowanceVal)
	allowance.SetCoin(iotajsonrpc.MustCoinTypeFromString(testcoinType.String()), iotajsonrpc.CoinValue(10))
	l2GasBudget := uint64(100)
	ptb = iscmoveclient.PTBCreateAndSendRequest(
		ptb,
		l1starter.ISCPackageID(),
		env.Chain.ChainID.AsObjectID(),
		argAssetsBag,
		msg,
		bcs.MustMarshal(allowance),
		l2GasBudget,
	)
	pt := ptb.Finish()

	coinsForGas, err := env.Clu.L1Client().FindCoinsForGasPayment(context.Background(), sender.Address().AsIotaAddress(), pt, iotaclient.DefaultGasPrice, 2*iotaclient.DefaultGasBudget)
	require.NoError(t, err)

	txData := iotago.NewProgrammable(
		sender.Address().AsIotaAddress(),
		pt,
		coinsForGas,
		2*iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	// Estimate L2 gas budget for that transaction
	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOnledger(context.Background()).Request(apiclient.EstimateGasRequestOnledger{
		TransactionBytes: hexutil.Encode(txBytes),
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	l2GasBudget, err = strconv.ParseUint(estimatedReceipt.GasFeeCharged, 10, 64)
	require.NoError(t, err)

	// Execute same transaction, but with proper L2 gas budget
	execRes, err := env.Clu.L1Client().SignAndExecuteTransaction(context.Background(), &iotaclient.SignAndExecuteTransactionRequest{
		TxDataBytes: txBytes,
		Signer:      cryptolib.SignerToIotaSigner(sender),
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:        true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
		},
	})
	require.NoError(t, err)

	recs, err := env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, execRes, false, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, recs[0].GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.GasFeeCharged)
}

func testEstimateGasOffLedger(t *testing.T, env *ChainEnv) {
	// estimate off-ledger request, then send the same request, assert the gas used/fees match
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	env.DepositFunds(10*isc.Million, keyPair)

	estimationReq := isc.NewOffLedgerRequest(
		env.Chain.ChainID,
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		0,
		1*isc.Million,
	).WithAllowance(isc.NewAssets(5000)).
		WithSender(keyPair.GetPublicKey())

	// Test that the API will fail if the FromAddress is missing
	estimatedReceiptFail, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOffledger(context.Background()).Request(apiclient.EstimateGasRequestOffledger{
		RequestBytes: cryptolib.EncodeHex(estimationReq.Bytes()),
	}).Execute()
	require.Error(t, err)
	require.Nil(t, estimatedReceiptFail)
	///

	requestHex := cryptolib.EncodeHex(estimationReq.Bytes())

	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOffledger(context.Background()).Request(apiclient.EstimateGasRequestOffledger{
		FromAddress:  keyPair.Address().String(),
		RequestBytes: requestHex,
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	client := env.Chain.Client(keyPair)
	par := chainclient.PostRequestParams{
		Allowance:   isc.NewAssets(5000),
		GasBudget:   iotaclient.DefaultGasBudget,
		L2GasBudget: 1 * isc.Million,
	}
	req, err := client.PostOffLedgerRequest(
		context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		par,
	)
	require.NoError(t, err)
	rec, err := env.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), env.Chain.ChainID, req.ID(), false, 30*time.Second)
	require.NoError(t, err)
	require.Equal(t, rec.GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, rec.GasFeeCharged, estimatedReceipt.GasFeeCharged)
}
