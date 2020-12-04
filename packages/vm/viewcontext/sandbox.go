package viewcontext

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
)

var log *logger.Logger

var logOnce sync.Once

func initLogger() {
	log = logger.NewLogger("viewcontext")
}

type sandboxview struct {
	vctx       *viewcontext
	params     dict.Dict
	state      kv.KVStore
	contractID coretypes.ContractID
	events     vm.ContractEventPublisher
}

func newSandboxView(vctx *viewcontext, contractID coretypes.ContractID, params dict.Dict) *sandboxview {
	logOnce.Do(initLogger)
	return &sandboxview{
		vctx:       vctx,
		params:     params,
		state:      contractStateSubpartition(vctx.state, contractID.Hname()),
		contractID: contractID,
		events:     vm.NewContractEventPublisher(contractID, log),
	}
}

func (s *sandboxview) Params() dict.Dict {
	return s.params
}

func (s *sandboxview) State() kv.KVStore {
	return s.state
}

func (s *sandboxview) Balances() coretypes.ColoredBalances {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	return s.vctx.CallView(contractHname, entryPoint, params)
}

func (s *sandboxview) ContractID() coretypes.ContractID {
	return s.contractID
}

func (s *sandboxview) Event(msg string) {
	s.events.Publish(msg)
}

func (s *sandboxview) Eventf(format string, args ...interface{}) {
	s.events.Publishf(format, args...)
}
