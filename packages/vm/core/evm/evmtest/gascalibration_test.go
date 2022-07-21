package evmtest

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/gascalibration"
	"github.com/stretchr/testify/require"
)

const factor = 10

func TestGasUsageMemoryContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	gasTest := env.deployGasTestMemoryContract(ethKey)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * factor
		res, err := gasTest.f(n)
		require.NoError(t, err)
		t.Logf("n = %d, gas used: %d", n, res.iscpReceipt.GasBurned)
		results[n] = res.iscpReceipt.GasBurned
	}
	gascalibration.SaveTestResultAsJSON("memory_sol.json", results)
}

func TestGasUsageStorageContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	gasTest := env.deployGasTestStorageContract(ethKey)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * factor
		res, err := gasTest.f(n)
		require.NoError(t, err)
		t.Logf("n = %d, gas used: %d", n, res.iscpReceipt.GasBurned)
		results[n] = res.iscpReceipt.GasBurned
	}
	gascalibration.SaveTestResultAsJSON("storage_sol.json", results)
}

func TestGasUsageExecutionTimeContract(t *testing.T) {
	env := initEVM(t)
	ethKey, _ := env.soloChain.NewEthereumAccountWithL2Funds()
	gasTestContract := env.deployGasTestExecutionTimeContract(ethKey)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * factor
		res, err := gasTestContract.f(n)
		require.NoError(t, err)
		t.Logf("n = %d, gas used: %d", n, res.iscpReceipt.GasBurned)
		results[n] = res.iscpReceipt.GasBurned
	}
	gascalibration.SaveTestResultAsJSON("executiontime_sol.json", results)
}
