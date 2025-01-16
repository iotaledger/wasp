package tests

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

func TestOffledgerRequestAccessNode(t *testing.T) {
	const clusterSize = 10
	clu := newCluster(t, waspClusterOpts{nNodes: clusterSize})

	cmt := []int{0, 1, 2, 3}

	addr, err := clu.RunDKG(cmt, 3)
	require.NoError(t, err)

	chain, err := clu.DeployChain(clu.Config.AllNodes(), cmt, 3, addr)
	require.NoError(t, err)

	e := newChainEnv(t, clu, chain)

	// use an access node to create the chainClient
	chClient := newWalletWithFunds(e, 5, 0, 2, 4, 5, 7)

	// send off-ledger request via Web API (to the access node)
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		inccounter.FuncIncCounter.Message(nil),
	)
	require.NoError(t, err)

	waitUntil(t, e.counterEquals(43), clu.Config.AllNodes(), 30*time.Second)

	// check off-ledger request was successfully processed (check by asking another access node)
	ret, err := apiextensions.CallView(
		context.Background(),
		clu.WaspClient(6),
		e.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: inccounter.Contract.Hname().String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		})
	require.NoError(t, err)
	require.EqualValues(t, 43, lo.Must(inccounter.ViewGetCounter.DecodeOutput(ret)))
}

// executed in cluster_test.go
func testOffledgerRequest(t *testing.T, e *ChainEnv) {
	chClient := newWalletWithFunds(e, 0, 0, 1, 2, 3)

	// send off-ledger request via Web API
	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		inccounter.FuncIncCounter.Message(nil),
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, offledgerReq.ID(), false, 30*time.Second)
	require.NoError(t, err)

	ret, err := apiextensions.CallView(
		context.Background(),
		e.Chain.Cluster.WaspClient(0),
		e.Chain.ChainID.String(),
		apiclient.ContractCallViewRequest{
			ContractHName: inccounter.Contract.Hname().String(),
			FunctionHName: inccounter.ViewGetCounter.Hname().String(),
		})
	require.NoError(t, err)
	require.EqualValues(t, 43, lo.Must(inccounter.ViewGetCounter.DecodeOutput(ret)))
}

// executed in cluster_test.go
func testOffledgerNonce(t *testing.T, e *ChainEnv) {
	chClient := newWalletWithFunds(e, 0, 0, 1, 2, 3)

	// send off-ledger request with a high nonce
	offledgerReq, err := chClient.PostOffLedgerRequest(context.Background(),
		inccounter.FuncIncCounter.Message(nil),
		chainclient.PostRequestParams{
			Nonce: 1_000_000,
		},
	)
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, offledgerReq.ID(), false, 5*time.Second)
	require.Error(t, err) // wont' be processed

	// send off-ledger requests with the correct nonce
	for i := uint64(0); i < 5; i++ {
		req, err2 := chClient.PostOffLedgerRequest(context.Background(),
			inccounter.FuncIncCounter.Message(nil),
			chainclient.PostRequestParams{
				Nonce: i,
			},
		)
		require.NoError(t, err2)
		_, err2 = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, req.ID(), false, 10*time.Second)
		require.NoError(t, err2)
	}

	// try replaying an older nonce
	_, err = chClient.PostOffLedgerRequest(context.Background(),
		accounts.FuncTransferAccountToChain.Message(nil),
		chainclient.PostRequestParams{
			Nonce: 1,
		},
	)
	require.Error(t, err)
	require.Contains(t, string(err.(*apiclient.GenericOpenAPIError).Body()), "not added to the mempool")
}

func newWalletWithFunds(e *ChainEnv, waspnode int, waitOnNodes ...int) *chainclient.Client {
	baseTokes := coin.Value(1000 * isc.Million)
	userWallet, userAddress, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)
	userAgentID := isc.NewAddressAgentID(userAddress)

	chClient := chainclient.New(e.Clu.L1Client(), e.Clu.WaspClient(waspnode), e.Chain.ChainID, e.Clu.Config.ISCPackageID(), userWallet)

	// deposit funds before sending the off-ledger requestargs
	reqTx, err := chClient.PostRequest(context.Background(), accounts.FuncDeposit.Message(), chainclient.PostRequestParams{
		Transfer: isc.NewAssets(baseTokes),
	})
	require.NoError(e.t, err)

	receipts, err := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(e.t, err)

	gasFeeCharged, err := iotago.DecodeUint64(receipts[0].GasFeeCharged)
	require.NoError(e.t, err)

	expectedBaseTokens := baseTokes - coin.Value(gasFeeCharged)
	e.checkBalanceOnChain(userAgentID, coin.BaseTokenType, expectedBaseTokens)

	// wait until access node syncs with account
	if len(waitOnNodes) > 0 {
		waitUntil(e.t, e.accountExists(userAgentID), waitOnNodes, 10*time.Second)
	}
	return chClient
}
