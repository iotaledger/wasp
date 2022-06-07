// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
)

func NewServer(evmChain *EVMChain, accountManager *AccountManager) *rpc.Server {
	rpcsrv := rpc.NewServer()
	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"web3", NewWeb3Service()},
		{"net", NewNetService(int(evmChain.chainID))},
		{"eth", NewEthService(evmChain, accountManager)},
		{"txpool", NewTxPoolService()},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			panic(err)
		}
	}
	return rpcsrv
}
