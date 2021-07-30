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
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

var Processor = root.Contract.Processor(initialize,
	root.FuncClaimChainOwnership.WithHandler(claimChainOwnership),
	root.FuncDelegateChainOwnership.WithHandler(delegateChainOwnership),
	root.FuncDeployContract.WithHandler(deployContract),
	root.FuncGrantDeployPermission.WithHandler(grantDeployPermission),
	root.FuncRevokeDeployPermission.WithHandler(revokeDeployPermission),
	root.FuncSetContractFee.WithHandler(setContractFee),
	root.FuncSetChainInfo.WithHandler(setChainConfig),
	root.FuncFindContract.WithHandler(findContract),
	root.FuncGetChainInfo.WithHandler(getChainInfo),
	root.FuncGetFeeInfo.WithHandler(getFeeInfo),
	root.FuncGetMaxBlobSize.WithHandler(getMaxBlobSize),
)

// initialize handles constructor, the "init" request. This is the first call to the chain
// if it fails, chain is not initialized. Does the following:
// - stores chain ID and chain description in the state
// - sets state ownership to the caller
// - creates record in the registry for the 'root' itself
// - deploys other core contracts: 'accounts', 'blob', 'blocklog' by creating records in the registry and calling constructors
// Input:
// - ParamChainID iscp.ChainID. ID of the chain. Cannot be changed
// - ParamChainColor ledgerstate.Color
// - ParamChainAddress ledgerstate.Address
// - ParamDescription string defaults to "N/A"
// - ParamFeeColor ledgerstate.Color fee color code. Defaults to IOTA color. It cannot be changed
func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.initialize.begin")
	state := ctx.State()
	a := assert.NewAssert(ctx.Log())

	a.Require(state.MustGet(root.VarStateInitialized) == nil, "root.initialize.fail: already initialized")
	a.Require(ctx.Caller().Hname() == 0, "root.init.fail: chain deployer can't be another smart contract")

	// retrieving init parameters
	// -- chain ID
	params := kvdecoder.New(ctx.Params(), ctx.Log())

	chainID := params.MustGetChainID(root.ParamChainID)
	chainDescription := params.MustGetString(root.ParamDescription, "N/A")
	feeColor := params.MustGetColor(root.ParamFeeColor, colored.IOTA)
	feeColorSet := feeColor != colored.IOTA

	contractRegistry := collections.NewMap(state, root.VarContractRegistry)
	a.Require(contractRegistry.MustLen() == 0, "root.initialize.fail: registry not empty")

	mustStoreContract(ctx, _default.Contract, a)
	mustStoreContract(ctx, root.Contract, a)
	mustStoreAndInitCoreContract(ctx, blob.Contract, a)
	mustStoreAndInitCoreContract(ctx, accounts.Contract, a)
	mustStoreAndInitCoreContract(ctx, blocklog.Contract, a)
	mustStoreAndInitCoreContract(ctx, governance.Contract, a)

	state.Set(root.VarStateInitialized, []byte{0xFF})
	state.Set(root.VarChainID, codec.EncodeChainID(*chainID))
	state.Set(root.VarChainOwnerID, codec.EncodeAgentID(ctx.Caller())) // chain owner is whoever sends init request
	state.Set(root.VarDescription, codec.EncodeString(chainDescription))

	state.Set(root.VarMaxBlobSize, codec.Encode(root.DefaultMaxBlobSize))
	state.Set(root.VarMaxEventSize, codec.Encode(root.DefaultMaxEventSize))
	state.Set(root.VarMaxEventsPerReq, codec.Encode(root.DefaultMaxEventsPerRequest))

	if feeColorSet {
		state.Set(root.VarFeeColor, codec.EncodeColor(feeColor))
	}

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
	if !isAuthorizedToDeploy(ctx) {
		return nil, fmt.Errorf("root.deployContract: deploy not permitted for: %s", ctx.Caller())
	}
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	progHash := params.MustGetHashValue(root.ParamProgramHash)
	description := params.MustGetString(root.ParamDescription, "N/A")
	name := params.MustGetString(root.ParamName)
	a.Require(name != "", "wrong name")

	// pass to init function all params not consumed so far
	initParams := dict.New()
	for key, value := range ctx.Params() {
		if key != root.ParamProgramHash && key != root.ParamName && key != root.ParamDescription {
			initParams.Set(key, value)
		}
	}
	// calls to loads VM from binary to check if it loads successfully
	err := ctx.DeployContract(progHash, "", "", nil)
	a.Require(err == nil, "root.deployContract.fail 1: %v", err)

	// VM loaded successfully. Storing contract in the registry and calling constructor
	mustStoreContractRecord(ctx, &root.ContractRecord{
		ProgramHash: progHash,
		Description: description,
		Name:        name,
		Creator:     ctx.Caller(),
	}, a)
	_, err = ctx.Call(iscp.Hn(name), iscp.EntryPointInit, initParams, nil)
	a.RequireNoError(err)

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
	params := kvdecoder.New(ctx.Params())
	hname, err := params.GetHname(root.ParamHname)
	if err != nil {
		return nil, err
	}
	rec, found := FindContract(ctx.State(), hname)
	ret := dict.New()
	ret.Set(root.ParamContractRecData, rec.Bytes())
	var foundByte [1]byte
	if found {
		foundByte[0] = 0xFF
	}
	ret.Set(root.ParamContractFound, foundByte[:])
	return ret, nil
}

