// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpctest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/evm/evmtest"
)

func TestSubscriptionNewHeads(t *testing.T) {
	env := newSoloTestEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := make(chan *types.Header, 10)

	sub, err := env.Client.SubscribeNewHead(ctx, ch)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	require.EqualValues(t, 0, env.BlockNumber())

	// this will create a new header, which should be enqueued in ch
	_, _ = env.soloChain.NewEthereumAccountWithL2Funds()

	for {
		select {
		case header := <-ch:
			require.EqualValues(t, 1, header.Number.Uint64())
			return

		case err := <-sub.Err():
			require.NoError(t, err)

		case <-ctx.Done():
			require.FailNow(t, "timeout")
		}
	}
}

func TestSubscriptionLogs(t *testing.T) {
	env := newSoloTestEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	creator, creatorAddress := env.NewAccountWithL2Funds()
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(env.T, err)
	contractAddress := crypto.CreateAddress(creatorAddress, env.NonceAt(creatorAddress))

	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	ch := make(chan types.Log, 10)
	sub, err := env.Client.SubscribeFilterLogs(ctx, filterQuery, ch)
	require.NoError(t, err)
	defer sub.Unsubscribe()

	_, receipt, _ := env.DeployEVMContract(creator, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")
	require.Equal(env.T, 1, len(receipt.Logs))

	// assert that we received 1 log in the subscription
	for {
		select {
		case <-ch:
			return

		case err := <-sub.Err():
			require.NoError(t, err)

		case <-ctx.Done():
			require.FailNow(t, "timeout")
		}
	}
}
