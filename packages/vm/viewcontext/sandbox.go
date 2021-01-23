package viewcontext

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

var (
	logDefault *logger.Logger
)

func InitLogger() {
	logDefault = logger.NewLogger("view")
}

type sandboxview struct {
	vctx          *viewcontext
	contractHname coretypes.Hname
	params        dict.Dict
	state         kv.KVStore // TODO change to KVStoreReader when Writable store removed from wasmhost
	events        vm.ContractEventPublisher
}

func newSandboxView(vctx *viewcontext, contractHname coretypes.Hname, params dict.Dict) *sandboxview {
	return &sandboxview{
		vctx:          vctx,
		contractHname: contractHname,
		params:        params,
		state:         contractStateSubpartition(vctx.state, contractHname),
		events:        vm.NewContractEventPublisher(coretypes.NewContractID(vctx.chainID, contractHname), vctx.log),
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
	return coretypes.NewContractID(s.vctx.chainID, s.contractHname)
}

func (s *sandboxview) Log() coretypes.LogInterface {
	return s.vctx
}

func (s *sandboxview) ChainID() coretypes.ChainID {
	return s.vctx.chainID
}

var getChainInfoHname = coretypes.Hn(root.FuncGetChainInfo)

func (s *sandboxview) ChainOwnerID() coretypes.AgentID {
	r, err := s.Call(root.Interface.Hname(), getChainInfoHname, nil)
	if err != nil {
		s.Log().Panicf("ChainOwnerID: %v", err)
	}
	ret, exists, err := codec.DecodeAgentID(r.MustGet(root.VarChainOwnerID))
	if err != nil || !exists {
		s.Log().Panicf("ChainOwnerID: %v", err)
	}
	return ret
}

func (s *sandboxview) ContractCreator() coretypes.AgentID {
	return s.vctx.contractRecord.Creator
}

func (s *sandboxview) GetTimestamp() int64 {
	return s.vctx.timestamp
}
