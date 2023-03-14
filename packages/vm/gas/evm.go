package gas

import "github.com/iotaledger/wasp/packages/util"

// <ISC gas> = <EVM Gas> * <A> / <B>
var DefaultEVMGasRatio = util.Ratio32{A: 1, B: 1}

func ISCGasBudgetToEVM(iscGasBudget uint64, gasRatio *util.Ratio32) uint64 {
	// EVM gas budget = floor(ISC gas budget * B / A)
	return gasRatio.YFloor64(iscGasBudget)
}

func ISCGasBurnedToEVM(iscGasBurned uint64, gasRatio *util.Ratio32) uint64 {
	// estimated EVM gas = ceil(ISC gas burned * B / A)
	return gasRatio.YCeil64(iscGasBurned)
}

func EVMGasToISC(evmGas uint64, gasRatio *util.Ratio32) uint64 {
	// ISC gas burned = ceil(EVM gas * A / B)
	return gasRatio.XCeil64(evmGas)
}

// EVMBlockGasLimit returns the ISC block gas limit converted to EVM gas units
func EVMBlockGasLimit(gasLimits *Limits, gasRatio *util.Ratio32) uint64 {
	return ISCGasBudgetToEVM(gasLimits.MaxGasPerBlock, gasRatio)
}

// EVMCallGasLimit returns the maximum gas limit accepted for an EVM tx
func EVMCallGasLimit(gasLimits *Limits, gasRatio *util.Ratio32) uint64 {
	return ISCGasBudgetToEVM(gasLimits.MaxGasPerRequest, gasRatio)
}
