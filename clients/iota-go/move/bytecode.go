package move

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// PackageBytecode is the decoded information produced by the command `iota move build --dump-bytecode-as-base64`
type PackageBytecode struct {
	Modules      []*iotago.Base64Data `json:"modules"`
	Dependencies []*iotago.Address    `json:"dependencies"`
	Digest       []int                `json:"digest"`
}

// DecodePackageBytecode decodes the output of the command `iota move build --dump-bytecode-as-base64`
func DecodePackageBytecode(bytecodeJSON []byte) (ret PackageBytecode) {
	err := json.Unmarshal(bytecodeJSON, &ret)
	if err != nil {
		panic(err)
	}
	return
}
