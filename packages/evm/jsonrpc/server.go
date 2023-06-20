// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/packages/metrics"
)

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
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			return nil, err
		}
	}
	return rpcsrv, nil
}
