package move

import (
	"encoding/json"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
)

// PackageBytecode is the decoded information produced by the command `sui move build --dump-bytecode-as-base64`
type PackageBytecode struct {
	Modules      []*sui.Base64Data `json:"modules"`
	Dependencies []*sui.Address    `json:"dependencies"`
	Digest       []int             `json:"digest"`
}

// DecodePackageBytecode decodes the output of the command `sui move build --dump-bytecode-as-base64`
func DecodePackageBytecode(bytecodeJSON []byte) (ret PackageBytecode) {
	err := json.Unmarshal(bytecodeJSON, &ret)
	if err != nil {
		panic(err)
	}
	return
}
