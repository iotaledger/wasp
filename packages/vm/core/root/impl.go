// 'root' a core contract on the chain. It is responsible for:
// - initial setup of the chain during chain deployment
// - maintaining of core parameters of the chain
// - maintaining (setting, delegating) chain owner ID
// - maintaining (granting, revoking) smart contract deployment rights
// - deployment of smart contracts on the chain and maintenance of contract registry
package root

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	assert2 "github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
)

// initialize handles constructor, the "init" request. This is the first call to the chain
// if it fails, chain is not initialized. Does the following:
// - stores chain ID and chain description in the state
// - sets state ownership to the caller
// - creates record in the registry for the 'root' itself
// - deploys other core contracts: 'accounts', 'blob', 'eventlog' by creating records in the registry and calling constructors
// Input:
// - ParamChainID coretypes.ChainID. ID of the chain. Cannot be changed
// - ParamChainColor balance.Color
// - ParamChainAddress address.Address
// - ParamDescription string defaults to "N/A"
// - ParamFeeColor balance.Color fee color code. Defaults to IOTA color. It cannot be changed
func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.initialize.begin")
	state := ctx.State()
	if state.MustGet(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("root.initialize.fail: already initialized")
	}
	// retrieving init parameters
	// -- chain ID
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert2.NewAssert(ctx.Log())

	chainID := params.MustGetChainID(ParamChainID)
	chainColor := params.MustGetColor(ParamChainColor)
	chainAddress := params.MustGetAddress(ParamChainAddress)
	chainDescription := params.MustGetString(ParamDescription, "N/A")
	feeColor := params.MustGetColor(ParamFeeColor, balance.ColorIOTA)
	feeColorSet := feeColor != balance.ColorIOTA

	contractRegistry := collections.NewMap(state, VarContractRegistry)
	a.Require(contractRegistry.MustLen() == 0, "root.initialize.fail: registry not empty")

	rec := NewContractRecord(Interface, coretypes.AgentID{})
	contractRegistry.MustSetAt(Interface.Hname().Bytes(), EncodeContractRecord(&rec))

	// deploy blob
	rec = NewContractRecord(blob.Interface, ctx.Caller())
	err := storeAndInitContract(ctx, &rec, nil)
	a.Require(err == nil, "root.init.fail: %v", err)

	// deploy accounts
	rec = NewContractRecord(accounts.Interface, ctx.Caller())
	err = storeAndInitContract(ctx, &rec, nil)
	a.Require(err == nil, "root.init.fail: %v", err)

	// deploy chainlog
	rec = NewContractRecord(eventlog.Interface, ctx.Caller())
	err = storeAndInitContract(ctx, &rec, nil)
	a.Require(err == nil, "root.init.fail: %v", err)

	state.Set(VarStateInitialized, []byte{0xFF})
	state.Set(VarChainID, codec.EncodeChainID(chainID))
	state.Set(VarChainColor, codec.EncodeColor(chainColor))
	state.Set(VarChainAddress, codec.EncodeAddress(chainAddress))
	state.Set(VarChainOwnerID, codec.EncodeAgentID(ctx.Caller())) // chain owner is whoever sends init request
	state.Set(VarDescription, codec.EncodeString(chainDescription))
	if feeColorSet {
		state.Set(VarFeeColor, codec.EncodeColor(feeColor))
	}
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", Interface.Name, Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", blob.Interface.Name, blob.Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", accounts.Interface.Name, accounts.Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", eventlog.Interface.Name, eventlog.Interface.Hname().String())
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
func deployContract(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.deployContract.begin")
	if !isAuthorizedToDeploy(ctx) {
		return nil, fmt.Errorf("root.deployContract: deploy not permitted for: %s", ctx.Caller())
	}
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert2.NewAssert(ctx.Log())

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
	a.Require(err == nil, "root.deployContract.fail: %v", err)

	// VM loaded successfully. Storing contract in the registry and calling constructor
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: progHash,
		Description: description,
		Name:        name,
		Creator:     ctx.Caller(),
	}, initParams)
	a.Require(err == nil, "root.deployContract.fail: %v", err)

	ctx.Event(fmt.Sprintf("[deploy] name: %s hname: %s, progHash: %s, dscr: '%s'",
		name, coretypes.Hn(name), progHash.String(), description))
	return nil, nil
}

