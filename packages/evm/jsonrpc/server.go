// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"time"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/packages/metrics"
)

type Parameters struct {
	Logs                                LogsLimits
	WebsocketRateLimitMessagesPerSecond int
	WebsocketRateLimitBurst             int
	WebsocketRateLimitEnabled           bool
	WebsocketConnectionCleanupDuration  time.Duration
	WebsocketClientBlockDuration        time.Duration
}

func NewParameters(
	maxBlocksInLogsFilterRange int,
	maxLogsInResult int,
	websocketRateLimitMessagesPerSecond int,
	websocketRateLimitBurst int,
	websocketConnectionCleanupDuration time.Duration,
	websocketClientBlockDuration time.Duration,
	WebsocketRateLimitEnabled bool,
) *Parameters {
	return &Parameters{
		Logs: LogsLimits{
			MaxBlocksInLogsFilterRange: maxBlocksInLogsFilterRange,
			MaxLogsInResult:            maxLogsInResult,
		},
		WebsocketRateLimitMessagesPerSecond: websocketRateLimitMessagesPerSecond,
		WebsocketRateLimitBurst:             websocketRateLimitBurst,
		WebsocketRateLimitEnabled:           WebsocketRateLimitEnabled,
		WebsocketConnectionCleanupDuration:  websocketConnectionCleanupDuration,
		WebsocketClientBlockDuration:        websocketClientBlockDuration,
	}
}

func ParametersDefault() *Parameters {
	return &Parameters{
		Logs: LogsLimits{
			MaxBlocksInLogsFilterRange: 1000,
			MaxLogsInResult:            10000,
		},
		WebsocketRateLimitMessagesPerSecond: 20,
		WebsocketRateLimitBurst:             5,
		WebsocketRateLimitEnabled:           true,
		WebsocketConnectionCleanupDuration:  5 * time.Minute,
		WebsocketClientBlockDuration:        5 * time.Minute,
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
		{"trace", NewTraceService(evmChain, metrics)},
	} {
		err := rpcsrv.RegisterName(srv.namespace, srv.service)
		if err != nil {
			return nil, err
		}
	}
	return rpcsrv, nil
}
