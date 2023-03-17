package root

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot, "Root Contract")

// state variables
const (
	StateVarSchemaVersion = "v"

	StateVarContractRegistry          = "r"
	StateVarDeployPermissionsEnabled  = "a"
	StateVarDeployPermissions         = "p"
	StateVarBlockContextSubscriptions = "b"
)

// param variables
const (
	ParamDeployer                     = "dp"
	ParamHname                        = "hn"
	ParamName                         = "nm"
	ParamProgramHash                  = "ph"
	ParamContractRecData              = "dt"
	ParamContractFound                = "cf"
	ParamDescription                  = "ds"
	ParamDeployPermissionsEnabled     = "de"
	ParamStorageDepositAssumptionsBin = "db"
)

// function names
var (
	FuncDeployContract           = coreutil.Func("deployContract")
	FuncGrantDeployPermission    = coreutil.Func("grantDeployPermission")
	FuncRevokeDeployPermission   = coreutil.Func("revokeDeployPermission")
	FuncRequireDeployPermissions = coreutil.Func("requireDeployPermissions")
	ViewFindContract             = coreutil.ViewFunc("findContract")
	ViewGetContractRecords       = coreutil.ViewFunc("getContractRecords")
)
