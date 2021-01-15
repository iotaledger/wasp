package viewcontext

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

var (
	logDefault *logger.Logger
)

func InitLogger() {
	logDefault = logger.NewLogger("view")
}

type sandboxview struct {
	vctx       *viewcontext
	params     dict.Dict
	state      kv.KVStore
	contractID coretypes.ContractID
	events     vm.ContractEventPublisher
}

func newSandboxView(vctx *viewcontext, contractID coretypes.ContractID, params dict.Dict) *sandboxview {
	return &sandboxview{
		vctx:       vctx,
		params:     params,
		state:      contractStateSubpartition(vctx.state, contractID.Hname()),
		contractID: contractID,
		events:     vm.NewContractEventPublisher(contractID, vctx.log),
	}
}

func (s *sandboxview) Params() dict.Dict {
	return s.params
}

func (s *sandboxview) State() kv.KVStoreReader {
	return s.state
}

func (s *sandboxview) WriteableState() kv.KVStore {
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

func (s *sandboxview) Log() vmtypes.LogInterface {
	return s.vctx
}

func (s *sandboxview) ChainID() coretypes.ChainID {
	return s.vctx.chainID
}

func (s *sandboxview) ChainOwnerID() coretypes.AgentID {
	panic("Implement me")
}

func (s *sandboxview) ContractCreator() coretypes.AgentID {
	panic("Implement me")
}

func (s *sandboxview) GetTimestamp() int64 {
	return s.vctx.timestamp
}
