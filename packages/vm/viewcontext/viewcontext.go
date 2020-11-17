package viewcontext

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type viewcontext struct {
	processors *processors.ProcessorCache
	state      codec.ImmutableMustCodec
}

func New(chain chain.Chain) (*viewcontext, error) {
	state, _, ok, err := state.LoadSolidState(chain.ID())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("State not found for chain %s", chain.ID())
	}

	return &viewcontext{
		processors: chain.Processors(),
		state:      codec.NewMustCodec(state.Variables()),
	}, nil
}

func (v *viewcontext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	rec, err := root.FindContract(v.state, contractHname)
	if err != nil {
		return nil, fmt.Errorf("failed to find contract %s: %v", contractHname, err)
	}

	proc, err := v.processors.GetOrCreateProcessor(rec, v.getBinary)
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("can't find entry point for entry point '%s'", epCode.String())
	}

	if !ep.IsView() {
		return nil, fmt.Errorf("only view entry point can be called in this context")
	}

	return ep.CallView(NewSandboxView(v, params))
}

func (v *viewcontext) getBinary(deploymentHash *hashing.HashValue) ([]byte, error) {
	return root.GetBinary(v.state, *deploymentHash)
}
