// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc/jsonrpctest"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type clusterTestEnv struct {
	jsonrpctest.Env
	cluster *cluster.Cluster
	chain   *cluster.Chain
}

func newClusterTestEnv(t *testing.T) *clusterTestEnv {
	// TODO these tests need to be re-written, since the RPC server will be an integral part of wasp, rather than spawned by wasp-cli
	t.Skip()

	evmtest.InitGoEthLogger(t)

	clu := newCluster(t)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	chainID := evm.DefaultChainID

	signer, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	backend := jsonrpc.NewWaspClientBackend(chain.Client(signer))
	evmChain := jsonrpc.NewEVMChain(backend, chainID)

	accountManager := jsonrpc.NewAccountManager(evmtest.Accounts)

	rpcsrv := jsonrpc.NewServer(evmChain, accountManager)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &clusterTestEnv{
		Env: jsonrpctest.Env{
			T:         t,
			Server:    rpcsrv,
			Client:    client,
			RawClient: rawClient,
			ChainID:   chainID,
		},
		cluster: clu,
		chain:   chain,
	}
}

func TestEVMJsonRPCClusterGetLogs(t *testing.T) {
	newClusterTestEnv(t).TestRPCGetLogs()
}

func TestEVMJsonRPCClusterGasLimit(t *testing.T) {
	newClusterTestEnv(t).TestRPCGasLimit()
}

func TestEVMJsonRPCClusterInvalidNonce(t *testing.T) {
	newClusterTestEnv(t).TestRPCInvalidNonce()
}
