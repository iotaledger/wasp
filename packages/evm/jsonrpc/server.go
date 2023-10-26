// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/time/rate"

	"github.com/iotaledger/wasp/packages/metrics"
)

type Parameters struct {
	Logs                                LogsLimits
	WebsocketRateLimitMessagesPerSecond rate.Limit
	WebsocketRateLimitBurst             int
}

func NewParameters(
	maxBlocksInLogsFilterRange int,
	maxLogsInResult int,
	websocketRateLimitMessagesPerSecond rate.Limit,
	websocketRateLimitBurst int,

) *Parameters {
	return &Parameters{
		Logs: LogsLimits{
			MaxBlocksInLogsFilterRange: maxBlocksInLogsFilterRange,
			MaxLogsInResult:            maxLogsInResult,
		},
		WebsocketRateLimitMessagesPerSecond: websocketRateLimitMessagesPerSecond,
		WebsocketRateLimitBurst:             websocketRateLimitBurst,
	}
}

func ParametersDefault() *Parameters {
	return &Parameters{
		Logs: LogsLimits{
			MaxBlocksInLogsFilterRange: 1000,
			MaxLogsInResult:            10000,
		},
		WebsocketRateLimitMessagesPerSecond: 100,
		WebsocketRateLimitBurst:             1,
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
