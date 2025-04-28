package gas

import "fmt"

const (
	BurnCodeStorage1P = BurnCode(iota)
	BurnCodeReadFromState1P
	BurnCodeCallTargetNotFound
	BurnCodeGetContext
	BurnCodeGetCallerData
	BurnCodeGetAllowance
	BurnCodeGetBalance
	BurnCodeGetCoinInfo
	BurnCodeCallContract
	BurnCodeEmitEvent1P
	BurnCodeTransferAllowance
	BurnCodeSendL1Request

	// Sandbox util codes

	BurnCodeUtilsHashingBlake2b
	BurnCodeUtilsHashingSha3
	BurnCodeUtilsHashingHname
	BurnCodeUtilsHexEncode
	BurnCodeUtilsHexDecode
	BurnCodeUtilsED25519AddrFromPubKey
	BurnCodeUtilsBLSValidSignature
	BurnCodeUtilsBLSAggregateBLS1P

	BurnCodeMinimumGasPerRequest1P

	BurnCodeEVM1P
)

// burnTable contains all possible burn codes with their burn value computing functions
var burnTable = BurnTable{
	BurnCodeCallTargetNotFound:         {"target n/f", constValue(10)},
	BurnCodeGetContext:                 {"context", constValue(10)},
	BurnCodeGetCallerData:              {"caller", constValue(10)},
	BurnCodeGetBalance:                 {"balance", constValue(20)},
	BurnCodeGetCoinInfo:                {"coin info", constValue(10)},
	BurnCodeCallContract:               {"call", constValue(100)},
	BurnCodeEmitEvent1P:                {"event", linear(1)}, // 1 gas per byte
	BurnCodeGetAllowance:               {"allowance", constValue(10)},
	BurnCodeTransferAllowance:          {"transfer", constValue(10)},
	BurnCodeSendL1Request:              {"send", linear(Coef1Send)},
	BurnCodeStorage1P:                  {"storage", linear(55)}, // 55 gas per byte
	BurnCodeReadFromState1P:            {"state read", linear(1)},
	BurnCodeUtilsHashingBlake2b:        {"blake2b", constValue(50)},
	BurnCodeUtilsHashingSha3:           {"sha3", constValue(80)},
	BurnCodeUtilsHashingHname:          {"hname", constValue(50)},
	BurnCodeUtilsHexEncode:             {"hex encode", linear(50)},
	BurnCodeUtilsHexDecode:             {"hex decode", linear(5)},
	BurnCodeUtilsED25519AddrFromPubKey: {"ed25517 addr", constValue(50)},
	BurnCodeUtilsBLSValidSignature:     {"bls valid", constValue(2000)},
	BurnCodeUtilsBLSAggregateBLS1P:     {"bls aggregate", linear(CoefBLSAggregate)},
	BurnCodeMinimumGasPerRequest1P:     {"minimum gas per request", minBurn(10000)},
	BurnCodeEVM1P:                      {"evm", linear(1)},
}

const (
	Coef1Send        = 200
	CoefBLSAggregate = 400
)

func constValue(constGas uint64) BurnFunction {
	g := constGas
	return func(_ uint64) uint64 {
		return g
	}
}

func (c BurnCode) Cost(p ...uint64) uint64 {
	x := uint64(0)
	if len(p) > 0 {
		x = p[0]
	}
	if r, ok := burnTable[c]; ok {
		return r.BurnFunction(x)
	}
	panic(fmt.Errorf("%v: %d", ErrUnknownBurnCode, c))
}

func linear(a uint64) BurnFunction {
	return func(x uint64) uint64 {
		return a * x
	}
}

func minBurn(minGasBurn uint64) BurnFunction {
	return func(currentBurnedGas uint64) uint64 {
		if minGasBurn < currentBurnedGas {
			// prevent overflow
			return 0
		}

		return minGasBurn - currentBurnedGas
	}
}
