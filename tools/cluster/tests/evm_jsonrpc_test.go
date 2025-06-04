// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc/jsonrpctest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type clusterTestEnv struct {
	jsonrpctest.Env
	ChainEnv
}

func newClusterTestEnv(t *testing.T, env *ChainEnv, nodeIndex int) *clusterTestEnv {
	evmJSONRPCPath := "/v1/chain/evm"
	jsonRPCEndpoint := env.Clu.Config.APIHost(nodeIndex) + evmJSONRPCPath
	rawClient, err := rpc.DialHTTP(jsonRPCEndpoint)
	require.NoError(t, err)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	waitTxConfirmed := func(txHash common.Hash) error {
		c := env.Chain.Client(nil, nodeIndex)
		reqID := isc.RequestIDFromEVMTxHash(txHash)
		receipt, _, err := c.WaspClient.ChainsAPI.
			WaitForRequest(context.Background(), reqID.String()).
			TimeoutSeconds(10).
			Execute()
		if err != nil {
			return err
		}

		if receipt.ErrorMessage != nil {
			return errors.New(*receipt.ErrorMessage)
		}

		return nil
	}

	e := &clusterTestEnv{
		Env: jsonrpctest.Env{
			T:               t,
			Client:          client,
			RawClient:       rawClient,
			ChainID:         evm.DefaultChainID,
			WaitTxConfirmed: waitTxConfirmed,
		},
		ChainEnv: *env,
	}
	e.Env.NewAccountWithL2Funds = e.newEthereumAccountWithL2Funds
	return e
}

const transferAllowanceToGasBudgetBaseTokens = 1 * isc.Million

func (e *clusterTestEnv) newEthereumAccountWithL2Funds(baseTokens ...coin.Value) (*ecdsa.PrivateKey, common.Address) {
	ethKey, ethAddr := newEthereumAccount()
	walletKey, walletAddr, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.T, err)

	var amount coin.Value
	if len(baseTokens) > 0 {
		amount = baseTokens[0]
	} else {
		amount = e.Clu.L1BaseTokens(walletAddr) - transferAllowanceToGasBudgetBaseTokens - iotaclient.DefaultGasBudget
	}
	tx, err := e.Chain.Client(walletKey).PostRequest(
		context.Background(),
		accounts.FuncTransferAllowanceTo.Message(isc.NewEthereumAddressAgentID(ethAddr)),
		chainclient.PostRequestParams{
			Transfer:  isc.NewAssets(amount + transferAllowanceToGasBudgetBaseTokens),
			Allowance: isc.NewAssets(amount),
			GasBudget: iotaclient.DefaultGasBudget,
		},
	)
	require.NoError(e.T, err)

	// We have to wait not only for the committee to process the request, but also for access nodes to get that info.
	_, err = e.Chain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, tx, false, 30*time.Second)
	require.NoError(e.T, err)

	return ethKey, ethAddr
}

// executed in cluster_test.go
func testEVMJsonRPCCluster(t *testing.T, env *ChainEnv) {
	e := newClusterTestEnv(t, env, 0)
	e.TestRPCGetLogs()
	e.TestRPCInvalidNonce()
	e.TestRPCGasLimitTooLow()
	e.TestRPCAccessHistoricalState()
	e.TestGasPrice()
}

func TestEVMJsonRPCClusterAccessNode(t *testing.T) {
	t.Skip("Cluster tests currently disabled")

	clu := newCluster(t, waspClusterOpts{nNodes: 5})
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), []int{0, 1, 2, 3}, uint16(3))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)
	e := newClusterTestEnv(t, env, 4) // node #4 is an access node
	e.TestRPCGetLogs()
}

func TestEVMJsonRPCZeroGasFee(t *testing.T) {
	t.Skip("Cluster tests currently disabled")

	clu := newCluster(t, waspClusterOpts{nNodes: 5})
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), []int{0, 1, 2, 3}, uint16(3))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)
	e := newClusterTestEnv(t, env, 4) // node #4 is an access node

	fp1 := gas.DefaultFeePolicy()
	fp1.GasPerToken = util.Ratio32{
		A: 0,
		B: 0,
	}
	govClient := e.Chain.Client(e.Chain.OriginatorKeyPair)
	reqTx, err := govClient.PostRequest(context.Background(), governance.FuncSetFeePolicy.Message(fp1), chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(iotaclient.DefaultGasBudget + 10),
		GasBudget: iotaclient.DefaultGasBudget,
	})
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(t, err)

	d, err := govClient.CallView(context.Background(), governance.ViewGetFeePolicy.Message())
	require.NoError(t, err)
	fp2, err := governance.ViewGetFeePolicy.DecodeOutput(d)
	require.NoError(t, err)
	require.Equal(t, fp1, fp2)
	e.TestRPCGetLogs()
}
