// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"crypto/ecdsa"
	"math"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc/jsonrpctest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
)

type clusterTestEnv struct {
	jsonrpctest.Env
	ChainEnv
}

func newClusterTestEnv(t *testing.T, env *ChainEnv, nodeIndex int) *clusterTestEnv {
	evmtest.InitGoEthLogger(t)

	jsonRPCEndpoint := "http://" + env.Clu.Config.APIHost(nodeIndex) + routes.EVMJSONRPC(env.Chain.ChainID.String())
	rawClient, err := rpc.DialHTTP(jsonRPCEndpoint)
	require.NoError(t, err)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	waitTxConfirmed := func(txHash common.Hash) error {
		c := env.Chain.Client(nil, nodeIndex)
		reqID, err := c.RequestIDByEVMTransactionHash(txHash)
		if err != nil {
			return err
		}
		receipt, err := c.WaspClient.WaitUntilRequestProcessed(env.Chain.ChainID, reqID, 1*time.Minute)
		if err != nil {
			return err
		}
		if receipt.Error != nil {
			resolved, err := errors.Resolve(receipt.Error, func(contractName string, funcName string, params dict.Dict) (dict.Dict, error) {
				return c.CallView(isc.Hn(contractName), funcName, params)
			})
			if err != nil {
				return err
			}
			return resolved
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

func newEthereumAccount() (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey)
}

const transferAllowanceToGasBudgetBaseTokens = 1 * isc.Million

func (e *clusterTestEnv) newEthereumAccountWithL2Funds(baseTokens ...uint64) (*ecdsa.PrivateKey, common.Address) {
	ethKey, ethAddr := newEthereumAccount()
	walletKey, walletAddr, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.T, err)

	var amount uint64
	if len(baseTokens) > 0 {
		amount = baseTokens[0]
	} else {
		amount = e.Clu.L1BaseTokens(walletAddr) - transferAllowanceToGasBudgetBaseTokens
	}
	gasBudget := uint64(math.MaxUint64)
	tx, err := e.Chain.Client(walletKey).Post1Request(accounts.Contract.Hname(), accounts.FuncTransferAllowanceTo.Hname(), chainclient.PostRequestParams{
		Transfer: isc.NewAssets(amount+transferAllowanceToGasBudgetBaseTokens, nil),
		Args: map[kv.Key][]byte{
			accounts.ParamAgentID: codec.EncodeAgentID(isc.NewEthereumAddressAgentID(ethAddr)),
		},
		Allowance: isc.NewAssetsBaseTokens(amount),
		GasBudget: &gasBudget,
	})
	require.NoError(e.T, err)

	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.Chain.ChainID, tx, 30*time.Second)
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
	clu := newCluster(t, waspClusterOpts{nNodes: 5})
	chain, err := clu.DeployChainWithDKG("testchain", clu.Config.AllNodes(), []int{0, 1, 2, 3}, uint16(3))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)
	e := newClusterTestEnv(t, env, 4) // node #4 is an access node
	e.TestRPCGetLogs()
}
