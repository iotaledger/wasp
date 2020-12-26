// 'root' a core contract on the chain. It is responsible for:
// - initial setup of the chain during chain deployment
// - maintaining of core parameters of the chain
// - maintaining (setting, delegating) chain owner ID
// - deployment of smart contracts on the chain and maintenance of contract registry
package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize handles constructor, the "init" request. This is the first call to the chain
// if it fails, chain is not initialized. Does the following:
// - stores chain ID and chain description in the state
// - sets state ownership to the caller
// - creates record in the registry for the 'root' itself
// - deploys other core contracts: 'accountsc', 'blob' by creating records in the registry and calling constructors
// Input:
// - ParamChainID coretypes.ChainID. ID of the chain. Cannot be changed
// - ParamDescription string defaults to "N/A"
// - ParamFeeColor balance.Color fee color code. Defaults to IOTA color. It cannot be changed
// - ParamOwnerFee int64 globally set default fee value. Defaults to 0
func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.initialize.begin")
	params := ctx.Params()
	state := ctx.State()
	if state.MustGet(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("root.initialize.fail: already initialized")
	}
	// retrieving init parameters
	// -- chain ID
	chainID, ok, err := codec.DecodeChainID(params.MustGet(ParamChainID))
	if !ok || err != nil {
		ctx.Log().Panicf("root.initialize.fail: can't read expected request argument '%s': %v", ParamChainID, err)
	}
	chainColor, ok, err := codec.DecodeColor(params.MustGet(ParamChainColor))
	if !ok || err != nil {
		ctx.Log().Panicf("root.initialize.fail: can't read expected request argument '%s': %v", ParamChainColor, err)
	}
	chainAddress, ok, err := codec.DecodeAddress(params.MustGet(ParamChainAddress))
	if !ok || err != nil {
		ctx.Log().Panicf("root.initialize.fail: can't read expected request argument '%s': %v", ParamChainAddress, err)
	}
	// -- description
	chainDescription, ok, err := codec.DecodeString(params.MustGet(ParamDescription))
	if err != nil {
		ctx.Log().Panicf("root.initialize.fail: can't read expected request argument '%s': %s", ParamDescription, err)
	}
	if !ok {
		chainDescription = "N/A"
	}
	feeColor, feeColorSet, err := codec.DecodeColor(params.MustGet(ParamFeeColor))
	if err != nil {
		ctx.Log().Panicf("root.initialize.fail: can't read expected request argument '%s': %s", ParamFeeColor, err)
	}
	defaultFee, defaultFeeSet, err := codec.DecodeInt64(params.MustGet(ParamOwnerFee))
	if err != nil {
		ctx.Log().Panicf("root.initialize.fail: can't read expected request argument '%s': %s", ParamOwnerFee, err)
	}
	contractRegistry := datatypes.NewMustMap(state, VarContractRegistry)
	if contractRegistry.Len() != 0 {
		ctx.Log().Panicf("root.initialize.fail: registry not empty")
	}
	// record for root
	rec := NewContractRecord(Interface, coretypes.AgentID{})
	contractRegistry.SetAt(Interface.Hname().Bytes(), EncodeContractRecord(&rec))
	// deploy blob
	rec = NewContractRecord(blob.Interface, ctx.Caller())
	err = storeAndInitContract(ctx, &rec, nil)
	if err != nil {
		ctx.Log().Panicf("root.init.fail: %v", err)
	}
	// deploy accountsc
	rec = NewContractRecord(accounts.Interface, ctx.Caller())
	err = storeAndInitContract(ctx, &rec, nil)
	if err != nil {
		ctx.Log().Panicf("root.init.fail: %v", err)
	}
	// deploy chainlog
	rec = NewContractRecord(chainlog.Interface, ctx.Caller())
	err = storeAndInitContract(ctx, &rec, nil)
	if err != nil {
		ctx.Log().Panicf("root.init.fail: %v", err)
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.Set(VarChainID, codec.EncodeChainID(chainID))
	state.Set(VarChainColor, codec.EncodeColor(chainColor))
	state.Set(VarChainAddress, codec.EncodeAddress(chainAddress))
	state.Set(VarChainOwnerID, codec.EncodeAgentID(ctx.Caller())) // chain owner is whoever sends init request
	state.Set(VarDescription, codec.EncodeString(chainDescription))
	if feeColorSet {
		state.Set(VarFeeColor, codec.EncodeColor(feeColor))
	}
	if defaultFeeSet {
		if defaultFee < 0 {
			defaultFee = 0
		}
		if defaultFee > 0 {
			state.Set(VarDefaultOwnerFee, codec.EncodeInt64(defaultFee))
		}
	}
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", Interface.Name, Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", blob.Interface.Name, blob.Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", accounts.Interface.Name, accounts.Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.deployed: '%s', hname = %s", chainlog.Interface.Name, chainlog.Interface.Hname().String())
	ctx.Log().Debugf("root.initialize.success")
	return nil, nil
}

