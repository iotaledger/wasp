package root

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var (
	Contract = coreutil.NewContract(coreutil.CoreContractRoot, "Root Contract")
)

// state variables
const (
	StateVarContractRegistry         = "r"
	StateVarDeployPermissionsEnabled = "a"
	StateVarDeployPermissions        = "p"
	StateVarStateInitialized         = "i"
	StateVarDustDepositAssumptions   = "d"
)

// param variables
const (
	ParamDeployer                  = "dp"
	ParamHname                     = "hn"
	ParamName                      = "nm"
	ParamProgramHash               = "ph"
	ParamContractRecData           = "dt"
	ParamContractFound             = "cf"
	ParamDescription               = "ds"
	ParamDeployPermissionsEnabled  = "de"
	ParamDustDepositAssumptionsBin = "db"
)

// function names
var (
	FuncDeployContract           = coreutil.Func("deployContract")
	FuncGrantDeployPermission    = coreutil.Func("grantDeployPermission")
	FuncRevokeDeployPermission   = coreutil.Func("revokeDeployPermission")
	FuncRequireDeployPermissions = coreutil.Func("requireDeployPermissions")
	FuncFindContract             = coreutil.ViewFunc("findContract")
	FuncGetContractRecords       = coreutil.ViewFunc("getContractRecords")
)

func GasToDeploy(programHash hashing.HashValue) uint64 {
	return gas.CoreRootDeployContract
}
