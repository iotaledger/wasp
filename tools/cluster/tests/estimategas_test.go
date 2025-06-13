package tests

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
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

	createTx := func(l1GasBudget, l2GasBudget uint64) []byte {
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
			iotajsonrpc.CoinValue(l1GasBudget),
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
		allowanceVal := iotajsonrpc.CoinValue(1 * isc.Million)
		allowance := iscmove.NewAssets(allowanceVal)
		allowance.SetCoin(iotajsonrpc.MustCoinTypeFromString(testcoinType.String()), iotajsonrpc.CoinValue(10))

		ptb = iscmoveclient.PTBCreateAndSendRequest(
			ptb,
			l1starter.ISCPackageID(),
			env.Chain.ChainID.AsObjectID(),
			argAssetsBag,
			&iscmove.Message{
				Contract: uint32(isc.Hn("accounts")),
				Function: uint32(isc.Hn("deposit")),
			},
			bcs.MustMarshal(allowance),
			l2GasBudget,
		)

		pt := ptb.Finish()

		// Find proper coin objects to pay for gas
		coinsForGas, err := env.Clu.L1Client().FindCoinsForGasPayment(context.Background(), sender.Address().AsIotaAddress(), pt, iotaclient.DefaultGasPrice, l1GasBudget)
		require.NoError(t, err)

		txData := iotago.NewProgrammable(
			sender.Address().AsIotaAddress(),
			pt,
			coinsForGas,
			2*iotaclient.DefaultGasBudget, // TODO: l1GasBudget here fails test.
			iotaclient.DefaultGasPrice,
		)

		txBytes, err := bcs.Marshal(&txData)
		require.NoError(t, err)

		return txBytes
	}

	// Create transaction for estimation
	const minL1GasBudget = 1000000
	txBytesForEstimation := createTx(minL1GasBudget, 0)
	// TODO:
	// Ideally the this call should look like this: createTx(0, 0).
	// But if we pass 0 as l1 budget we get error upon estimation:
	//    Error checking transaction input objects: GasBudgetTooLow { gas_budget: 0, min_budget: 1000000 }"
	// If we try to pass math.MaxUint64, we get an error upon trying to find coins for gas:
	//    insufficient account balance
	// So we need to specify some meaningful value of gas budget for estimation. But this becomes chicken-and-egg problem.
	// So I assumed that 1000000 is some theoretical min budget for any transaction, because if we remove one of requests from the transaction,
	// the estimated gas still stays at 1000000.

	// Estimate L1 and L2 gas budget for that transaction
	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOnledger(context.Background()).Request(apiclient.EstimateGasRequestOnledger{
		TransactionBytes: hexutil.Encode(txBytesForEstimation),
	}).Execute()
	if err != nil {
		var msg string
		if _, ok := err.(*apiclient.GenericOpenAPIError); ok {
			msg = string(err.(*apiclient.GenericOpenAPIError).Body())
		}
		require.NoError(t, err, msg)
	}
	require.Empty(t, estimatedReceipt.L2.ErrorMessage, lo.FromPtr(estimatedReceipt.L2.ErrorMessage))

	l1GasBudget, err := strconv.ParseUint(estimatedReceipt.L1.GasFeeCharged, 10, 64)
	require.NoError(t, err)
	l2GasBudget, err := strconv.ParseUint(estimatedReceipt.L2.GasFeeCharged, 10, 64)
	require.NoError(t, err)

	executeTx := func(txBytes []byte) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
		execRes, err := env.Clu.L1Client().SignAndExecuteTransaction(context.Background(), &iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txBytes,
			Signer:      cryptolib.SignerToIotaSigner(sender),
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:        true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
		})
		return execRes, err
	}

	// TODO: This check won't work, because right now l1GasBudget == minL1GasBudget, so L1 tx actually executes successfully.
	//
	// Checking that execution fails with zero L1 and L2 gas budgets
	// res := executeTx(txBytesForEstimation)
	// require.NotEmpty(t, res.Errors)

	// TODO: This check does not work because we have 2*iotaclient.DefaultGasBudget in NewProgrammable.
	//
	// Checking that transaction execution fails with wrong L1 gas budget
	// txBytesWithWrongL1GasBudget := createTx(l1GasBudget-1, l2GasBudget)
	// res, err := executeTx(txBytesWithWrongL1GasBudget)
	// require.Error(t, err)

	// TODO: Should we make an attempt to execute tx with wrong L2 gas budget?
	//       How then will it work in relation to next attempt? The request will stuck and then unstuck?
	//
	// Checking that transaction execution fails with wrong L2 gas budget
	// txBytesWithWrongL2GasBudget := createTx(l1GasBudget, l2GasBudget-1)
	// res = executeTx(txBytesWithWrongL2GasBudget)
	// require.NotEmpty(t, res.Errors)

	// Executing transaction with proper L1 and L2 gas budgets
	txBytes := createTx(l1GasBudget, l2GasBudget)
	res, err := executeTx(txBytes)
	require.NoError(t, err)
	require.Empty(t, res.Errors)
	var totalL1GasUsed big.Int
	totalL1GasUsed.Add(&totalL1GasUsed, res.Effects.Data.V1.GasUsed.ComputationCost.Int)
	totalL1GasUsed.Add(&totalL1GasUsed, res.Effects.Data.V1.GasUsed.StorageCost.Int)
	totalL1GasUsed.Sub(&totalL1GasUsed, res.Effects.Data.V1.GasUsed.StorageRebate.Int)
	require.Equal(t, estimatedReceipt.L1.GasFeeCharged, totalL1GasUsed.String())

	recs, err := env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, res, false, 10*time.Second)
	require.NoError(t, err)
	require.Empty(t, recs[0].ErrorMessage, lo.FromPtr(recs[0].ErrorMessage))
	require.Equal(t, recs[0].GasBurned, estimatedReceipt.L2.GasBurned)
	require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.L2.GasFeeCharged)
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
