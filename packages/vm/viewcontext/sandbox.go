package viewcontext

import (
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm"
)

var log *logger.Logger

var logOnce sync.Once

func initLogger() {
	log = logger.NewLogger("viewcontext")
}

type sandboxview struct {
	vctx       *viewcontext
	params     codec.ImmutableCodec
	state      codec.ImmutableMustCodec
	contractID coret.ContractID
	events     vm.ContractEventPublisher
}

func newSandboxView(vctx *viewcontext, contractID coret.ContractID, params codec.ImmutableCodec) *sandboxview {
	logOnce.Do(initLogger)
	return &sandboxview{
		vctx:       vctx,
		params:     params,
		state:      contractStateSubpartition(vctx.state, contractID.Hname()),
		contractID: contractID,
		events:     vm.NewContractEventPublisher(contractID, log),
	}
}

func (s *sandboxview) Params() codec.ImmutableCodec {
	return s.params
}

func (s *sandboxview) State() codec.ImmutableMustCodec {
	return s.state
}

func (s *sandboxview) MyBalances() coret.ColoredBalances {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractHname coret.Hname, entryPoint coret.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vctx.CallView(contractHname, entryPoint, params)
}

func (s *sandboxview) MyContractID() coret.ContractID {
	return s.contractID
}

func (s *sandboxview) Event(msg string) {
	s.events.Publish(msg)
}

func (s *sandboxview) Eventf(format string, args ...interface{}) {
	s.events.Publishf(format, args...)
}
