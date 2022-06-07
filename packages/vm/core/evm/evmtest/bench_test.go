// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func initBenchmark(b *testing.B) (*solo.Chain, []iscp.Request) {
	// setup: deploy the EVM chain
	log := testlogger.NewSilentLogger(b.Name(), true)
	s := solo.New(b, &solo.InitOptions{AutoAdjustDustDeposit: true, Log: log})
	env := initEVMWithSolo(b, s)
	// setup: deploy the `storage` EVM contract
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey, 42)

	gasLimit := uint64(100000)

	// setup: prepare N requests that call FuncSendTransaction with an EVM tx
	// that calls `storage.store()`
	reqs := make([]iscp.Request, b.N)
	for i := 0; i < b.N; i++ {
		ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
		tx := storage.buildEthTx([]ethCallOptions{{
			sender:   ethKey,
			gasLimit: gasLimit,
		}}, "store", uint32(i))
		var err error
		reqs[i], err = iscp.NewEVMOffLedgerRequest(env.soloChain.ChainID, tx)
		require.NoError(b, err)
	}

	return env.soloChain, reqs
}

// run benchmarks with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'

// BenchmarkEVMSingle sends a single Ethereum tx and waits until it is processed, N times
func BenchmarkEVMSingle(b *testing.B) {
	ch, reqs := initBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.RunOffLedgerRequest(reqs[i])
	}
}

// BenchmarkEVMMulti sends N Ethereum txs and waits for them to be processed
func BenchmarkEVMMulti(b *testing.B) {
	ch, reqs := initBenchmark(b)
	b.ResetTimer()
	ch.RunOffLedgerRequests(reqs)
}
