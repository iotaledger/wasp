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
	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func (e *ChainEnv) testEstimateGasOnLedger(t *testing.T) {
	// We decrease min gas per request, so that we can test L2 gas estimation negative cases.
	// Without this configuration using value of (l2GasBudget - 1) would still work, because in our
	// case l2GasBudget is lower then minGasPerRequest, so it is automatically increased to minGasPerRequest.
	govClient := e.Chain.Client(e.Clu.OriginatorKeyPair)
	tx, err := govClient.PostRequest(context.Background(), governance.FuncSetGasLimits.Message(&gas.Limits{
		MinGasPerRequest:       1,
		MaxGasPerBlock:         gas.LimitsDefault.MaxGasPerBlock,
		MaxGasPerRequest:       gas.LimitsDefault.MaxGasPerRequest,
		MaxGasExternalViewCall: gas.LimitsDefault.MaxGasExternalViewCall,
	}), chainclient.PostRequestParams{
		GasBudget: iotaclient.DefaultGasBudget,
	})
	require.NoError(t, err)
	_, err = e.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), tx, true, 10*time.Second)
	require.NoError(t, err)

	// Create tx sender with some funds
	sender := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	e.DepositFunds(100*isc.Million, sender.(*cryptolib.KeyPair))

	createTx := func(l1GasBudget, l2GasBudget uint64) []byte {
		// Mind some TESTCOINs
		// TODO: Can we avoid doing this for each tx? If we place this code outside of the function,
		// we get an error regarding version of some object (presumably treasuryCap).
		coinPackageID, treasuryCap := iotaclienttest.DeployCoinPackage(
			t,
			e.Clu.L1Client().IotaClient(),
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
			e.Clu.L1Client().IotaClient(),
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
			iotajsonrpc.CoinValue(2*iotaclient.DefaultGasBudget),
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

		ptb = iscmoveclient.PTBOptionNoneIotaCoin(ptb)
		argInitCoin := ptb.LastCommandResultArg()

		// We start new chain only to get an some real object (in this case it's anchor),
		// which we can place into asset bag. It could be any other object, but using anchor is just very simple.
		assetObj := ptb.Command(
			iotago.Command{
				MoveCall: &iotago.ProgrammableMoveCall{
					Package:       lo.ToPtr(l1starter.ISCPackageID()),
					Module:        iscmove.AnchorModuleName,
					Function:      "start_new_chain",
					TypeArguments: []iotago.TypeTag{},
					Arguments: []iotago.Argument{
						ptb.MustPure([]byte(nil)),
						argInitCoin,
					},
				},
			},
		)

		// Place dummy object into the asset bag
		ptb = iscmoveclient.PTBAssetsBagPlaceObject(
			ptb,
			l1starter.ISCPackageID(),
			argAssetsBag,
			assetObj,
			lo.Must(iotago.ObjectTypeFromString(l1starter.ISCPackageID().String()+"::anchor::Anchor")),
		)

		allowanceVal := iotajsonrpc.CoinValue(1 * isc.Million)
		allowance := iscmove.NewAssets(allowanceVal)
		allowance.SetCoin(iotajsonrpc.MustCoinTypeFromString(testcoinType.String()), iotajsonrpc.CoinValue(10))

		// Deposit funds from L1 asset bag into L2 account
		ptb = iscmoveclient.PTBCreateAndSendRequest(
			ptb,
			l1starter.ISCPackageID(),
			e.Chain.ChainID.AsObjectID(),
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
		coinsForGas, err := e.Clu.L1Client().FindCoinsForGasPayment(context.Background(), sender.Address().AsIotaAddress(), pt, iotaclient.DefaultGasPrice, l1GasBudget)
		require.NoError(t, err)

		txData := iotago.NewProgrammable(
			sender.Address().AsIotaAddress(),
			pt,
			coinsForGas,
			l1GasBudget,
			iotaclient.DefaultGasPrice,
		)

		txBytes, err := bcs.Marshal(&txData)
		require.NoError(t, err)

		return txBytes
	}

	// Create transaction for estimation
	txBytesForEstimation := createTx(0, 0)

	// Estimate L1 and L2 gas budget for that transaction
	estimatedReceipt, _, err := e.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOnledger(context.Background()).Request(apiclient.EstimateGasRequestOnledger{
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

	// TODO: The naming is confusing: for L1 we use "GasBudget", but for L2 we use "GasBurned".
	l1GasBudget := lo.Must(strconv.ParseUint(estimatedReceipt.L1.GasBudget, 10, 64))
	l2GasBudget := lo.Must(strconv.ParseUint(estimatedReceipt.L2.GasBurned, 10, 64))

	executeTx := func(txBytes []byte) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
		execRes, err := e.Clu.L1Client().SignAndExecuteTransaction(context.Background(), &iotaclient.SignAndExecuteTransactionRequest{
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

	// Executing transaction with proper L1 and L2 gas budgets
	txBytes := createTx(l1GasBudget, l2GasBudget)
	res, err := executeTx(txBytes)
	require.NoError(t, err)
	require.Empty(t, res.Errors)
	require.Empty(t, res.Effects.Data.V1.Status.Error, res.Effects.Data.V1.Status.Status)

	// Checked that actual used gas was not greater than estimated fee or budget.
	estimatedGasFee := lo.Must(strconv.ParseUint(estimatedReceipt.L1.GasFeeCharged, 10, 64))
	require.LessOrEqual(t, estimatedGasFee, l1GasBudget)
	var totalL1GasUsed big.Int
	totalL1GasUsed.Add(&totalL1GasUsed, res.Effects.Data.V1.GasUsed.ComputationCost.Int)
	totalL1GasUsed.Add(&totalL1GasUsed, res.Effects.Data.V1.GasUsed.StorageCost.Int)
	totalL1GasUsed.Sub(&totalL1GasUsed, res.Effects.Data.V1.GasUsed.StorageRebate.Int)
	require.LessOrEqual(t, totalL1GasUsed.Uint64(), estimatedGasFee)
	require.LessOrEqual(t, totalL1GasUsed.Uint64(), l1GasBudget)

	// Checking that computation and storage fees exactly match the estimated values.
	// Actual rebate value could be bigger (but not smaller) than estimated, because
	// more than objects were destroyed than expected - e.g. when gas coin was consumed.
	estimatedComputationFee := lo.Must(strconv.ParseUint(estimatedReceipt.L1.ComputationFee, 10, 64))
	estimatedStorageFee := lo.Must(strconv.ParseUint(estimatedReceipt.L1.StorageFee, 10, 64))
	estimatedStorageRebate := lo.Must(strconv.ParseUint(estimatedReceipt.L1.StorageRebate, 10, 64))
	require.Equal(t, estimatedComputationFee, res.Effects.Data.V1.GasUsed.ComputationCost.Int.Uint64())
	require.Equal(t, estimatedStorageFee, res.Effects.Data.V1.GasUsed.StorageCost.Int.Uint64())
	require.LessOrEqual(t, estimatedStorageRebate, res.Effects.Data.V1.GasUsed.StorageRebate.Int.Uint64())

	recs, err := e.Clu.MultiClient().WaitUntilAllRequestsProcessed(context.Background(), res, false, 10*time.Second)
	require.NoError(t, err, recs)
	require.Empty(t, recs[0].ErrorMessage, lo.FromPtr(recs[0].ErrorMessage))
	require.Equal(t, recs[0].GasBurned, estimatedReceipt.L2.GasBurned)
	require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.L2.GasFeeCharged)

	// Checking that transaction execution fails with wrong L1 gas budget
	// NOTE: The reason we use l1GasBudget-1200000 here is because actual storage rebate is bigger than estimated
	//       due to mock gas coins being used for estimation. So we cannot use l1GasBudget-1 - it simply won't trigger an error.
	//       For L1 estimated budget is not strictly equal to actual needed value, it is greater or equal. So such test is hard to write.
	txBytesWithWrongL1GasBudget := createTx(l1GasBudget-1200000, l2GasBudget)
	res, _ = executeTx(txBytesWithWrongL1GasBudget)
	require.Equal(t, "InsufficientGas", res.Effects.Data.V1.Status.Error, res.Effects.Data.V1.Status.Status)

	// Checking that transaction execution fails with wrong L2 gas budget
	txBytesWithWrongL2GasBudget := createTx(l1GasBudget, l2GasBudget-1)
	res, err = executeTx(txBytesWithWrongL2GasBudget)
	require.NoError(t, err)
	require.Empty(t, res.Errors)
	require.Empty(t, res.Effects.Data.V1.Status.Error, res.Effects.Data.V1.Status.Status)
	recs, _ = e.Clu.MultiClient().WaitUntilAllRequestsProcessed(context.Background(), res, false, 10*time.Second)
	require.Equal(t, "gas budget exceeded", lo.FromPtr(recs[0].ErrorMessage))
}

func (e *ChainEnv) testEstimateGasOffLedger(t *testing.T) {
	// estimate off-ledger request, then send the same request, assert the gas used/fees match
	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	e.DepositFunds(10*isc.Million, keyPair)

	estimationReq := isc.NewOffLedgerRequest(
		e.Chain.ChainID,
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		0,
		1*isc.Million,
	).WithAllowance(isc.NewAssets(5000)).
		WithSender(keyPair.GetPublicKey())

	// Test that the API will fail if the FromAddress is missing
	estimatedReceiptFail, _, err := e.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOffledger(context.Background()).Request(apiclient.EstimateGasRequestOffledger{
		RequestBytes: cryptolib.EncodeHex(estimationReq.Bytes()),
	}).Execute()
	require.Error(t, err)
	require.Nil(t, estimatedReceiptFail)
	///

	requestHex := cryptolib.EncodeHex(estimationReq.Bytes())

	estimatedReceipt, _, err := e.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOffledger(context.Background()).Request(apiclient.EstimateGasRequestOffledger{
		FromAddress:  keyPair.Address().String(),
		RequestBytes: requestHex,
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	client := e.Chain.Client(keyPair)
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
	rec, err := e.Clu.MultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), req.ID(), false, 30*time.Second)
	require.NoError(t, err)
	require.Equal(t, rec.GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, rec.GasFeeCharged, estimatedReceipt.GasFeeCharged)
}
