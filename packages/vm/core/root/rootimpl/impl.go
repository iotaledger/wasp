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
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
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
func initialize(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("root.initialize.begin")

	state := ctx.State()
	stateAnchor := ctx.StateAnchor()
	contractRegistry := collections.NewMap(state, root.StateVarContractRegistry)
	creator := stateAnchor.Sender

	initConditionsCorrect :=
		stateAnchor.IsOrigin &&
			state.MustGet(root.StateVarStateInitialized) == nil &&
			ctx.Caller().Hname() == 0 &&
			creator != nil &&
			creator.Equal(ctx.Caller().Address()) &&
			contractRegistry.MustLen() == 0
	ctx.Requiref(initConditionsCorrect, "root.initialize.fail: %v", root.ErrChainInitConditionsFailed)

	assetsOnStateAnchor := iscp.NewAssets(stateAnchor.Deposit, nil)
	ctx.Requiref(len(assetsOnStateAnchor.Tokens) == 0, "root.initialize.fail: native tokens in origin output are not allowed")

	extParams := ctx.Params().Clone()

	// store 'root' into the registry
	mustStoreContract(ctx, root.Contract)
	// store 'blob' into the registry and run init
	mustStoreAndInitCoreContract(ctx, blob.Contract, nil)
	// store 'accounts' into the registry  and run init
	// passing dust assumptions
	extParams.Set(accounts.ParamDustDepositAssumptionsBin, ctx.Params().MustGet(root.ParamDustDepositAssumptionsBin))
	mustStoreAndInitCoreContract(ctx, accounts.Contract, extParams)
	// store 'blocklog' into the registry and run init
	mustStoreAndInitCoreContract(ctx, blocklog.Contract, nil)

	// store 'governance' into the registry and run init
	// passing init parameters
	extParams.Set(governance.ParamChainID, codec.EncodeChainID(ctx.ChainID()))
	// chain owner is whoever creates origin and sends the 'init' request
	extParams.Set(governance.ParamChainOwner, ctx.Caller().Bytes())
	mustStoreAndInitCoreContract(ctx, governance.Contract, extParams)

	state.Set(root.StateVarDeployPermissionsEnabled, codec.EncodeBool(true))
	state.Set(root.StateVarStateInitialized, []byte{0xFF})

	ctx.Log().Debugf("root.initialize.success")
	return nil
}

// deployContract deploys contract and calls its 'init' constructor.
// If call to the constructor returns an error or an other error occurs,
// removes smart contract form the registry as if it was never attempted to deploy
// Inputs:
// - ParamName string, the unique name of the contract in the chain. Later used as hname
// - ParamProgramHash HashValue is a hash of the blob which represents program binary in the 'blob' contract.
//     In case of hardcoded examples its an arbitrary unique hash set in the global call examples.AddProcessor
// - ParamDescription string is an arbitrary string. Defaults to "N/A"
func deployContract(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("root.deployContract.begin")
	ctx.Requiref(isAuthorizedToDeploy(ctx), "root.deployContract: deploy not permitted for: %s", ctx.Caller())

	progHash := ctx.ParamDecoder().MustGetHashValue(root.ParamProgramHash)
	description := ctx.ParamDecoder().MustGetString(root.ParamDescription, "N/A")
	name := ctx.ParamDecoder().MustGetString(root.ParamName)
	ctx.Requiref(name != "", "wrong name")

	// pass to init function all params not consumed so far
	initParams := dict.New()
	for key, value := range ctx.Params() {
		if key != root.ParamProgramHash && key != root.ParamName && key != root.ParamDescription {
			initParams.Set(key, value)
		}
	}
	// call to load VM from binary to check if it loads successfully
	err := ctx.Privileged().TryLoadContract(progHash)
	ctx.Requiref(err == nil, "root.deployContract.fail 1: %v", err)

	// VM loaded successfully. Storing contract in the registry and calling constructor
	mustStoreContractRecord(ctx, &root.ContractRecord{
		ProgramHash: progHash,
		Description: description,
		Name:        name,
		Creator:     ctx.Caller(),
	})
	ctx.Call(iscp.Hn(name), iscp.EntryPointInit, initParams, nil)
	ctx.Event(fmt.Sprintf("[deploy] name: %s hname: %s, progHash: %s, dscr: '%s'",
		name, iscp.Hn(name), progHash.String(), description))
	return nil
}

// findContract view finds and returns encoded record of the contract
// Input:
// - ParamHname
// Output:
// - ParamData
func findContract(ctx iscp.SandboxView) dict.Dict {
	hname := ctx.ParamDecoder().MustGetHname(root.ParamHname)
	rec := root.FindContract(ctx.State(), hname)
	ret := dict.New()
	found := rec != nil
	ret.Set(root.ParamContractFound, codec.EncodeBool(found))
	if found {
		ret.Set(root.ParamContractRecData, rec.Bytes())
	}
	return ret
}

// grantDeployPermission grants permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func grantDeployPermission(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner("root.grantDeployPermissions: not authorized")

	deployer := ctx.ParamDecoder().MustGetAgentID(root.ParamDeployer)

	collections.NewMap(ctx.State(), root.StateVarDeployPermissions).MustSetAt(deployer.Bytes(), []byte{0xFF})
	ctx.Event(fmt.Sprintf("[grant deploy permission] to agentID: %s", deployer))
	return nil
}

// revokeDeployPermission revokes permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func revokeDeployPermission(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner("root.revokeDeployPermissions: not authorized")

	deployer := ctx.ParamDecoder().MustGetAgentID(root.ParamDeployer)

	collections.NewMap(ctx.State(), root.StateVarDeployPermissions).MustDelAt(deployer.Bytes())
	ctx.Event(fmt.Sprintf("[revoke deploy permission] from agentID: %s", deployer))
	return nil
}

func getContractRecords(ctx iscp.SandboxView) dict.Dict {
	src := root.GetContractRegistryR(ctx.State())

	ret := dict.New()
	dst := collections.NewMap(ret, root.StateVarContractRegistry)
	src.MustIterate(func(elemKey []byte, value []byte) bool {
		dst.MustSetAt(elemKey, value)
		return true
	})

	return ret
}

func requireDeployPermissions(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner("root.revokeDeployPermissions: not authorized")
	permissionsEnabled := ctx.ParamDecoder().MustGetBool(root.ParamDeployPermissionsEnabled)
	ctx.State().Set(root.StateVarDeployPermissionsEnabled, codec.EncodeBool(permissionsEnabled))
	return nil
}
