package inccounter

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreIncCounter)

var (
	FuncIncCounter = coreutil.NewEP1(Contract, "incCounter",
		coreutil.FieldOptional[int64](),
	)
	FuncIncAndRepeatOnceAfter2s = coreutil.NewEP0(Contract, "incAndRepeatOnceAfter5s")
	FuncIncAndRepeatMany        = coreutil.NewEP2(Contract, "incAndRepeatMany",
		coreutil.FieldOptional[int64](),
		coreutil.FieldOptional[int64](),
	)
	ViewGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.Field[int64](),
	)
)
