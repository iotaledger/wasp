//go:build gascalibration

package evmtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/gascalibration"
)

const factor = 10

func TestGasUsageMemoryContract(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	gasTest := env.deployGasTestMemoryContract(ethKey)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * factor
		res, err := gasTest.f(n)
		require.NoError(t, err)
		t.Logf("n = %d, gas used: %d", n, res.iscReceipt.GasBurned)
		results[n] = res.iscReceipt.GasBurned
	}
	gascalibration.SaveTestResultAsJSON("memory_sol.json", results)
}

func TestGasUsageStorageContract(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	gasTest := env.deployGasTestStorageContract(ethKey)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * factor
		res, err := gasTest.f(n)
		require.NoError(t, err)
		t.Logf("n = %d, gas used: %d", n, res.iscReceipt.GasBurned)
		results[n] = res.iscReceipt.GasBurned
	}
	gascalibration.SaveTestResultAsJSON("storage_sol.json", results)
}

func TestGasUsageExecutionTimeContract(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	gasTestContract := env.deployGasTestExecutionTimeContract(ethKey)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * factor
		res, err := gasTestContract.f(n)
		require.NoError(t, err)
		t.Logf("n = %d, gas used: %d", n, res.iscReceipt.GasBurned)
		results[n] = res.iscReceipt.GasBurned
	}
	gascalibration.SaveTestResultAsJSON("executiontime_sol.json", results)
}
