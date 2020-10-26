// bootup implement bootup processor. Its the only purpose:
// - to deploy core smart contract at index according to provided data
// - store chain id in the state
package factory

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) {
	ctx.Publishf("factory.initialize.begon")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		ctx.Publishf("factory.initialize.fail: already_initalized")
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
	registry := state.GetDictionary(VarContractRegistry)
	nextIndex := coretypes.Uint16(registry.Len())

	if nextIndex != 0 {
		ctx.Publishf("factory.initialize.fail: internal error, registry not empty")
		return
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	state.SetChainID(VarChainID, chainID)
	// at index 0 always this contract
	registry.SetAt(nextIndex.Bytes(), chainID[:])

	ctx.Publishf("bootup.success")
}

func newContract(ctx vmtypes.Sandbox) {
}
