package gas

import "golang.org/x/xerrors"

const (
	MaxGasPerBlock = uint64(10_000_000)
	MaxGasPerCall  = uint64(500_000)
)

const (
	BurnCodeStorage1P = BurnCode(iota)
	BurnCodeReadFromState1P
	BurnCodeCallTargetNotFound
	BurnCodeGetContext
	BurnCodeGetCallerData
	BurnCodeGetAllowance
	BurnCodeGetStateAnchorInfo
	BurnCodeGetBalance
	BurnCodeCallContract
	BurnCodeDeployContract
	BurnCodeEmitEventFixed
	BurnCodeTransferAllowance
	BurnCodeSendL1Request

	// Sandbox util codes

	BurnCodeUtilsHashingBlake2b
	BurnCodeUtilsHashingSha3
	BurnCodeUtilsHashingHname
	BurnCodeUtilsBase58Encode
	BurnCodeUtilsBase58Decode
	BurnCodeUtilsED25519ValidSig
	BurnCodeUtilsED25519AddrFromPubKey
	BurnCodeUtilsBLSValidSignature
	BurnCodeUtilsBLSAddrFromPubKey
	BurnCodeUtilsBLSAggregateBLS1P

	BurnCodeWasm1P
	BurnCodeMinimumGasPerRequest
)

// burnTable contains all possible burn codes with their burn value computing functions
var burnTable = BurnTable{
	BurnCodeCallTargetNotFound:         {"target n/f", constValue(10)},
	BurnCodeGetContext:                 {"context", constValue(10)},
	BurnCodeGetCallerData:              {"caller", constValue(10)},
	BurnCodeGetStateAnchorInfo:         {"anchor", constValue(10)},
	BurnCodeGetBalance:                 {"balance", constValue(20)},
	BurnCodeCallContract:               {"call", constValue(10)},
	BurnCodeEmitEventFixed:             {"event", constValue(10)},
	BurnCodeGetAllowance:               {"allowance", constValue(10)},
	BurnCodeTransferAllowance:          {"transfer", constValue(10)},
	BurnCodeSendL1Request:              {"send", linear(Coef1Send)},
	BurnCodeDeployContract:             {"deploy", constValue(10)},
	BurnCodeStorage1P:                  {"storage", linear(100)},
	BurnCodeReadFromState1P:            {"state read", linear(1)},
	BurnCodeWasm1P:                     {"wasm", linear(1)},
	BurnCodeUtilsHashingBlake2b:        {"blake2b", constValue(50)},
	BurnCodeUtilsHashingSha3:           {"sha3", constValue(80)},
	BurnCodeUtilsHashingHname:          {"hname", constValue(50)},
	BurnCodeUtilsBase58Encode:          {"base58enc", linear(50)},
	BurnCodeUtilsBase58Decode:          {"base58dec", linear(5)},
	BurnCodeUtilsED25519ValidSig:       {"ed25517 valid", constValue(200)},
	BurnCodeUtilsED25519AddrFromPubKey: {"ed25517 addr", constValue(50)},
	BurnCodeUtilsBLSValidSignature:     {"bls valid", constValue(2000)},
	BurnCodeUtilsBLSAddrFromPubKey:     {"bls addr", constValue(50)},
	BurnCodeUtilsBLSAggregateBLS1P:     {"bls aggregate", linear(CoefBLSAggregate)},
	BurnCodeMinimumGasPerRequest:       {"minimum gas per request", constValue(100)}, // TODO maybe make it configurable (gov contract?)
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

func (c BurnCode) Cost(p ...int) uint64 {
	x := uint64(0)
	if len(p) > 0 {
		x = uint64(p[0])
	}
	if r, ok := burnTable[c]; ok {
		return r.BurnFunction(x)
	}
	panic(xerrors.Errorf("%v: %d", ErrUnknownBurnCode, c))
}

func linear(a uint64) BurnFunction {
	return func(x uint64) uint64 {
		return a * x
	}
}
