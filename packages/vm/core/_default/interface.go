package _default

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
)

const (
	Name        = "_default"
	description = "Default Contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
	}
)

func init() {
	Interface.WithFunctions(nil, []coreutil.ContractFunctionInterface{})
}
