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

// initialize is a handler for the "init" request
// It stores chain ID in the state and creates record for root contract in the contract registry
func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	params := ctx.Params()
	ctx.Eventf("root.initialize.begin")
	state := ctx.State()
	if state.MustGet(VarStateInitialized) != nil {
		// can't be initialized twice
		return nil, fmt.Errorf("root.initialize.fail: already_initialized")
	}
	// retrieving init parameters
	// -- chain ID
	chainID, ok, err := codec.DecodeChainID(params.MustGet(ParamChainID))
	if !ok || err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamChainID, err.Error())
	}
	// -- description
	chainDescription, ok, err := codec.DecodeString(params.MustGet(ParamDescription))
	if err != nil {
		return nil, fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", ParamDescription, err.Error())
	}
	if !ok {
		chainDescription = "M/A"
	}
	sender := ctx.Caller()

	contractRegistry := datatypes.NewMustMap(state, VarContractRegistry)
	if contractRegistry.Len() != 0 {
		return nil, fmt.Errorf("root.initialize.fail: registry_not_empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.Set(VarChainID, codec.EncodeChainID(chainID))
	state.Set(VarChainOwnerID, codec.EncodeAgentID(sender)) // chain owner is whoever sends init request
	state.Set(VarDescription, codec.EncodeString(chainDescription))

	// record for root
	contractRegistry.SetAt(Interface.Hname().Bytes(), EncodeContractRecord(&RootContractRecord))
	// deploy blob
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: blob.Interface.ProgramHash,
		Description: blob.Interface.Description,
		Name:        blob.Interface.Name,
		Originator:  ctx.Caller(),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("root.init.fail: %v", err)
	}
	ctx.Eventf("root.initialize: deployed '%s', hname = %s", blob.Interface.Name, blob.Interface.Hname().String())

	// deploy accountsc
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: accountsc.Interface.ProgramHash,
		Description: accountsc.Interface.Description,
		Name:        accountsc.Interface.Name,
		Originator:  ctx.Caller(),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("root.init.fail: %v", err)
	}
	ctx.Eventf("root.initialize: deployed '%s', hname = %s", accountsc.Interface.Name, accountsc.Interface.Hname().String())
	ctx.Eventf("root.initialize.success: name: %s, hname: %s", Interface.Name, Interface.Hname().String())
	return nil, nil
}

func deployContract(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("root.deployContract.begin")

	if ctx.State().MustGet(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}
	params := ctx.Params()

	proghash, ok, err := codec.DecodeHashValue(params.MustGet(ParamProgramHash))
	if err != nil {
		ctx.Eventf("root.deployContract.error 1: %v", err)
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("root.deployContract.error: ProgramHash undefined")
	}
	description, ok, err := codec.DecodeString(params.MustGet(ParamDescription))
	if err != nil {
		ctx.Eventf("root.deployContract.error 3: %v", err)
		return nil, err
	}
	if !ok {
		description = "N/A"
	}
	name, ok, err := codec.DecodeString(params.MustGet(ParamName))
	if err != nil {
		ctx.Eventf("root.deployContract.error 4: %v", err)
		return nil, err
	}
	if !ok || name == "" {
		return nil, fmt.Errorf("incorrect contract name")
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
		return nil, err
	}
	// VM loaded successfully. Storing contract in the registry and calling constructor
	err = storeAndInitContract(ctx, &ContractRecord{
		ProgramHash: *proghash,
		Description: description,
		Name:        name,
		Originator:  ctx.Caller(),
	}, initParams)
	if err != nil {
		return nil, err
	}
	ctx.Eventf("root.deployContract.success. Deployed contract hname = %s, name = '%s'", coretypes.Hn(name).String(), name)
	return nil, nil
}

// findContract is a view
func findContract(ctx vmtypes.SandboxView) (dict.Dict, error) {
	if ctx.State().MustGet(VarStateInitialized) == nil {
		return nil, fmt.Errorf("root.initialize.fail: not_initialized")
	}
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

func changeChainOwner(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("root.changeChainOwner.begin")

	state := ctx.State()

	currentOwner, _, _ := codec.DecodeAgentID(state.MustGet(VarChainOwnerID))
	if currentOwner != ctx.Caller() {
		ctx.Eventf("root.changeChainOwner.fail: not authorized")
		return nil, fmt.Errorf("not authorized")
	}
	newOwnerID, ok, err := codec.DecodeAgentID(ctx.Params().MustGet(ParamChainOwner))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("wrong parameter")
	}
	state.Set(VarChainOwnerID, codec.EncodeAgentID(newOwnerID))
	ctx.Eventf("root.changeChainOwner.success: owner changed from %s -> %s", currentOwner.String(), newOwnerID.String())
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
		return fmt.Errorf("contract with hname %s (name = %s) already exist", hname.String(), rec.Name)
	}
	contractRegistry.SetAt(hname.Bytes(), EncodeContractRecord(rec))
	// calling constructor
	if _, err := ctx.Call(coretypes.Hn(rec.Name), coretypes.EntryPointInit, initParams, nil); err != nil {
		ctx.Eventf("root.deployContract.fail. Calling 'init' function for '%s': %v", rec.Name, err)
	}
	return nil
}
