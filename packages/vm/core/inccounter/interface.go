package inccounter

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

var Contract = coreutil.NewContract(coreutil.CoreIncCounter)

var (
	FuncIncCounter = coreutil.NewEP1(Contract, "incCounter",
		coreutil.FieldWithCodecOptional(codec.Int64),
	)
	FuncIncAndRepeatOnceAfter2s = coreutil.NewEP0(Contract, "incAndRepeatOnceAfter5s")
	FuncIncAndRepeatMany        = coreutil.NewEP2(Contract, "incAndRepeatMany",
		coreutil.FieldWithCodecOptional(codec.Int64),
		coreutil.FieldWithCodecOptional(codec.Int64),
	)
	ViewGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.FieldWithCodec(codec.Int64),
	)
)
