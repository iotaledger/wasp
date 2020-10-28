// factory implements factory processor. This processor is always present on the chain
// and most likely it will always be built in. It functions:
// - initialize state of the chain (store chain id and other parameters)
// - to handle 'ownership' of the chain
// - to provide constructors for deployment of new contracts
package factory

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func Initialize(ctx vmtypes.Sandbox, chainID *coretypes.ChainID) error {
	ctx.Publishf("factory.Initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		return fmt.Errorf("factory.Initialize.fail already_initialized")
	}
	registry := state.GetDictionary(VarContractRegistry)
	nextIndex := coretypes.Uint16(registry.Len())

	if nextIndex != 0 {
		return fmt.Errorf("factory.initialize.fail: internal error, registry not empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	// at index 0 always this contract
	registry.SetAt(nextIndex.Bytes(), chainID[:])
	return nil
}

func initialize(ctx vmtypes.Sandbox, params ...interface{}) {
	ctx.Publishf("factory.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		ctx.Publishf("factory.initialize.fail: already_initialized")
		return
	}
	request := ctx.AccessRequest()
	chainID, ok, err := request.Args().GetChainID(VarChainID)
	if err != nil {
		ctx.Publishf("factory.initialize.fail: can't read request argument '%s': %s", VarChainID, err.Error())
		return
	}
	if !ok {
		ctx.Publishf("factory.initialize.fail: 'chainID' not found")
		return
	}
	if err := Initialize(ctx, chainID); err != nil {
		ctx.Publishf("factory.initialize.fail: %v", err)
	}

	ctx.Publishf("bootup.success")
}

func newContract(ctx vmtypes.Sandbox, params ...interface{}) {
}