// getChainInfo view returns general info about the chain: chain ID, chain owner ID,
// description and the whole contract registry
// Input: none
// Output:
// - VarChainID - ChainID
// - VarChainOwnerID - AgentID
// - VarDescription - string
// - VarContractRegistry: a map of contract registry
func getChainInfo(ctx iscp.SandboxView) (dict.Dict, error) {
	info := MustGetChainInfo(ctx.State())
	ret := dict.New()
	ret.Set(root.VarChainID, codec.EncodeChainID(info.ChainID))
	ret.Set(root.VarChainOwnerID, codec.EncodeAgentID(&info.ChainOwnerID))
	ret.Set(root.VarDescription, codec.EncodeString(info.Description))
	ret.Set(root.VarFeeColor, codec.EncodeColor(info.FeeColor))
	ret.Set(root.VarDefaultOwnerFee, codec.EncodeInt64(info.DefaultOwnerFee))
	ret.Set(root.VarDefaultValidatorFee, codec.EncodeInt64(info.DefaultValidatorFee))
	ret.Set(root.VarMaxBlobSize, codec.EncodeUint32(info.MaxBlobSize))
	ret.Set(root.VarMaxEventSize, codec.EncodeUint16(info.MaxEventSize))
	ret.Set(root.VarMaxEventsPerReq, codec.EncodeUint16(info.MaxEventsPerReq))

	src := collections.NewMapReadOnly(ctx.State(), root.VarContractRegistry)
	dst := collections.NewMap(ret, root.VarContractRegistry)
	src.MustIterate(func(elemKey []byte, value []byte) bool {
		dst.MustSetAt(elemKey, value)
		return true
	})
	return ret, nil
}

// delegateChainOwnership stores next possible (delegated) chain owner to another agentID
// checks authorisation by the current owner
// Two step process allow/change is in order to avoid mistakes
func delegateChainOwnership(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.delegateChainOwnership.begin")
	a := assert.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.delegateChainOwnership: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	newOwnerID := params.MustGetAgentID(root.ParamChainOwner)
	ctx.State().Set(root.VarChainOwnerIDDelegated, codec.EncodeAgentID(newOwnerID))
	ctx.Log().Debugf("root.delegateChainOwnership.success: chain ownership delegated to %s", newOwnerID.String())
	return nil, nil
}

// claimChainOwnership changes the chain owner to the delegated agentID (if any)
// Checks authorisation if the caller is the one to which the ownership is delegated
// Note that ownership is only changed by the successful call to  claimChainOwnership
func claimChainOwnership(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.delegateChainOwnership.begin")
	state := ctx.State()
	a := assert.NewAssert(ctx.Log())

	stateDecoder := kvdecoder.New(state, ctx.Log())
	currentOwner := stateDecoder.MustGetAgentID(root.VarChainOwnerID)
	nextOwner := stateDecoder.MustGetAgentID(root.VarChainOwnerIDDelegated, *currentOwner)

	a.Require(!nextOwner.Equals(currentOwner), "root.claimChainOwnership: not delegated to another chain owner")
	a.Require(nextOwner.Equals(ctx.Caller()), "root.claimChainOwnership: not authorized")

	state.Set(root.VarChainOwnerID, codec.EncodeAgentID(nextOwner))
	state.Del(root.VarChainOwnerIDDelegated)
	ctx.Log().Debugf("root.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(), nextOwner.String())
	return nil, nil
}

// getFeeInfo returns fee information for the contract.
// Input:
// - ParamHname iscp.Hname contract id
// Output:
// - ParamFeeColor ledgerstate.Color color of tokens accepted for fees
// - ParamValidatorFee int64 minimum fee for contract
// Note: return default chain values if contract doesn't exist
func getFeeInfo(ctx iscp.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	hname, err := params.GetHname(root.ParamHname)
	if err != nil {
		return nil, err
	}
	feeColor, ownerFee, validatorFee := GetFeeInfo(ctx, hname)
	ret := dict.New()
	ret.Set(root.VarFeeColor, codec.EncodeColor(feeColor))
	ret.Set(root.VarOwnerFee, codec.EncodeUint64(ownerFee))
	ret.Set(root.VarValidatorFee, codec.EncodeUint64(validatorFee))
	return ret, nil
}

// setContractFee sets fee for the particular smart contract
// Input:
// - ParamHname iscp.Hname smart contract ID
// - ParamOwnerFee int64 non-negative value of the owner fee. May be skipped, then it is not set
// - ParamValidatorFee int64 non-negative value of the contract fee. May be skipped, then it is not set
func setContractFee(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.setContractFee: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())

	hname := params.MustGetHname(root.ParamHname)
	rec, found := FindContract(ctx.State(), hname)
	a.Require(found, "contract not found")

	ownerFee := params.MustGetUint64(root.ParamOwnerFee, 0)
	ownerFeeSet := ownerFee > 0
	validatorFee := params.MustGetUint64(root.ParamValidatorFee, 0)
	validatorFeeSet := validatorFee > 0

	a.Require(ownerFeeSet || validatorFeeSet, "root.setContractFee: wrong parameters")
	if ownerFeeSet {
		rec.OwnerFee = ownerFee
	}
	if validatorFeeSet {
		rec.ValidatorFee = validatorFee
	}
	collections.NewMap(ctx.State(), root.VarContractRegistry).MustSetAt(hname.Bytes(), rec.Bytes())
	return nil, nil
}

