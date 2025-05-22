package inccounter

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreIncCounter)

var (
	FuncIncCounter = coreutil.NewEP1(Contract, "incCounter",
		coreutil.FieldOptional[int64]("value"),
	)
	ViewGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.Field[int64]("counter"),
	)
)
