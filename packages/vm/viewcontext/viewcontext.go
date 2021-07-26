package viewcontext

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"golang.org/x/xerrors"
)

type Viewcontext struct {
	processors  *processors.Cache
	stateReader state.OptimisticStateReader
	chainID     iscp.ChainID
	log         *logger.Logger
}

func NewFromChain(ch chain.ChainCore) *Viewcontext {
	return New(*ch.ID(), ch.GetStateReader(), ch.Processors(), ch.Log().Named("view"))
}

func New(chainID iscp.ChainID, stateReader state.OptimisticStateReader, proc *processors.Cache, log *logger.Logger) *Viewcontext {
	return &Viewcontext{
		processors:  proc,
		stateReader: stateReader,
		chainID:     chainID,
		log:         log,
	}
}

// CallView in viewcontext implements own panic catcher.
func (v *Viewcontext) CallView(contractHname, epCode iscp.Hname, params dict.Dict) (dict.Dict, error) {
	var ret dict.Dict
	var err error
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			ret = nil
			switch err1 := r.(type) {
			case *kv.DBError:
				v.log.Panicf("DB error: %v", err1)
			case *optimism.ErrorStateInvalidated:
				err = err1
			default:
				err = xerrors.Errorf("viewcontext: panic in VM: %v", err1)
			}
			v.log.Infof("+++++ debug: error returned from the view call: %v", err) // TODO remove
		}()
		ret, err = v.callView(contractHname, epCode, params)
	}()
	return ret, err
}

func (v *Viewcontext) callView(contractHname, epCode iscp.Hname, params dict.Dict) (dict.Dict, error) {
	var err error
	contractRecord, err := root.FindContract(contractStateSubpartition(v.stateReader.KVStoreReader(), root.Contract.Hname()), contractHname)
	if err != nil {
		return nil, fmt.Errorf("inconsistency while searching for contract %s: %v", contractHname, err)
	}
	if contractHname != _default.Contract.Hname() && contractRecord.Hname() == _default.Contract.Hname() {
		// in the view call we do not run default contract
		return nil, fmt.Errorf("can't find contract '%s'", contractHname)
	}
	proc, err := v.processors.GetOrCreateProcessor(contractRecord, func(programHash hashing.HashValue) (string, []byte, error) {
		if vmtype, ok := v.processors.Config.GetNativeProcessorType(programHash); ok {
			return vmtype, nil, nil
		}
		return blob.LocateProgram(contractStateSubpartition(v.stateReader.KVStoreReader(), blob.Contract.Hname()), programHash)
	})
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("trying to call contract '%s': can't find entry point '%s'", proc.GetDescription(), epCode.String())
	}

	if !ep.IsView() {
		return nil, fmt.Errorf("only view entry point can be called in this context")
	}
	return ep.Call(newSandboxView(v, contractHname, params))
}

func contractStateSubpartition(stateKvReader kv.KVStoreReader, contractHname iscp.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(stateKvReader, kv.Key(contractHname.Bytes()))
}

func (v *Viewcontext) Infof(format string, params ...interface{}) {
	v.log.Infof(format, params...)
}

func (v *Viewcontext) Debugf(format string, params ...interface{}) {
	v.log.Debugf(format, params...)
}

func (v *Viewcontext) Panicf(format string, params ...interface{}) {
	v.log.Panicf(format, params...)
}