// deployContract deploys contract and calls its 'init' constructor.
// If call to the constructor returns an error or an other error occurs,
// removes smart contract form the registry as if it was never attempted to deploy
// Inputs:
// - ParamName string, the unique name of the contract in the chain. Latter used as a hname
// - ParamProgramHash HashValue is a hash of the blob which represents program binary in the 'blob' contract.
//     In case of hardcoded examples its an arbitrary unique hash set in the global call examples.AddProcessor
// - ParamDescription string is an arbitrary string. Defaults to "N/A"
func deployContract(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.deployContract.begin")
	if !isAuthorizedToDeploy(ctx) {
		return nil, fmt.Errorf("root.deployContract.deployContract: not permitted")
	}
	params := ctx.Params()
	proghash, ok, err := codec.DecodeHashValue(params.MustGet(ParamProgramHash))
	if err != nil {
		return nil, fmt.Errorf("root.deployContract.wrong.param %s: %v", ParamProgramHash, err)
	}
	if !ok {
		return nil, fmt.Errorf("root.deployContract.error: ProgramHash undefined")
	}
	description, ok, err := codec.DecodeString(params.MustGet(ParamDescription))
	if err != nil {
		return nil, fmt.Errorf("root.deployContract.wrong.param %s: %v", ParamDescription, err)
	}
	if !ok {
		description = "N/A"
	}
	name, ok, err := codec.DecodeString(params.MustGet(ParamName))
	if err != nil {
		return nil, fmt.Errorf("root.deployContract.wrong.param %s: %v", ParamName, err)
	}
	if !ok || name == "" {
		return nil, fmt.Errorf("root.deployContract.fail: wrong contract name")
	}
	// pass to init function all params not consumed so far
	initParams := dict.New()
	for key, value := range params {
		if key != ParamProgramHash && key != ParamName && key != ParamDescription {
			initParams.Set(key, value)
		}
	}
	// calls to loads VM from binary to check if it loads successfully
	err = ctx.DeployContract(*proghash, "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("root.deployContract.fail: %v", err)
	}
	// VM loaded successfully. Storing contract in the registry and calling constructor
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: *proghash,
		Description: description,
		Name:        name,
		Creator:     ctx.Caller(),
	}, initParams)
	if err != nil {
		return nil, fmt.Errorf("root.deployContract.fail: %v", err)
	}

	ctx.Event(fmt.Sprintf("[deploy] name: %s hname: %s, progHash: %s, dscr: '%s'",
		name, coretypes.Hn(name), proghash.String(), description))
	return nil, nil
}

