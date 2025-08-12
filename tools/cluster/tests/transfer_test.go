package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

func TestDepositWithdraw(t *testing.T) {
	e := SetupWithChain(t)

	userKeypair, userAddr, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)

	require.True(t, e.Clu.AssertAddressBalances(userAddr, isc.NewAssets(iotaclient.FundsFromFaucetAmount)))

	myAgentID := isc.NewAddressAgentID(userAddr)
	e.checkBalanceOnChain(myAgentID, coin.BaseTokenType, 0)

	// deposit some base tokens to the chain
	var depositBaseTokens coin.Value = 10 * isc.Million
	chClient := e.NewChainClient(userKeypair)

	params := chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(depositBaseTokens),
		GasBudget: iotaclient.DefaultGasBudget,
	}
	reqTx, err := chClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), params)
	require.NoError(t, err)

	receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), reqTx, true, 30*time.Second)
	require.NoError(t, err)

	gasFees1, err := util.DecodeUint64(receipts[0].GasFeeCharged)
	require.NoError(t, err)

	var onChainBalance coin.Value = depositBaseTokens - coin.Value(gasFees1)
	e.checkBalanceOnChain(myAgentID, coin.BaseTokenType, onChainBalance)
	require.True(t,
		e.Clu.AssertAddressBalances(userAddr, isc.NewAssets(iotaclient.FundsFromFaucetAmount-depositBaseTokens-coin.Value(reqTx.Effects.Data.GasFee()))),
	)

	// withdraw some base tokens back
	var baseTokensToWithdraw coin.Value = 1 * isc.Million
	req, err := chClient.PostOffLedgerRequest(context.Background(), accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(baseTokensToWithdraw),
		},
	)
	require.NoError(t, err)
	receipt, err := e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), req.ID(), true, 30*time.Second)
	require.NoError(t, err)

	gasFees2, err := util.DecodeUint64(receipt.GasFeeCharged)
	require.NoError(t, err)

	e.checkBalanceOnChain(myAgentID, coin.BaseTokenType, onChainBalance-baseTokensToWithdraw-coin.Value(gasFees2))
	require.True(t,
		e.Clu.AssertAddressBalances(userAddr, isc.NewAssets(iotaclient.FundsFromFaucetAmount-depositBaseTokens+baseTokensToWithdraw-coin.Value(reqTx.Effects.Data.GasFee()))),
	)

	// TODO use "withdraw all base tokens" entrypoint to withdraw all remaining base tokens
}
