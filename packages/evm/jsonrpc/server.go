// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/packages/metrics"
)

type Parameters struct {
	Logs                                LogsLimits
	WebsocketRateLimitMessagesPerMinute int
}

func NewParameters(
	maxBlocksInLogsFilterRange int,
	maxLogsInResult int,
	websocketRateLimitMessagesPerMinute int,
) *Parameters {
	return &Parameters{
		Logs: LogsLimits{
			MaxBlocksInLogsFilterRange: maxBlocksInLogsFilterRange,
			MaxLogsInResult:            maxLogsInResult,
		},
		WebsocketRateLimitMessagesPerMinute: websocketRateLimitMessagesPerMinute,
	}
}

func ParametersDefault() *Parameters {
	return &Parameters{
		Logs: LogsLimits{
			MaxBlocksInLogsFilterRange: 1000,
			MaxLogsInResult:            10000,
		},
		WebsocketRateLimitMessagesPerMinute: 1000,
	}
}

func NewServer(
	evmChain *EVMChain,
	accountManager *AccountManager,
	metrics *metrics.ChainWebAPIMetrics,
	params *Parameters,
) (*rpc.Server, error) {
	chainID := evmChain.ChainID()
	rpcsrv := rpc.NewServer()

	for _, srv := range []struct {
		namespace string
		service   interface{}
	}{
		{"web3", NewWeb3Service()},
		{"net", NewNetService(int(chainID))},
		{"eth", NewEthService(evmChain, accountManager, metrics, params)},
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
