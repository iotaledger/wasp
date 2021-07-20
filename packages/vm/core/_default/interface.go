package _default

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var (
	Contract = coreutil.NewContract(coreutil.CoreContractDefault, "Default core contract")
	Processor = Contract.Processor(nil)
)
