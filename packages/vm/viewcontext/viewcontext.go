package viewcontext

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type viewcontext struct {
	processors *processors.ProcessorCache
	state      buffered.BufferedKVStore
	chainID    coretypes.ChainID
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
		state:      state.Variables(),
		chainID:    *chain.ID(),
	}, nil
}

func (v *viewcontext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	rec, err := root.FindContract(contractStateSubpartition(v.state, root.Hname), contractHname)
	if err != nil {
		return nil, fmt.Errorf("failed to find contract %s: %v", contractHname, err)
	}

	proc, err := v.processors.GetOrCreateProcessor(rec, func(deploymentHash *hashing.HashValue) ([]byte, error) {
		return root.GetBinary(contractStateSubpartition(v.state, contractHname), *deploymentHash)
	})
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("%s: can't find entry point '%s'", proc.GetDescription(), epCode.String())
	}

	if !ep.IsView() {
		return nil, fmt.Errorf("only view entry point can be called in this context")
	}

	return ep.CallView(newSandboxView(v, coretypes.NewContractID(v.chainID, contractHname), params))
}

//func (v *viewcontext) getBinary(deploymentHash *hashing.HashValue) ([]byte, error) {
//	return root.GetBinary(codec.NewMustCodec(v.state), *deploymentHash)
//}

func contractStateSubpartition(state kv.KVStore, contractHname coretypes.Hname) codec.ImmutableMustCodec {
	return codec.NewMustCodec(subrealm.New(state, kv.Key(contractHname.Bytes())))
}
