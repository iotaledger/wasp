// 'root' a core contract on the chain. It is responsible for:
// - initial setup of the chain during chain deployment
// - maintaining of core parameters of the chain
// - maintaining (setting, delegating) chain owner ID
// - maintaining (granting, revoking) smart contract deployment rights
// - deployment of smart contracts on the chain and maintenance of contract registry
package rootimpl

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

var Processor = root.Contract.Processor(initialize,
	root.FuncDeployContract.WithHandler(deployContract),
	root.FuncGrantDeployPermission.WithHandler(grantDeployPermission),
	root.FuncRevokeDeployPermission.WithHandler(revokeDeployPermission),
	root.FuncFindContract.WithHandler(findContract),
	root.FuncGetContractRecords.WithHandler(getContractRecords),
	root.FuncRequireDeployPermissions.WithHandler(requireDeployPermissions),
)

// initialize handles constructor, the "init" request. This is the first call to the chain
// if it fails, chain is not initialized. Does the following:
// - stores chain ID and chain description in the state
// - sets state ownership to the caller
// - creates record in the registry for the 'root' itself
// - deploys other core contracts: 'accounts', 'blob', 'blocklog' by creating records in the registry and calling constructors
// Input:
// - ParamChainID iscp.ChainID. ID of the chain. Cannot be changed
// - ParamDescription string defaults to "N/A"
// - ParamDustDepositAssumptionsBin encoded assumptions about minimum dust deposit for internal outputs
func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.initialize.begin")
	state := ctx.State()

	ctx.Require(state.MustGet(root.StateVarStateInitialized) == nil, "root.initialize.fail: already initialized")
	ctx.Require(ctx.Caller().Hname() == 0, "root.init.fail: chain deployer can't be another smart contract")
	creator := ctx.StateAnchor().Sender
	ctx.Require(creator != nil && creator.Equal(ctx.Caller().Address()), "only creator of the origin can send the 'init' request")

	contractRegistry := collections.NewMap(state, root.StateVarContractRegistry)
	ctx.Require(contractRegistry.MustLen() == 0, "root.initialize.fail: registry not empty")

	dustAssumptionsBin, err := ctx.Params().Get(root.ParamDustDepositAssumptionsBin)
	ctx.RequireNoError(err)
	_, err = vmtxbuilder.InternalDustDepositAssumptionFromBytes(dustAssumptionsBin)
	ctx.RequireNoError(err, "cannot initialize chain: 'dust deposit assumptions' parameter not specified or wrong")

	mustStoreContract(ctx, root.Contract)
	mustStoreAndInitCoreContract(ctx, blob.Contract)
	mustStoreAndInitCoreContract(ctx, accounts.Contract)
	mustStoreAndInitCoreContract(ctx, blocklog.Contract)
	govParams := ctx.Params().Clone()
	govParams.Set(governance.ParamChainID, codec.EncodeChainID(ctx.ChainID()))
	// chain owner is whoever creates origin and sends the 'init' request
	govParams.Set(governance.ParamChainOwner, ctx.Caller().Bytes())
	mustStoreAndInitCoreContract(ctx, governance.Contract, govParams)

	state.Set(root.StateVarDeployPermissionsEnabled, codec.EncodeBool(true))
	state.Set(root.StateVarDustDepositAssumptions, dustAssumptionsBin)
	state.Set(root.StateVarStateInitialized, []byte{0xFF})

	ctx.Log().Debugf("root.initialize.success")
	return nil, nil
}

