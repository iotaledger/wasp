package root

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var Contract = coreutil.NewContract(coreutil.CoreContractRoot, "Root Contract")

// state variables
const (
	StateVarContractRegistry          = "r"
	StateVarDeployPermissionsEnabled  = "a"
	StateVarDeployPermissions         = "p"
	StateVarStateInitialized          = "i"
	StateVarBlockContextSubscriptions = "b"
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
	ParamBlockContextOpenFunc      = "bco"
	ParamBlockContextCloseFunc     = "bcc"
)

// ParamEVM allows to pass init parameters to the EVM core contract, by decorating
// them with a prefix. For example:
//  ParamEVM(evm.FieldBlockKeepAmount)
func ParamEVM(k kv.Key) kv.Key { return "evm" + k }

// function names
var (
	FuncDeployContract           = coreutil.Func("deployContract")
	FuncGrantDeployPermission    = coreutil.Func("grantDeployPermission")
	FuncRevokeDeployPermission   = coreutil.Func("revokeDeployPermission")
	FuncRequireDeployPermissions = coreutil.Func("requireDeployPermissions")
	ViewFindContract             = coreutil.ViewFunc("findContract")
	ViewGetContractRecords       = coreutil.ViewFunc("getContractRecords")
	FuncSubscribeBlockContext    = coreutil.Func("subscribeBlockContext")
)

var ErrChainInitConditionsFailed = coreerrors.Register("root.init can't be called in this state").Create()
