// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
)

func NewServer(evmChain *EVMChain, accountManager *AccountManager) (*rpc.Server, error) {
	chainID, err := evmChain.ChainID()
	if err != nil {
		return nil, err
	}
	rpcsrv := rpc.NewServer()
	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"web3", NewWeb3Service()},
		{"net", NewNetService(int(chainID))},
		{"eth", NewEthService(evmChain, accountManager)},
		{"txpool", NewTxPoolService()},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			return nil, err
		}
	}
	return rpcsrv, nil
}
