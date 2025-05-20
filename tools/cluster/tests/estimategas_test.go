package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func testEstimateGasOnLedger(t *testing.T, env *ChainEnv) {
	// estimate on-ledger request, then send the same request, assert the gas used/fees match
	t.Skip("TODO: Fix me")
	/*panic("refactor me: transaction.BasicOutputFromPostData")
	var output iotago.Output

	outputBytes, err := output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	require.NoError(t, err)

	estimatedReceipt, _, err := env.Chain.Cluster.WaspClient(0).ChainsAPI.EstimateGasOnledger(context.Background(),
		env.Chain.ChainID.String(),
	).Request(apiclient.EstimateGasRequestOnledger{
		OutputBytes: cryptolib.EncodeHex(outputBytes),
	}).Execute()
	require.NoError(t, err)
	require.Empty(t, estimatedReceipt.ErrorMessage)

	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	feeCharged, err := strconv.ParseUint(estimatedReceipt.GasFeeCharged, 10, 64)
	require.NoError(t, err)

	client := env.Chain.Client(keyPair)
	par := chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(coin.Value(feeCharged)),
		Allowance: isc.NewAssets(5000),
	}
	gasBudget, err := strconv.ParseUint(estimatedReceipt.GasBurned, 10, 64)
	require.NoError(t, err)
	par.WithGasBudget(gasBudget)

	tx, err := client.PostRequest(
		context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())),
		par,
	)
	require.NoError(t, err)
	recs, err := env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), env.Chain.ChainID, tx, false, 10*time.Second)
	require.NoError(t, err)
	require.Equal(t, recs[0].GasBurned, estimatedReceipt.GasBurned)
	require.Equal(t, recs[0].GasFeeCharged, estimatedReceipt.GasFeeCharged)*/
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
