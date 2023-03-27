// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func initBenchmark(b *testing.B) (*solo.Chain, []isc.Request) {
	// setup: deploy the EVM chain
	log := testlogger.NewSilentLogger(b.Name(), true)
	s := solo.New(b, &solo.InitOptions{AutoAdjustStorageDeposit: true, Log: log})
	env := initEVMWithSolo(b, s)
	// setup: deploy the `storage` EVM contract
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	storage := env.deployStorageContract(ethKey)

	gasLimit := uint64(100000)

	// setup: prepare N requests that call FuncSendTransaction with an EVM tx
	// that calls `storage.store()`
	reqs := make([]isc.Request, b.N)
	for i := 0; i < b.N; i++ {
		ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
		tx, err := storage.buildEthTx([]ethCallOptions{{
			sender:   ethKey,
			gasLimit: gasLimit,
		}}, "store", uint32(i))
		require.NoError(b, err)
		reqs[i], err = isc.NewEVMOffLedgerTxRequest(env.soloChain.ChainID, tx)
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