// grantDeployPermission grants permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func grantDeployPermission(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.grantDeployPermissions: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	deployer := params.MustGetAgentID(root.ParamDeployer)

	collections.NewMap(ctx.State(), root.VarDeployPermissions).MustSetAt(deployer.Bytes(), []byte{0xFF})
	ctx.Event(fmt.Sprintf("[grant deploy permission] to agentID: %s", deployer))
	return nil, nil
}

// revokeDeployPermission revokes permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func revokeDeployPermission(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.revokeDeployPermissions: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	deployer := params.MustGetAgentID(root.ParamDeployer)

	collections.NewMap(ctx.State(), root.VarDeployPermissions).MustDelAt(deployer.Bytes())
	ctx.Event(fmt.Sprintf("[revoke deploy permission] from agentID: %s", deployer))
	return nil, nil
}

// setChainConfig sets the configuration parameters of the chain
// Input (all optional):
// - ParamMaxBlobSize         - uint32 maximum size of a blob to be saved in the blob contract.
// - ParamMaxEventSize        - uint16 maximum size of a single event.
// - ParamMaxEventsPerRequest - uint16 maximum number of events per request.
// - ParamOwnerFee            - int64 non-negative value of the owner fee.
// - ParamValidatorFee        - int64 non-negative value of the contract fee.
func setChainConfig(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.setContractFee: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())

	// max blob size
	maxBlobSize := params.MustGetUint32(root.ParamMaxBlobSize, 0)
	if maxBlobSize > 0 {
		ctx.State().Set(root.VarMaxBlobSize, codec.Encode(maxBlobSize))
		ctx.Event(fmt.Sprintf("[updated chain config] max blob size: %d", maxBlobSize))
	}

	// max event size
	maxEventSize := params.MustGetUint16(root.ParamMaxEventSize, 0)
	if maxEventSize > 0 {
		if maxEventSize < root.MinEventSize {
			// don't allow to set less than MinEventSize to prevent chain owner from bricking the chain
			maxEventSize = root.MinEventSize
		}
		ctx.State().Set(root.VarMaxEventSize, codec.Encode(maxEventSize))
		ctx.Event(fmt.Sprintf("[updated chain config] max event size: %d", maxEventSize))
	}

	// max events per request
	maxEventsPerReq := params.MustGetUint16(root.ParamMaxEventsPerRequest, 0)
	if maxEventsPerReq > 0 {
		if maxEventsPerReq < root.MinEventsPerRequest {
			maxEventsPerReq = root.MinEventsPerRequest
		}
		ctx.State().Set(root.VarMaxEventsPerReq, codec.Encode(maxEventsPerReq))
		ctx.Event(fmt.Sprintf("[updated chain config] max eventsPerRequest: %d", maxEventsPerReq))
	}

	// default owner fee
	ownerFee := params.MustGetInt64(root.ParamOwnerFee, -1)
	if ownerFee >= 0 {
		ctx.State().Set(root.VarDefaultOwnerFee, codec.EncodeInt64(ownerFee))
		ctx.Event(fmt.Sprintf("[updated chain config] default owner fee: %d", ownerFee))
	}

	// default validator fee
	validatorFee := params.MustGetInt64(root.ParamValidatorFee, -1)
	if validatorFee >= 0 {
		ctx.State().Set(root.VarDefaultValidatorFee, codec.EncodeInt64(validatorFee))
		ctx.Event(fmt.Sprintf("[updated chain config] default validator fee: %d", validatorFee))
	}
	return nil, nil
}

func getMaxBlobSize(ctx iscp.SandboxView) (dict.Dict, error) {
	maxBlobSize, err := ctx.State().Get(root.VarMaxBlobSize)
	if err != nil {
		ctx.Log().Panicf("error getting max blob size, %v", err)
	}
	ret := dict.New()
	ret.Set(root.ParamMaxBlobSize, maxBlobSize)
	return ret, nil
}
