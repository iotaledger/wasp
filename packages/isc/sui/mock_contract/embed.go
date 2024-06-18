package mock_contract

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
//go:embed bytecode.json
var iscBytecodeJSON []byte

func MockISCContract() MoveBytecode {
	return Load(iscBytecodeJSON)
}

func Load(bytecodeJSON []byte) (ret MoveBytecode) {
	err := json.Unmarshal(bytecodeJSON, &ret)
	if err != nil {
		panic(err)
	}
	return
}
