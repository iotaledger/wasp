// 'root' a core contract on the chain. It is responsible for:
// - initial setup of the chain during chain deployment
// - maintaining of core parameters of the chain
// - maintaining (setting, delegating) chain owner ID
// - maintaining (granting, revoking) smart contract deployment rights
// - deployment of smart contracts on the chain and maintenance of contract registry
package root

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
)

var Processor = Contract.Processor(initialize,
	FuncClaimChainOwnership.WithHandler(claimChainOwnership),
	FuncDelegateChainOwnership.WithHandler(delegateChainOwnership),
	FuncDeployContract.WithHandler(deployContract),
	FuncGrantDeployPermission.WithHandler(grantDeployPermission),
	FuncRevokeDeployPermission.WithHandler(revokeDeployPermission),
	FuncSetContractFee.WithHandler(setContractFee),
	FuncSetChainConfig.WithHandler(setChainConfig),
	FuncFindContract.WithHandler(findContract),
	FuncGetChainConfig.WithHandler(getChainConfig),
	FuncGetFeeInfo.WithHandler(getFeeInfo),
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

	a.Require(state.MustGet(VarStateInitialized) == nil, "root.initialize.fail: already initialized")
	a.Require(ctx.Caller().Hname() == 0, "root.init.fail: chain deployer can't be another smart contract")

	// retrieving init parameters
	// -- chain ID
	params := kvdecoder.New(ctx.Params(), ctx.Log())

	chainID := params.MustGetChainID(ParamChainID)
	chainDescription := params.MustGetString(ParamDescription, "N/A")
	feeColor := params.MustGetColor(ParamFeeColor, colored.IOTA)
	feeColorSet := feeColor != colored.IOTA

	contractRegistry := collections.NewMap(state, VarContractRegistry)
	a.Require(contractRegistry.MustLen() == 0, "root.initialize.fail: registry not empty")

	mustStoreContract(ctx, _default.Contract, a)
	mustStoreContract(ctx, Contract, a)
	mustStoreAndInitCoreContract(ctx, blob.Contract, a)
	mustStoreAndInitCoreContract(ctx, accounts.Contract, a)
	mustStoreAndInitCoreContract(ctx, blocklog.Contract, a)
	mustStoreAndInitCoreContract(ctx, governance.Contract, a)

	state.Set(VarStateInitialized, []byte{0xFF})
	state.Set(VarChainID, codec.EncodeChainID(*chainID))
	state.Set(VarChainOwnerID, codec.EncodeAgentID(ctx.Caller())) // chain owner is whoever sends init request
	state.Set(VarDescription, codec.EncodeString(chainDescription))

	state.Set(VarMaxBlobSize, codec.Encode(DefaultMaxBlobSize))
	state.Set(VarMaxEventSize, codec.Encode(DefaultMaxEventSize))
	state.Set(VarMaxEventsPerReq, codec.Encode(DefaultMaxEventsPerRequest))

	if feeColorSet {
		state.Set(VarFeeColor, codec.EncodeColor(feeColor))
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

	progHash := params.MustGetHashValue(ParamProgramHash)
	description := params.MustGetString(ParamDescription, "N/A")
	name := params.MustGetString(ParamName)
	a.Require(name != "", "wrong name")

	// pass to init function all params not consumed so far
	initParams := dict.New()
	for key, value := range ctx.Params() {
		if key != ParamProgramHash && key != ParamName && key != ParamDescription {
			initParams.Set(key, value)
		}
	}
	// calls to loads VM from binary to check if it loads successfully
	err := ctx.DeployContract(progHash, "", "", nil)
	a.Require(err == nil, "root.deployContract.fail 1: %v", err)

	// VM loaded successfully. Storing contract in the registry and calling constructor
	mustStoreContractRecord(ctx, &ContractRecord{
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
	hname, err := params.GetHname(ParamHname)
	if err != nil {
		return nil, err
	}
	rec, found := FindContract(ctx.State(), hname)
	ret := dict.New()
	ret.Set(ParamContractRecData, rec.Bytes())
	var foundByte [1]byte
	if found {
		foundByte[0] = 0xFF
	}
	ret.Set(ParamContractFound, foundByte[:])
	return ret, nil
}

// getChainConfig view returns general info about the chain: chain ID, chain owner ID,
// description and the whole contract registry
// Input: none
// Output:
// - VarChainID - ChainID
// - VarChainOwnerID - AgentID
// - VarDescription - string
// - VarContractRegistry: a map of contract registry
func getChainConfig(ctx iscp.SandboxView) (dict.Dict, error) {
	info := MustGetChainConfig(ctx.State())
	ret := dict.New()
	ret.Set(VarChainID, codec.EncodeChainID(info.ChainID))
	ret.Set(VarChainOwnerID, codec.EncodeAgentID(&info.ChainOwnerID))
	ret.Set(VarDescription, codec.EncodeString(info.Description))
	ret.Set(VarFeeColor, codec.EncodeColor(info.FeeColor))
	ret.Set(VarDefaultOwnerFee, codec.EncodeInt64(info.DefaultOwnerFee))
	ret.Set(VarDefaultValidatorFee, codec.EncodeInt64(info.DefaultValidatorFee))
	ret.Set(VarMaxBlobSize, codec.EncodeUint32(info.MaxBlobSize))
	ret.Set(VarMaxEventSize, codec.EncodeUint16(info.MaxEventSize))
	ret.Set(VarMaxEventsPerReq, codec.EncodeUint16(info.MaxEventsPerReq))

	src := collections.NewMapReadOnly(ctx.State(), VarContractRegistry)
	dst := collections.NewMap(ret, VarContractRegistry)
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
	newOwnerID := params.MustGetAgentID(ParamChainOwner)
	ctx.State().Set(VarChainOwnerIDDelegated, codec.EncodeAgentID(newOwnerID))
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
	currentOwner := stateDecoder.MustGetAgentID(VarChainOwnerID)
	nextOwner := stateDecoder.MustGetAgentID(VarChainOwnerIDDelegated, *currentOwner)

	a.Require(!nextOwner.Equals(currentOwner), "root.claimChainOwnership: not delegated to another chain owner")
	a.Require(nextOwner.Equals(ctx.Caller()), "root.claimChainOwnership: not authorized")

	state.Set(VarChainOwnerID, codec.EncodeAgentID(nextOwner))
	state.Del(VarChainOwnerIDDelegated)
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
	hname, err := params.GetHname(ParamHname)
	if err != nil {
		return nil, err
	}
	feeColor, ownerFee, validatorFee := GetFeeInfo(ctx, hname)
	ret := dict.New()
	ret.Set(VarFeeColor, codec.EncodeColor(feeColor))
	ret.Set(VarOwnerFee, codec.EncodeUint64(ownerFee))
	ret.Set(VarValidatorFee, codec.EncodeUint64(validatorFee))
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

	hname := params.MustGetHname(ParamHname)
	rec, found := FindContract(ctx.State(), hname)
	a.Require(found, "contract not found")

	ownerFee := params.MustGetUint64(ParamOwnerFee, 0)
	ownerFeeSet := ownerFee > 0
	validatorFee := params.MustGetUint64(ParamValidatorFee, 0)
	validatorFeeSet := validatorFee > 0

	a.Require(ownerFeeSet || validatorFeeSet, "root.setContractFee: wrong parameters")
	if ownerFeeSet {
		rec.OwnerFee = ownerFee
	}
	if validatorFeeSet {
		rec.ValidatorFee = validatorFee
	}
	collections.NewMap(ctx.State(), VarContractRegistry).MustSetAt(hname.Bytes(), rec.Bytes())
	return nil, nil
}

// grantDeployPermission grants permission to deploy contracts
// Input:
//  - ParamDeployer iscp.AgentID
func grantDeployPermission(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.grantDeployPermissions: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	deployer := params.MustGetAgentID(ParamDeployer)

	collections.NewMap(ctx.State(), VarDeployPermissions).MustSetAt(deployer.Bytes(), []byte{0xFF})
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
	deployer := params.MustGetAgentID(ParamDeployer)

	collections.NewMap(ctx.State(), VarDeployPermissions).MustDelAt(deployer.Bytes())
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
	maxBlobSize := params.MustGetUint32(ParamMaxBlobSize, 0)
	if maxBlobSize > 0 {
		ctx.State().Set(VarMaxBlobSize, codec.Encode(maxBlobSize))
		ctx.Event(fmt.Sprintf("[updated chain config] max blob size: %d", maxBlobSize))
	}

	// max event size
	maxEventSize := params.MustGetUint16(ParamMaxEventSize, 0)
	if maxEventSize > 0 {
		if maxEventSize < MinEventSize {
			// don't allow to set less than MinEventSize to prevent chain owner from bricking the chain
			maxEventSize = MinEventSize
		}
		ctx.State().Set(VarMaxEventSize, codec.Encode(maxEventSize))
		ctx.Event(fmt.Sprintf("[updated chain config] max event size: %d", maxEventSize))
	}

	// max events per request
	maxEventsPerReq := params.MustGetUint16(ParamMaxEventsPerRequest, 0)
	if maxEventsPerReq > 0 {
		if maxEventsPerReq < MinEventsPerRequest {
			maxEventsPerReq = MinEventsPerRequest
		}
		ctx.State().Set(VarMaxEventsPerReq, codec.Encode(maxEventsPerReq))
		ctx.Event(fmt.Sprintf("[updated chain config] max eventsPerRequest: %d", maxEventsPerReq))
	}

	// default owner fee
	ownerFee := params.MustGetInt64(ParamOwnerFee, -1)
	if ownerFee >= 0 {
		ctx.State().Set(VarDefaultOwnerFee, codec.EncodeInt64(ownerFee))
		ctx.Event(fmt.Sprintf("[updated chain config] default owner fee: %d", ownerFee))
	}

	// default validator fee
	validatorFee := params.MustGetInt64(ParamValidatorFee, -1)
	if validatorFee >= 0 {
		ctx.State().Set(VarDefaultValidatorFee, codec.EncodeInt64(validatorFee))
		ctx.Event(fmt.Sprintf("[updated chain config] default validator fee: %d", validatorFee))
	}

	return nil, nil

	// TODO set default values

	// then enforce the values and enforce them where needed
}
