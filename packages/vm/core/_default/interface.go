package _default

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var (
	Interface = coreutil.NewContractInterface(coreutil.CoreContractDefault, "Default core contract")
	Processor = Interface.Processor(nil)
)
