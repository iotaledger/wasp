package inccounter

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreIncCounter)

var (
	FuncIncCounter = coreutil.NewEP1(Contract, "incCounter",
		coreutil.FieldOptional[int64]("value"),
	)
	FuncIncAndRepeatOnceAfter2s = coreutil.NewEP0(Contract, "incAndRepeatOnceAfter5s")
	FuncIncAndRepeatMany        = coreutil.NewEP2(Contract, "incAndRepeatMany",
		coreutil.FieldOptional[int64]("value"),
		coreutil.FieldOptional[int64]("nTimes"),
	)
	ViewGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.Field[int64]("counter"),
	)
)
