// factory implements factory processor. This processor is always present on the chain
// and most likely it will always be built in. It functions:
// - initialize state of the chain (store chain id and other parameters)
// - to handle 'ownership' of the chain
// - to provide constructors for deployment of new contracts
package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox, params kv.RCodec) error {
	ctx.Publishf("root.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		return fmt.Errorf("root.initialize.fail: already_initialized")
	}
	chainID, ok, err := params.GetChainID(VarChainID)
	if err != nil {
		return fmt.Errorf("root.initialize.fail: can't read expected request argument '%s': %s", VarChainID, err.Error())
	}
	if !ok {
		return fmt.Errorf("root.initialize.fail: 'chainID' not found")
	}
	registry := state.GetDictionary(VarContractRegistry)
	nextIndex := coretypes.Uint16(registry.Len())

	if nextIndex != 0 {
		return fmt.Errorf("root.initialize.fail: registry_not_empty")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	// at index 0 always this contract
	registry.SetAt(nextIndex.Bytes(), chainID[:])
	ctx.Publishf("root.initialize.success")
	return nil
}

func newContract(ctx vmtypes.Sandbox, params kv.RCodec) error {
	ctx.Publishf("root.newContract.begin")

	return nil
}
