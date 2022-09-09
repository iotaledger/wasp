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
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type clusterTestEnv struct {
	jsonrpctest.Env
	cluster *cluster.Cluster
	chain   *cluster.Chain
}

func newClusterTestEnv(t *testing.T, committee []int, nodeIndex int, opt ...waspClusterOpts) *clusterTestEnv {
	evmtest.InitGoEthLogger(t)

	clu := newCluster(t, opt...)

	quorum := uint16((2*len(committee))/3 + 1)
	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID)

	jsonRPCEndpoint := "http://" + clu.Config.APIHost(nodeIndex) + routes.EVMJSONRPC(chain.ChainID.String())
	rawClient, err := rpc.DialHTTP(jsonRPCEndpoint)
	require.NoError(t, err)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	waitTxConfirmed := func(txHash common.Hash) error {
		c := chain.Client(nil, nodeIndex)
		reqID, err := c.RequestIDByEVMTransactionHash(txHash)
		if err != nil {
			return err
		}
		receipt, err := c.WaspClient.WaitUntilRequestProcessed(chain.ChainID, reqID, 1*time.Minute)
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

	return &clusterTestEnv{
		Env: jsonrpctest.Env{
			T:               t,
			Client:          client,
			RawClient:       rawClient,
			ChainID:         evm.DefaultChainID,
			WaitTxConfirmed: waitTxConfirmed,
		},
		cluster: clu,
		chain:   chain,
	}
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
	walletKey, walletAddr, err := e.cluster.NewKeyPairWithFunds()
	require.NoError(e.T, err)

	var amount uint64
	if len(baseTokens) > 0 {
		amount = baseTokens[0]
	} else {
		amount = e.cluster.L1BaseTokens(walletAddr) - transferAllowanceToGasBudgetBaseTokens
	}
	gasBudget := uint64(math.MaxUint64)
	tx, err := e.chain.Client(walletKey).Post1Request(accounts.Contract.Hname(), accounts.FuncTransferAllowanceTo.Hname(), chainclient.PostRequestParams{
		Transfer: isc.NewFungibleTokens(amount+transferAllowanceToGasBudgetBaseTokens, nil),
		Args: map[kv.Key][]byte{
			accounts.ParamAgentID:          codec.EncodeAgentID(isc.NewEthereumAddressAgentID(ethAddr)),
			accounts.ParamForceOpenAccount: codec.EncodeBool(true),
		},
		Allowance: isc.NewAllowanceBaseTokens(amount),
		GasBudget: &gasBudget,
	})
	require.NoError(e.T, err)

	_, err = e.chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(e.chain.ChainID, tx, 30*time.Second)
	require.NoError(e.T, err)

	return ethKey, ethAddr
}

func TestEVMJsonRPCClusterGetLogs(t *testing.T) {
	e := newClusterTestEnv(t, []int{0, 1, 2, 3}, 0, waspClusterOpts{
		nNodes: 4,
	})
	e.TestRPCGetLogs(e.newEthereumAccountWithL2Funds)
}

func TestEVMJsonRPCClusterInvalidTx(t *testing.T) {
	e := newClusterTestEnv(t, []int{0, 1, 2, 3}, 0, waspClusterOpts{
		nNodes: 4,
	})
	e.TestRPCInvalidNonce(e.newEthereumAccountWithL2Funds)
	e.TestRPCGasLimitTooLow(e.newEthereumAccountWithL2Funds)
}

func TestEVMJsonRPCClusterAccessNode(t *testing.T) {
	e := newClusterTestEnv(t, []int{0, 1, 2, 3}, 4, waspClusterOpts{
		nNodes: 5, // node #4 is an access node
	})
	e.TestRPCGetLogs(e.newEthereumAccountWithL2Funds)
}
