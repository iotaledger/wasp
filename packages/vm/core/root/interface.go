package root

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/samber/lo"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot)

var (
	// Funcs
	FuncGrantDeployPermission = coreutil.NewEP1(Contract, "grantDeployPermission",
		coreutil.Field[isc.AgentID](),
	)
	FuncRevokeDeployPermission = coreutil.NewEP1(Contract, "revokeDeployPermission",
		coreutil.Field[isc.AgentID](),
	)
	FuncRequireDeployPermissions = coreutil.NewEP1(Contract, "requireDeployPermissions",
		coreutil.Field[bool](),
	)

	// Views
	ViewFindContract = coreutil.NewViewEP12(Contract, "findContract",
		coreutil.Field[isc.Hname](),
		coreutil.Field[bool](),
		coreutil.FieldOptional[*ContractRecord](),
	)
	ViewGetContractRecords = coreutil.NewViewEP01(Contract, "getContractRecords",
		coreutil.Field[[]lo.Tuple2[*isc.Hname, *ContractRecord]](),
	)
)

// state variables
const (
	varSchemaVersion            = "v" // covered in: TestDeployNativeContract
	varContractRegistry         = "r" // covered in: TestDeployNativeContract
	varDeployPermissionsEnabled = "a" // covered in: TestDeployNativeContract
	varDeployPermissions        = "p" // covered in: TestDeployNativeContract
)
