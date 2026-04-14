package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

// executed in cluster_test.go
func (e *ChainEnv) testOffLedgerNonce(t *testing.T) {
	chClient := e.newWalletWithL2Funds(0, 0, 1, 2, 3)

	_, err := chClient.DepositFunds(1 * isc.Million)
	require.NoError(t, err)
	// send off-ledger request with a high nonce
	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(10),
			Nonce:     1_000_000,
		},
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, offledgerReq.ID(), false, 30*time.Second)
	require.Error(t, err) // wont' be processed

	// send off-ledger requests with the correct nonce
	for i := uint64(0); i < 2; i++ {
		req, err2 := chClient.PostOffLedgerRequest(context.Background(),
			accounts.FuncWithdraw.Message(),
			chainclient.PostRequestParams{
				Allowance: isc.NewAssets(10),
				Nonce:     i,
			},
		)
		require.NoError(t, err2)
		_, err2 = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, req.ID(), false, 10*time.Second)
		require.NoError(t, err2)
	}

	// try replaying an older nonce
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		accounts.FuncTransferAllowanceTo.Message(nil),
		chainclient.PostRequestParams{
			Nonce: 1,
		},
	)
	require.Error(t, err)
	require.Contains(t, string(err.(*apiclient.GenericOpenAPIError).Body()), "not added to the mempool")
}

func (e *ChainEnv) newWalletWithL2Funds(waspnode int, waitOnNodes ...int) *chainclient.Client {
	baseTokes := coin.Value(1000 * isc.Million)
	userWallet, userAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	userAgentID := isc.NewAddressAgentID(userAddress)

	e.NewChainClient(userWallet)
	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(waspnode), e.Chain.ChainID, userWallet)

	// deposit funds before sending the off-ledger requestargs
	reqTx, err := chClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(baseTokes),
		GasBudget: iotagraphql.DefaultGasBudget,
	})
	require.NoError(e.t, err)

	receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(e.t, err)

	gasFeeCharged, err := util.DecodeUint64(receipts[0].GasFeeCharged)
	require.NoError(e.t, err)

	expectedBaseTokens := baseTokes - coin.Value(gasFeeCharged)
	e.checkBalanceOnChain(userAgentID, coin.BaseTokenType, expectedBaseTokens)

	// wait until access node syncs with account
	if len(waitOnNodes) > 0 {
		waitUntil(e.t, e.accountExists(userAgentID), waitOnNodes, 30*time.Second)
	}
	return chClient
}
