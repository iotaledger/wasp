// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc/jsonrpctest"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	evmtest.InitGoEthLogger(t)

	clu := newCluster(t)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	chainID := evm.DefaultChainID

	_, err = chain.DeployContract(
		evm.Contract.Name,
		evm.Contract.ProgramHash.String(),
		"EVM chain on top of ISC",
		map[string]interface{}{
			evm.FieldChainID: codec.EncodeUint16(uint16(chainID)),
			evm.FieldGenesisAlloc: evmtypes.EncodeGenesisAlloc(core.GenesisAlloc{
				evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
			}),
		},
	)
	require.NoError(t, err)

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
