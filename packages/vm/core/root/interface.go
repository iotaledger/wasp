package root

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot)

var (
	// Views
	ViewFindContract = coreutil.NewViewEP12(Contract, "findContract",
		coreutil.Field[isc.Hname]("contractHName"),
		coreutil.Field[bool]("exists"),
		coreutil.FieldOptional[*ContractRecord]("contractRecord"),
	)
	ViewGetContractRecords = coreutil.NewViewEP01(Contract, "getContractRecords",
		coreutil.Field[[]lo.Tuple2[*isc.Hname, *ContractRecord]]("contractRecords"),
	)
)

// state variables
const (
	VarSchemaVersion    = "v" // covered in: TestDeployNativeContract
	VarContractRegistry = "r" // covered in: TestDeployNativeContract
)