// deployContract deploys contract and calls its 'init' constructor.
// If call to the constructor returns an error or an other error occurs,
// removes smart contract form the registry as if it was never attempted to deploy
// Inputs:
// - ParamName string, the unique name of the contract in the chain. Later used as hname
// - ParamProgramHash HashValue is a hash of the blob which represents program binary in the 'blob' contract.
//     In case of hardcoded examples its an arbitrary unique hash set in the global call examples.AddProcessor
// - ParamDescription string is an arbitrary string. Defaults to "N/A"
func deployContract(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.deployContract.begin")
	ctx.Require(isAuthorizedToDeploy(ctx), "root.deployContract: deploy not permitted for: %s", ctx.Caller())

	params := kvdecoder.New(ctx.Params(), ctx.Log())

	progHash := params.MustGetHashValue(root.ParamProgramHash)
	description := params.MustGetString(root.ParamDescription, "N/A")
	name := params.MustGetString(root.ParamName)
	ctx.Require(name != "", "wrong name")

	// pass to init function all params not consumed so far
	initParams := dict.New()
	for key, value := range ctx.Params() {
		if key != root.ParamProgramHash && key != root.ParamName && key != root.ParamDescription {
			initParams.Set(key, value)
		}
	}
	// call to load VM from binary to check if it loads successfully
	err := ctx.DeployContract(progHash, "", "", nil)
	ctx.Require(err == nil, "root.deployContract.fail 1: %v", err)

	// VM loaded successfully. Storing contract in the registry and calling constructor
	mustStoreContractRecord(ctx, &root.ContractRecord{
		ProgramHash: progHash,
		Description: description,
		Name:        name,
		Creator:     ctx.Caller(),
	})
	_, err = ctx.Call(iscp.Hn(name), iscp.EntryPointInit, initParams, nil)
	ctx.RequireNoError(err)

	ctx.Event(fmt.Sprintf("[deploy] name: %s hname: %s, progHash: %s, dscr: '%s'",
		name, iscp.Hn(name), progHash.String(), description))
	return nil, nil
}

// findContract view finds and returns encoded record of the contract
// Input:
// - ParamHname
// Output:
// - ParamData
func findContract(ctx iscp.SandboxView) (dict.Dict, error) {
	hname, err := kvdecoder.New(ctx.Params()).GetHname(root.ParamHname)
	if err != nil {
		return nil, err
	}
	rec := root.FindContract(ctx.State(), hname)
	ret := dict.New()
	found := rec != nil
	ret.Set(root.ParamContractFound, codec.EncodeBool(found))
	if found {
		ret.Set(root.ParamContractRecData, rec.Bytes())
	}
	return ret, nil
}

// grantDeployPermission grants permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func grantDeployPermission(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.RequireCallerIsChainOwner("root.grantDeployPermissions: not authorized")

	deployer := kvdecoder.New(ctx.Params(), ctx.Log()).MustGetAgentID(root.ParamDeployer)

	collections.NewMap(ctx.State(), root.StateVarDeployPermissions).MustSetAt(deployer.Bytes(), []byte{0xFF})
	ctx.Event(fmt.Sprintf("[grant deploy permission] to agentID: %s", deployer))
	return nil, nil
}

// revokeDeployPermission revokes permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func revokeDeployPermission(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.RequireCallerIsChainOwner("root.revokeDeployPermissions: not authorized")

	deployer := kvdecoder.New(ctx.Params(), ctx.Log()).MustGetAgentID(root.ParamDeployer)

	collections.NewMap(ctx.State(), root.StateVarDeployPermissions).MustDelAt(deployer.Bytes())
	ctx.Event(fmt.Sprintf("[revoke deploy permission] from agentID: %s", deployer))
	return nil, nil
}

func getContractRecords(ctx iscp.SandboxView) (dict.Dict, error) {
	src := collections.NewMapReadOnly(ctx.State(), root.StateVarContractRegistry)

	ret := dict.New()
	dst := collections.NewMap(ret, root.StateVarContractRegistry)
	src.MustIterate(func(elemKey []byte, value []byte) bool {
		dst.MustSetAt(elemKey, value)
		return true
	})

	return ret, nil
}

func requireDeployPermissions(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.RequireCallerIsChainOwner("root.revokeDeployPermissions: not authorized")
	permissionsEnabled := kvdecoder.New(ctx.Params()).MustGetBool(root.ParamDeployPermissionsEnabled)
	ctx.State().Set(root.StateVarDeployPermissionsEnabled, codec.EncodeBool(permissionsEnabled))
	return nil, nil
}
