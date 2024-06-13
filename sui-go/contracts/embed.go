package contracts

import (
	_ "embed"
	"encoding/json"

	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// If you change any of the move contracts, you must recompile.  You will need
// the `sui` binary installed in your system. Then, simply run `go generate`
// in this directory.

type MoveBytecode struct {
	Modules      []*sui_types.Base64Data `json:"modules"`
	Dependencies []*sui_types.SuiAddress `json:"dependencies"`
	Digest       []int                   `json:"digest"`
}

//go:generate sh -c "cd ./isc && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed isc/bytecode.json
var iscBytecodeJSON []byte

func ISC() MoveBytecode {
	return Load(iscBytecodeJSON)
}

//go:generate sh -c "cd ./sdk_verify && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed sdk_verify/bytecode.json
var sdkVerifyBytecodeJSON []byte

func SDKVerify() MoveBytecode {
	return Load(sdkVerifyBytecodeJSON)
}

//go:generate sh -c "cd ./testcoin && sui move build --dump-bytecode-as-base64 > bytecode.json"
//go:embed testcoin/bytecode.json
var testcoinBytecodeJSON []byte

func Testcoin() MoveBytecode {
	return Load(testcoinBytecodeJSON)
}

func Load(bytecodeJSON []byte) (ret MoveBytecode) {
	err := json.Unmarshal(bytecodeJSON, &ret)
	if err != nil {
		panic(err)
	}
	return
}