// findContract view finds and returns encoded record of the contract
// Input:
// - ParamHname
// Output:
// - ParamData
func findContract(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	hname, err := params.GetHname(ParamHname)
	if err != nil {
		return nil, err
	}
	rec, err := FindContract(ctx.State(), hname)
	if err != nil {
		return nil, err
	}
	retBin := EncodeContractRecord(rec)
	ret := dict.New()
	ret.Set(ParamData, retBin)
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
func getChainInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	info := MustGetChainInfo(ctx.State())
	ret := dict.New()
	ret.Set(VarChainID, codec.EncodeChainID(info.ChainID))
	ret.Set(VarChainOwnerID, codec.EncodeAgentID(info.ChainOwnerID))
	ret.Set(VarChainColor, codec.EncodeColor(info.ChainColor))
	ret.Set(VarChainAddress, codec.EncodeAddress(info.ChainAddress))
	ret.Set(VarDescription, codec.EncodeString(info.Description))
	ret.Set(VarFeeColor, codec.EncodeColor(info.FeeColor))
	ret.Set(VarDefaultOwnerFee, codec.EncodeInt64(info.DefaultOwnerFee))
	ret.Set(VarDefaultValidatorFee, codec.EncodeInt64(info.DefaultValidatorFee))

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
func delegateChainOwnership(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.delegateChainOwnership.begin")
	a := assert2.NewAssert(ctx.Log())
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
func claimChainOwnership(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.delegateChainOwnership.begin")
	state := ctx.State()
	a := assert2.NewAssert(ctx.Log())

	stateDecoder := kvdecoder.New(state, ctx.Log())
	currentOwner := stateDecoder.MustGetAgentID(VarChainOwnerID)
	nextOwner := stateDecoder.MustGetAgentID(VarChainOwnerIDDelegated, currentOwner)

	a.Require(nextOwner != currentOwner, "root.claimChainOwnership: not delegated to another chain owner")
	a.Require(nextOwner == ctx.Caller(), "root.claimChainOwnership: not authorized")

	state.Set(VarChainOwnerID, codec.EncodeAgentID(nextOwner))
	state.Del(VarChainOwnerIDDelegated)
	ctx.Log().Debugf("root.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(), nextOwner.String())
	return nil, nil
}

// getFeeInfo returns fee information for the contact.
// Input:
// - ParamHname coretypes.Hname contract id
// Output:
// - ParamFeeColor balance.Color color of tokens accepted for fees
// - ParamValidatorFee int64 minimum fee for contract
// Note: return default chain values if contract doesn't exist
func getFeeInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	hname, err := params.GetHname(ParamHname)
	if err != nil {
		return nil, err
	}
	feeColor, ownerFee, validatorFee := GetFeeInfo(ctx.State(), hname)
	ret := dict.New()
	ret.Set(ParamFeeColor, codec.EncodeColor(feeColor))
	ret.Set(ParamOwnerFee, codec.EncodeInt64(ownerFee))
	ret.Set(ParamValidatorFee, codec.EncodeInt64(validatorFee))
	return ret, nil
}

// setDefaultFee sets default fee values for the chain
// Input:
// - ParamOwnerFee int64 non-negative value of the owner fee. May be skipped, then it is not set
// - ParamValidatorFee int64 non-negative value of the contract fee. May be skipped, then it is not set
func setDefaultFee(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert2.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.setDefaultFee: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())

	ownerFee := params.MustGetInt64(ParamOwnerFee, -1)
	ownerFeeSet := ownerFee >= 0
	validatorFee := params.MustGetInt64(ParamValidatorFee, -1)
	validatorFeeSet := validatorFee >= 0

	a.Require(ownerFeeSet || validatorFeeSet, "root.setDefaultFee: wrong parameters")

	if ownerFeeSet {
		if ownerFee > 0 {
			ctx.State().Set(VarDefaultOwnerFee, codec.EncodeInt64(ownerFee))
		} else {
			ctx.State().Del(VarDefaultOwnerFee)
		}
	}
	if validatorFeeSet {
		if validatorFee > 0 {
			ctx.State().Set(VarDefaultValidatorFee, codec.EncodeInt64(validatorFee))
		} else {
			ctx.State().Del(VarDefaultValidatorFee)
		}
	}
	return nil, nil
}

// setContractFee sets fee for the particular smart contract
// Input:
// - ParamHname coretypes.Hname smart contract ID
// - ParamOwnerFee int64 non-negative value of the owner fee. May be skipped, then it is not set
// - ParamValidatorFee int64 non-negative value of the contract fee. May be skipped, then it is not set
func setContractFee(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert2.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.setContractFee: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())

	hname := params.MustGetHname(ParamHname)
	rec, err := FindContract(ctx.State(), hname)
	if err != nil {
		return nil, err
	}

	ownerFee := params.MustGetInt64(ParamOwnerFee, -1)
	ownerFeeSet := ownerFee >= 0
	validatorFee := params.MustGetInt64(ParamValidatorFee, -1)
	validatorFeeSet := validatorFee >= 0

	a.Require(ownerFeeSet || validatorFeeSet, "root.setContractFee: wrong parameters")
	if ownerFeeSet {
		rec.OwnerFee = ownerFee
	}
	if validatorFeeSet {
		rec.ValidatorFee = validatorFee
	}
	collections.NewMap(ctx.State(), VarContractRegistry).MustSetAt(hname.Bytes(), EncodeContractRecord(rec))
	return nil, nil
}

// grantDeployPermission grants permission to deploy contracts
// Input:
//  - ParamDeployer coretypes.AgentID
func grantDeployPermission(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert2.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.grantDeployPermissions: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	deployer := params.MustGetAgentID(ParamDeployer)

	collections.NewMap(ctx.State(), VarDeployPermissions).MustSetAt(deployer.Bytes(), []byte{0xFF})
	ctx.Event(fmt.Sprintf("[grant deploy permission] to agentID: %s", deployer))
	return nil, nil
}

// grantDeployPermission revokes permission to deploy contracts
// Input:
//  - ParamDeployer coretypes.AgentID
func revokeDeployPermission(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert2.NewAssert(ctx.Log())
	a.Require(CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()), "root.revokeDeployPermissions: not authorized")

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	deployer := params.MustGetAgentID(ParamDeployer)

	collections.NewMap(ctx.State(), VarDeployPermissions).MustDelAt(deployer.Bytes())
	ctx.Event(fmt.Sprintf("[revoke deploy permission] from agentID: %s", deployer))
	return nil, nil
}
