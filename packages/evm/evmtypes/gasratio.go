package evmtypes

import "github.com/iotaledger/wasp/packages/util"

var (
	// <ISC gas> = <EVM Gas> * <A> / <B>
	DefaultGasRatio = util.Ratio32{A: 1, B: 1}
)

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
