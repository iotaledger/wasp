package viewcontext

import (
	"fmt"
	"runtime/debug"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"golang.org/x/xerrors"
)

type viewcontext struct {
	processors *processors.ProcessorCache
	state      kv.KVStore
	chainID    coretypes.ChainID
	timestamp  int64
	log        *logger.Logger
}

func NewFromDB(dbp *dbprovider.DBProvider, chainID coretypes.ChainID, proc *processors.ProcessorCache) (*viewcontext, error) {
	state_, ok, err := state.LoadSolidState(dbp, &chainID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("solid state not found for chain %s", chainID.String())
	}
	return New(chainID, state_.KVStore(), state_.Timestamp(), proc, nil), nil
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

// CallView in viewcontext implements own panic catcher.
func (v *viewcontext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	var ret dict.Dict
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				ret = nil
				err = xerrors.Errorf("recovered from panic in VM: %v", r)
				v.log.Debugf(string(debug.Stack()))
				if dberr, ok := r.(*kv.DBError); ok {
					// There was an error accessing DB. The world stops
					v.log.Panicf("DB error: %v", dberr)
				}
			}
		}()
		ret, err = v.mustCallView(contractHname, epCode, params)
	}()
	return ret, err
}

func (v *viewcontext) mustCallView(contractHname coretypes.Hname, epCode coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	var err error
	contractRecord, err := root.FindContract(contractStateSubpartition(v.state, root.Interface.Hname()), contractHname)
	if err != nil {
		return nil, fmt.Errorf("failed to find contract %s: %v", contractHname, err)
	}
	proc, err := v.processors.GetOrCreateProcessor(contractRecord, func(programHash hashing.HashValue) (string, []byte, error) {
		if vmtype, ok := processors.GetBuiltinProcessorType(programHash); ok {
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
	return ep.Call(newSandboxView(v, contractHname, params))
}

func contractStateSubpartition(state kv.KVStore, contractHname coretypes.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(state, kv.Key(contractHname.Bytes()))
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
