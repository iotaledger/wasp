// nil processor takes any request and dos nothing, i.e. produces empty state update
// it is useful for testing
package vmnil

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const ProgramHash = "67F3YgmwXT23PuRwVzDYNLhyXxwQz8WubwmYoWK2hUmE"

type nilProcessor struct {
}

func GetProcessor() vmtypes.Processor {
	return nilProcessor{}
}

func (v nilProcessor) GetEntryPoint(_ coretypes.Hname) (vmtypes.EntryPoint, bool) {
	return v, true
}

func (v nilProcessor) GetDescription() string {
	return "Empty (nil) hard coded smart contract processor VMNil"
}

// does nothing, i.e. resulting state update is empty
func (v nilProcessor) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	reqId := ctx.AccessRequest().ID()
	ctx.Eventf("run nilProcessor. Req.code %s, Contract ID: %s, ts: %d, reqid: %s",
		ctx.AccessRequest().EntryPointCode().String(),
		ctx.CurrentContractID().String(),
		ctx.GetTimestamp(),
		reqId.String(),
	)
	return nil, nil
}

// TODO
func (ep nilProcessor) IsView() bool {
	return false
}

// TODO
func (ep nilProcessor) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	panic("implement me")
}

func (v nilProcessor) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}
