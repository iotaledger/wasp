package gas

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

var burnTable = BurnTable{
	// constant values
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
	BurnStorage1P:                  {"storage", proportion()},
	BurnWasm1P:                     {"storage", proportion()},
	BurnUtilsHashingBlake2b:        {"blake2b", constValue(50)},
	BurnUtilsHashingSha3:           {"sha3", constValue(80)},
	BurnUtilsHashingHname:          {"hname", constValue(50)},
	BurnUtilsBase58Encode:          {"base58enc", constValue(50)},
	BurnUtilsBase58Decode:          {"base58dec", constValue(50)},
	BurnUtilsED25519ValidSig:       {"ed25517 valid", constValue(200)},
	BurnUtilsED25519AddrFromPubKey: {"ed25517 addr", constValue(50)},
	BurnUtilsBLSValidSignature:     {"bls valid", constValue(2000)},
	BurnUtilsBLSAddrFromPubKey:     {"ed25517 addr", constValue(50)},
	BurnUtilsBLSAggregateBLS1P:     {"bls aggregate", proportion(400)},
}

// proportion construct a closure which returns gas proportional to the first parameter
// if proportion() without parameters return close which return gas equal to the first parameter
func proportion(p ...int) BurnFunction {
	if len(p) == 0 {
		return func(_ BurnCode, params []int) uint64 {
			if len(params) != 1 {
				panic("'proportion' burn function requires exactly 1 parameter")
			}
			return uint64(params[0])
		}
	}
	if p[0] <= 0 {
		panic("proportion: wrong parameter")
	}
	c := uint64(p[0])
	return func(_ BurnCode, params []int) uint64 {
		if len(params) != 1 {
			panic("'proportion' burn function requires exactly 1 parameter")
		}
		return c * uint64(params[0])
	}
}
