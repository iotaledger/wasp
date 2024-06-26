package contracts

import (
	_ "embed"

	"github.com/iotaledger/wasp/sui-go/move"
)

// If you change any of the move contracts, you must recompile.  You will need
// the `sui` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate sh -c "cd ./isc && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed isc/bytecode.json
var iscBytecodeJSON []byte

func ISC() move.PackageBytecode {
	return move.DecodePackageBytecode(iscBytecodeJSON)
}

//go:generate sh -c "cd ./sdk_verify && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed sdk_verify/bytecode.json
var sdkVerifyBytecodeJSON []byte

func SDKVerify() move.PackageBytecode {
	return move.DecodePackageBytecode(sdkVerifyBytecodeJSON)
}

//go:generate sh -c "cd ./testcoin && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed testcoin/bytecode.json
var testcoinBytecodeJSON []byte

func Testcoin() move.PackageBytecode {
	return move.DecodePackageBytecode(testcoinBytecodeJSON)
}
