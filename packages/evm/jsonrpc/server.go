// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/packages/metrics"
)

type ParametersJSONRPC struct {
	MaxBlocksInLogsFilterRange int `default:"1000" usage:"maximum amount of blocks in eth_getLogs filter range"`
	MaxLogsInResult            int `default:"10000" usage:"maximum amount of logs in eth_getLogs result"`
}

var Params = &ParametersJSONRPC{
	MaxBlocksInLogsFilterRange: 1000,
	MaxLogsInResult:            10000,
}

func NewServer(
	evmChain *EVMChain,
	accountManager *AccountManager,
	metrics *metrics.ChainWebAPIMetrics,
) (*rpc.Server, error) {
	chainID := evmChain.ChainID()
	rpcsrv := rpc.NewServer()
	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"web3", NewWeb3Service()},
		{"net", NewNetService(int(chainID))},
		{"eth", NewEthService(evmChain, accountManager, metrics)},
		{"debug", NewDebugService(evmChain, metrics)},
		{"txpool", NewTxPoolService()},
		{"evm", NewEVMService(evmChain)},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			return nil, err
		}
	}
	return rpcsrv, nil
}
