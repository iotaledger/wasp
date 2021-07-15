package _default

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

const description = "Default core contract"

var (
	Interface = &coreutil.ContractInterface{
		Name:        coreutil.CoreContractDefault,
		Description: description,
	}
)

func init() {
	Interface.WithFunctions(nil, []coreutil.ContractFunctionInterface{})
}