// findContract view finds and returns encoded record of the contract
// Input:
// - ParamHname
// Output:
// - ParamData
func findContract(ctx vmtypes.SandboxView) (dict.Dict, error) {
	params := ctx.Params()
	hname, ok, err := codec.DecodeHname(params.MustGet(ParamHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'hname' undefined")
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
func getChainInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {
	info, err := GetChainInfo(ctx.State())
	if err != nil {
		return nil, err
	}
	ret := dict.New()
	ret.Set(VarChainID, codec.EncodeChainID(info.ChainID))
	ret.Set(VarChainOwnerID, codec.EncodeAgentID(info.ChainOwnerID))
	ret.Set(VarChainColor, codec.EncodeColor(info.ChainColor))
	ret.Set(VarChainAddress, codec.EncodeAddress(info.ChainAddress))
	ret.Set(VarDescription, codec.EncodeString(info.Description))
	ret.Set(VarFeeColor, codec.EncodeColor(info.FeeColor))
	ret.Set(VarDefaultOwnerFee, codec.EncodeInt64(info.DefaultOwnerFee))
	ret.Set(VarDefaultValidatorFee, codec.EncodeInt64(info.DefaultValidatorFee))

	src := datatypes.NewMustMap(ctx.State(), VarContractRegistry)
	dst := datatypes.NewMustMap(ret, VarContractRegistry)
	src.Iterate(func(elemKey []byte, value []byte) bool {
		dst.SetAt(elemKey, value)
		return true
	})
	return ret, nil
}

// delegateChainOwnership stores next possible (delegated) chain owner to another agentID
// checks authorisation by the current owner
// Two step process allow/change is in order to avoid mistakes
func delegateChainOwnership(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.delegateChainOwnership.begin")
	if !CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()) {
		return nil, fmt.Errorf("root.delegateChainOwnership: not authorized")
	}
	newOwnerID, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamChainOwner))
	if err != nil {
		return nil, fmt.Errorf("root.delegateChainOwnership: wrong parameter: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("root.delegateChainOwnership.fail: wrong parameter")
	}
	ctx.State().Set(VarChainOwnerIDDelegated, codec.EncodeAgentID(newOwnerID))
	ctx.Log().Debugf("root.delegateChainOwnership.success: chain ownership delegated to %s", newOwnerID.String())
	return nil, nil
}

// claimChainOwnership changes the chain owner to the delegated agentID (if any)
// Checks authorisation if the caller is the one to which the ownership is delegated
// Note that ownership is only changed by the successful call to  claimChainOwnership
func claimChainOwnership(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("root.delegateChainOwnership.begin")
	state := ctx.State()
	currentOwner, _, _ := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	nextOwner, ok, err := codec.DecodeAgentID(state.MustGet(VarChainOwnerIDDelegated))
	if err != nil || !ok {
		return nil, fmt.Errorf("root.claimChainOwnership: not delegated to another chain owner")
	}
	if nextOwner == currentOwner {
		// no need to change
		return nil, nil
	}
	if nextOwner != ctx.Caller() {
		// can be changed only by the caller to which ownership is delegated
		return nil, fmt.Errorf("root.delegateChainOwnership: not authorized")
	}
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
func getFeeInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {
	params := ctx.Params()
	hname, ok, err := codec.DecodeHname(params.MustGet(ParamHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'hname' undefined")
	}
	feeColor, ownerFee, validatorFee, err := GetFeeInfo(ctx.State(), hname)
	if err != nil {
		return nil, fmt.Errorf("GetFeeInfo: %v", err)

	}
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
func setDefaultFee(ctx vmtypes.Sandbox) (dict.Dict, error) {
	if !CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()) {
		return nil, fmt.Errorf("root.setDefaultFee: not authorized")
	}
	ownerFee, ownerFeeOk, err := codec.DecodeInt64(ctx.Params().MustGet(ParamOwnerFee))
	if err != nil {
		return nil, err
	}
	if ownerFeeOk && ownerFee < 0 {
		return nil, fmt.Errorf("parameter 'owner fee' is invalid")
	}
	validatorFee, validatorFeeOk, err := codec.DecodeInt64(ctx.Params().MustGet(ParamValidatorFee))
	if err != nil {
		return nil, err
	}
	if validatorFeeOk && validatorFee < 0 {
		return nil, fmt.Errorf("parameter 'validator fee' is invalid")
	}
	if !ownerFeeOk && !validatorFeeOk {
		return nil, fmt.Errorf("missing parameters")
	}
	if ownerFeeOk {
		if ownerFee > 0 {
			ctx.State().Set(VarDefaultOwnerFee, codec.EncodeInt64(ownerFee))
		} else {
			ctx.State().Del(VarDefaultOwnerFee)
		}
	}
	if validatorFeeOk {
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
func setContractFee(ctx vmtypes.Sandbox) (dict.Dict, error) {
	if !CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()) {
		return nil, fmt.Errorf("root.setContractFee: not authorized")
	}
	hname, ok, err := codec.DecodeHname(ctx.Params().MustGet(ParamHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'hname' undefined")
	}
	rec, err := FindContract(ctx.State(), hname)
	if err != nil {
		// contract not found
		return nil, err
	}
	ownerFee, ownerFeeOk, err := codec.DecodeInt64(ctx.Params().MustGet(ParamOwnerFee))
	if err != nil {
		return nil, err
	}
	if ownerFeeOk && ownerFee < 0 {
		return nil, fmt.Errorf("parameter 'owner fee' is invalid")
	}
	rec.OwnerFee = ownerFee
	validatorFee, validatorFeeOk, err := codec.DecodeInt64(ctx.Params().MustGet(ParamValidatorFee))
	if err != nil {
		return nil, err
	}
	if validatorFeeOk && validatorFee < 0 {
		return nil, fmt.Errorf("parameter 'validator fee' is invalid")
	}
	if !ownerFeeOk && !validatorFeeOk {
		return nil, fmt.Errorf("missing parameters")
	}
	rec.ValidatorFee = validatorFee
	datatypes.NewMustMap(ctx.State(), VarContractRegistry).SetAt(hname.Bytes(), EncodeContractRecord(rec))
	return nil, nil
}

// grantDeploy grants permission to deploy contracts
// Input:
//  - ParamDeployer coretypes.AgentID
func grantDeploy(ctx vmtypes.Sandbox) (dict.Dict, error) {
	if !CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()) {
		return nil, fmt.Errorf("root.grantDeployer: not authorized")
	}
	deployer, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamDeployer))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'deployer' undefined")
	}

	datatypes.NewMustMap(ctx.State(), VarDeployAuthorisations).SetAt(deployer.Bytes(), []byte{0xFF})
	return nil, nil
}

// grantDeploy revokes permission to deploy contracts
// Input:
//  - ParamDeployer coretypes.AgentID
func revokeDeploy(ctx vmtypes.Sandbox) (dict.Dict, error) {
	if !CheckAuthorizationByChainOwner(ctx.State(), ctx.Caller()) {
		return nil, fmt.Errorf("root.revokeDeployer: not authorized")
	}
	deployer, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamDeployer))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'deployer' undefined")
	}
	datatypes.NewMustMap(ctx.State(), VarDeployAuthorisations).DelAt(deployer.Bytes())
	return nil, nil
}
