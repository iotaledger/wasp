// factory implement processor which is always present at the index 0
// it initializes and operates contract registry: creates contracts and provides search
package root

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize is a handler for the "init" request. This is the first call to the chain
// if it fails, chain is not initialized
// It stores chain ID in the state and creates record for root contract in the contract registry
func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("root.initialize.begin")
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
		ctx.Panic(fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %v", ParamChainID, err))
	}
	// -- description
	chainDescription, ok, err := codec.DecodeString(params.MustGet(ParamDescription))
	if err != nil {
		ctx.Panic(fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamDescription, err))
	}
	if !ok {
		chainDescription = "M/A"
	}
	contractRegistry := datatypes.NewMustMap(state, VarContractRegistry)
	if contractRegistry.Len() != 0 {
		ctx.Panic(fmt.Errorf("root.initialize.fail: registry not empty"))
	}
	// record for root
	contractRegistry.SetAt(Interface.Hname().Bytes(), EncodeContractRecord(&RootContractRecord))
	// deploy blob
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: blob.Interface.ProgramHash,
		Description: blob.Interface.Description,
		Name:        blob.Interface.Name,
		Creator:     ctx.Caller(),
	}, nil)
	if err != nil {
		ctx.Panic(fmt.Errorf("root.init.fail: %v", err))
	}
	// deploy accountsc
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: accountsc.Interface.ProgramHash,
		Description: accountsc.Interface.Description,
		Name:        accountsc.Interface.Name,
		Creator:     ctx.Caller(),
	}, nil)
	if err != nil {
		ctx.Panic(fmt.Errorf("root.init.fail: %v", err))
	}

	state.Set(VarStateInitialized, []byte{0xFF})
	state.Set(VarChainID, codec.EncodeChainID(chainID))
	state.Set(VarChainOwnerID, codec.EncodeAgentID(ctx.Caller())) // chain owner is whoever sends init request
	state.Set(VarDescription, codec.EncodeString(chainDescription))

	ctx.Eventf("root.initialize.deployed: '%s', hname = %s", Interface.Name, Interface.Hname().String())
	ctx.Eventf("root.initialize.deployed: '%s', hname = %s", blob.Interface.Name, blob.Interface.Hname().String())
	ctx.Eventf("root.initialize.deployed: '%s', hname = %s", accountsc.Interface.Name, accountsc.Interface.Hname().String())
	ctx.Eventf("root.initialize.success")
	return nil, nil
}

func deployContract(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("root.deployContract.begin")
	params := ctx.Params()

	proghash, ok, err := codec.DecodeHashValue(params.MustGet(ParamProgramHash))
	if err != nil {
		ctx.Eventf("root.deployContract.wrong.param %s: %v", ParamProgramHash, err)
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("root.deployContract.error: ProgramHash undefined")
	}
	description, ok, err := codec.DecodeString(params.MustGet(ParamDescription))
	if err != nil {
		ctx.Eventf("root.deployContract.wrong.param %s: %v", ParamDescription, err)
		return nil, err
	}
	if !ok {
		description = "N/A"
	}
	name, ok, err := codec.DecodeString(params.MustGet(ParamName))
	if err != nil {
		ctx.Eventf("root.deployContract.wrong.param %s: %v", ParamName, err)
		return nil, err
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
	// only loads VM
	err = ctx.CreateContract(*proghash, "", "", nil)
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
	ctx.Eventf("root.deployContract.success. Deployed contract '%s', hname = %s", name, coretypes.Hn(name).String())
	return nil, nil
}

// findContract is a view
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

// changeChainOwner changes the chain owner to another agentID
// checks authorisation
func changeChainOwner(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("root.allowChangeChainOwner.begin")
	state := ctx.State()
	currentOwner, _, _ := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	nextOwner, ok, err := codec.DecodeAgentID(state.MustGet(VarChainOwnerIDNext))
	if err != nil || !ok {
		return nil, fmt.Errorf("root.changeChainOwner: unknown next owner ID")
	}
	if nextOwner == currentOwner {
		// no need to change
		return nil, nil
	}
	if nextOwner != ctx.Caller() {
		// can be changed only  by the caller if it is equal to the nextOwner
		ctx.Eventf("root.changeChainOwner.fail: not authorized")
		return nil, fmt.Errorf("root.allowChangeChainOwner: not authorized")
	}
	state.Set(VarChainOwnerID, codec.EncodeAgentID(nextOwner))
	ctx.Eventf("root.chainChainOwner.success: chain owner changed: %s --> %s",
		currentOwner.String(), nextOwner.String())
	return nil, nil
}

// allowChangeChainOwner stores next possible chain owner to another agentID
// checks authorisation
// two step process allow/change is in order to avoid mistakes
func allowChangeChainOwner(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("root.allowChangeChainOwner.begin")
	state := ctx.State()
	currentOwner, _, _ := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if currentOwner != ctx.Caller() {
		ctx.Eventf("root.allowChangeChainOwner.fail: not authorized")
		return nil, fmt.Errorf("root.allowChangeChainOwner: not authorized")
	}
	newOwnerID, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamChainOwner))
	if err != nil {
		return nil, fmt.Errorf("root.allowChangeChainOwner: wrong parameter: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("root.allowChangeChainOwner.fail: wrong parameter")
	}
	state.Set(VarChainOwnerIDNext, codec.EncodeAgentID(newOwnerID))
	ctx.Eventf("root.allowChangeChainOwner.success: next owner stored: current %s --> next %s",
		currentOwner.String(), newOwnerID.String())
	return nil, nil
}

func getInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {
	d := dict.New()

	chainID, _, _ := codec.DecodeChainID(ctx.State().MustGet(VarChainID))
	d.Set(VarChainID, codec.EncodeChainID(chainID))

	chainOwner, _, _ := codec.DecodeAgentID(ctx.State().MustGet(VarChainOwnerID))
	d.Set(VarChainOwnerID, codec.EncodeAgentID(chainOwner))

	cr := datatypes.NewMustMap(ctx.State(), VarContractRegistry)
	cr2 := datatypes.NewMustMap(d, VarContractRegistry)
	cr.Iterate(func(elemKey []byte, value []byte) bool {
		cr2.SetAt(elemKey, value)
		return true
	})

	return d, nil
}

//------------------------------ utility function
func storeAndInitContract(ctx vmtypes.Sandbox, rec *ContractRecord, initParams dict.Dict) error {
	hname := coretypes.Hn(rec.Name)
	contractRegistry := datatypes.NewMustMap(ctx.State(), VarContractRegistry)
	if contractRegistry.HasAt(hname.Bytes()) {
		return fmt.Errorf("contract '%s'/%s already exist", rec.Name, hname.String())
	}
	contractRegistry.SetAt(hname.Bytes(), EncodeContractRecord(rec))
	_, err := ctx.Call(coretypes.Hn(rec.Name), coretypes.EntryPointInit, initParams, nil)
	if err != nil {
		// call to 'init' failed: delete record
		contractRegistry.DelAt(hname.Bytes())
		err = fmt.Errorf("contract '%s'/%s: calling 'init': %v", rec.Name, hname.String(), err)
	}
	return err
}
