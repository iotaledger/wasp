package gas

import "golang.org/x/xerrors"

const (
	BurnStorage1P = BurnCode(iota)
	BurnCallTargetNotFound
	BurnGetContext
	BurnGetCallerData
	BurnGetAllowance
	BurnGetStateAnchorInfo
	BurnGetBalance
	BurnCallContract
	BurnDeployContract
	BurnEmitEventFixed
	BurnTransferAllowance
	BurnSendL1Request

	// Sandbox util codes

	BurnUtilsHashingBlake2b
	BurnUtilsHashingSha3
	BurnUtilsHashingHname
	BurnUtilsBase58Encode
	BurnUtilsBase58Decode
	BurnUtilsED25519ValidSig
	BurnUtilsED25519AddrFromPubKey
	BurnUtilsBLSValidSignature
	BurnUtilsBLSAddrFromPubKey
	BurnUtilsBLSAggregateBLS1P

	BurnWasm1P
)

// burnTable contains all possible burn codes with their burn value computing functions
var burnTable = BurnTable{
	BurnCallTargetNotFound:         {"target n/f", constValue(10)},
	BurnGetContext:                 {"context", constValue(10)},
	BurnGetCallerData:              {"caller", constValue(10)},
	BurnGetStateAnchorInfo:         {"anchor", constValue(10)},
	BurnGetBalance:                 {"balance", constValue(20)},
	BurnCallContract:               {"call", constValue(10)},
	BurnEmitEventFixed:             {"event", constValue(10)},
	BurnGetAllowance:               {"allowance", constValue(10)},
	BurnTransferAllowance:          {"transfer", constValue(10)},
	BurnSendL1Request:              {"send", constValue(10)},
	BurnDeployContract:             {"deploy", constValue(10)},
	BurnStorage1P:                  {"storage", linear()},
	BurnWasm1P:                     {"wasm", linear()},
	BurnUtilsHashingBlake2b:        {"blake2b", constValue(50)},
	BurnUtilsHashingSha3:           {"sha3", constValue(80)},
	BurnUtilsHashingHname:          {"hname", constValue(50)},
	BurnUtilsBase58Encode:          {"base58enc", constValue(50)},
	BurnUtilsBase58Decode:          {"base58dec", constValue(50)},
	BurnUtilsED25519ValidSig:       {"ed25517 valid", constValue(200)},
	BurnUtilsED25519AddrFromPubKey: {"ed25517 addr", constValue(50)},
	BurnUtilsBLSValidSignature:     {"bls valid", constValue(2000)},
	BurnUtilsBLSAddrFromPubKey:     {"bls addr", constValue(50)},
	BurnUtilsBLSAggregateBLS1P:     {"bls aggregate", linear(400)},
}

func constValue(constGas uint64) BurnFunction {
	g := constGas
	return func(_ BurnCode, _ []int) uint64 {
		return g
	}
}

func notImplemented() BurnFunction {
	return func(code BurnCode, _ []int) uint64 {
		panic(xerrors.Errorf("burn code %d not implemented", code))
	}
}

func (c BurnCode) Value(p ...int) uint64 {
	return Value(c, p...)
}

// linear takes A and B as parameters ans construct a closure which returns gas as a linear function A*x+B of the parameter x.
// linear() without parameters returns closure corresponding to identity function 1*x+0
func linear(p ...int) BurnFunction {
	switch len(p) {
	case 0:
		// burn value == X
		return func(_ BurnCode, x []int) uint64 {
			if len(x) != 1 {
				panic(ErrInLinear1ParameterExpected)
			}
			return uint64(x[0])
		}
	case 1:
		a := uint64(p[0])
		// burn value == a*X
		return func(_ BurnCode, x []int) uint64 {
			if len(x) != 1 {
				panic(ErrInLinear1ParameterExpected)
			}
			return a * uint64(x[0])
		}
	case 2:
		a := uint64(p[0])
		b := uint64(p[1])
		if a == 0 {
			// burn value = 0*X+b
			return constValue(b)
		}
		if b == 0 {
			// burn value = a*X
			return linear(p[0])
		}
		// burn value = a*X+b
		return func(_ BurnCode, x []int) uint64 {
			if len(x) != 1 {
				panic(ErrInLinear1ParameterExpected)
			}
			return a*uint64(x[0]) + b
		}
	default:
		panic("function requires 0, 1 or 2 parameter")
	}
}
