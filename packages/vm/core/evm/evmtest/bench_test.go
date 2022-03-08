// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/solo/solobench"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/stretchr/testify/require"
)

func initBenchmark(b *testing.B) (*solo.Chain, []*solo.CallParams) {
	// setup: deploy the EVM chain
	log := testlogger.NewSilentLogger(b.Name(), true)
	env := solo.New(b, &solo.InitOptions{Log: log})
	evmChain := initEVMWithSolo(b, env)
	// setup: deploy the `storage` EVM contract
	storage := evmChain.deployStorageContract(evmChain.faucetKey, 42)

	// setup: prepare N requests that call FuncSendTransaction with an EVM tx
	// that calls `storage.store()`
	reqs := make([]*solo.CallParams, b.N)
	for i := 0; i < b.N; i++ {
		sender, err := crypto.GenerateKey() // send from a new address so that nonce is always 0
		require.NoError(b, err)

		txdata, _, _ := storage.buildEthTxData([]ethCallOptions{{sender: sender}}, "store", uint32(i))
		reqs[i] = storage.chain.buildSoloRequest(evm.FuncSendTransaction.Name, evm.FieldTransactionData, txdata)
		reqs[i].WithMaxAffordableGasBudget()
	}

	return evmChain.soloChain, reqs
}

// run benchmarks with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'

func doBenchmark(b *testing.B, f solobench.Func) {
	ch, reqs := initBenchmark(b)
	f(b, ch, reqs, nil)
}

func BenchmarkEVMSync(b *testing.B) {
	doBenchmark(b, solobench.RunBenchmarkSync)
}

func BenchmarkEVMAsync(b *testing.B) {
	doBenchmark(b, solobench.RunBenchmarkAsync)
}
