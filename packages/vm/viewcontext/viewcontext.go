package viewcontext

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/hardcoded"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type viewcontext struct {
	processors *processors.ProcessorCache
	state      kv.KVStore //buffered.BufferedKVStore
	chainID    coretypes.ChainID
	timestamp  int64
	log        *logger.Logger
}

func NewFromDB(chainID coretypes.ChainID, proc *processors.ProcessorCache) (*viewcontext, error) {
	state_, _, ok, err := state.LoadSolidState(&chainID)

	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("solid state not found for chain %s", chainID.String())
	}
	return New(chainID, state_.Variables(), state_.Timestamp(), proc, nil), nil
}

func New(chainID coretypes.ChainID, state kv.KVStore, ts int64, proc *processors.ProcessorCache, logSet *logger.Logger) *viewcontext {
	if logSet == nil {
		logSet = logDefault
	} else {
		logSet = logSet.Named("view")
	}
	return &viewcontext{
		processors: proc,
		state:      state,
		chainID:    chainID,
		timestamp:  ts,
		log:        logSet,
	}
}

func (v *viewcontext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	rec, err := root.FindContract(contractStateSubpartition(v.state, root.Interface.Hname()), contractHname)
	if err != nil {
		return nil, fmt.Errorf("failed to find contract %s: %v", contractHname, err)
	}

	proc, err := v.processors.GetOrCreateProcessor(rec, func(programHash hashing.HashValue) (string, []byte, error) {
		if vmtype, ok := hardcoded.LocateHardcodedProgram(programHash); ok {
			return vmtype, nil, nil
		}
		return blob.LocateProgram(contractStateSubpartition(v.state, blob.Interface.Hname()), programHash)
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

func contractStateSubpartition(state kv.KVStore, contractHname coretypes.Hname) kv.KVStore {
	return subrealm.New(state, kv.Key(contractHname.Bytes()))
}

func (v *viewcontext) Infof(format string, params ...interface{}) {
	v.log.Infof(format, params...)
}

func (v *viewcontext) Debugf(format string, params ...interface{}) {
	v.log.Debugf(format, params...)
}

func (v *viewcontext) Panicf(format string, params ...interface{}) {
	v.log.Panicf(format, params...)
}
