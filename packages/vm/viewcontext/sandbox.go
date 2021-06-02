package viewcontext

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/sandbox/sandbox_utils"
)

var logDefault *logger.Logger

func InitLogger() {
	logDefault = logger.NewLogger("view")
}

type sandboxview struct {
	contractHname coretypes.Hname
	events        vm.ContractEventPublisher
	params        dict.Dict
	state         kv.KVStoreReader
	vctx          *viewcontext
}

var _ coretypes.SandboxView = &sandboxview{}

var getChainInfoHname = coretypes.Hn(root.FuncGetChainInfo)

func newSandboxView(vctx *viewcontext, contractHname coretypes.Hname, params dict.Dict) *sandboxview {
	return &sandboxview{
		vctx:          vctx,
		contractHname: contractHname,
		params:        params,
		state:         contractStateSubpartition(vctx.stateReader.KVStoreReader(), contractHname),
		events:        vm.NewContractEventPublisher(&vctx.chainID, contractHname, vctx.log),
	}
}

func (s *sandboxview) AccountID() *coretypes.AgentID {
	hname := s.contractHname
	switch hname {
	case root.Interface.Hname(), accounts.Interface.Hname(), blob.Interface.Hname(), eventlog.Interface.Hname():
		hname = 0
	}
	return coretypes.NewAgentID(s.vctx.chainID.AsAddress(), hname)
}

func (s *sandboxview) Balances() *ledgerstate.ColoredBalances {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	return s.vctx.CallView(contractHname, entryPoint, params)
}

func (s *sandboxview) ChainID() *coretypes.ChainID {
	return &s.vctx.chainID
}

func (s *sandboxview) ChainOwnerID() *coretypes.AgentID {
	r, err := s.Call(root.Interface.Hname(), getChainInfoHname, nil)
	a := assert.NewAssert(s.Log())
	a.RequireNoError(err)
	res := kvdecoder.New(r, s.Log())
	return res.MustGetAgentID(root.VarChainOwnerID)
}

func (s *sandboxview) Contract() coretypes.Hname {
	return s.contractHname
}

func (s *sandboxview) ContractCreator() *coretypes.AgentID {
	contractRecord, err := root.FindContract(contractStateSubpartition(s.vctx.stateReader.KVStoreReader(), root.Interface.Hname()), s.contractHname)
	if err != nil {
		s.Log().Panicf("failed to find contract %s: %v", s.contractHname, err)
	}
	return contractRecord.Creator
}

func (s *sandboxview) GetTimestamp() int64 {
	return s.vctx.stateReader.Timestamp().UnixNano()
}

func (s *sandboxview) Log() coretypes.LogInterface {
	return s.vctx
}

func (s *sandboxview) Params() dict.Dict {
	return s.params
}

func (s *sandboxview) State() kv.KVStoreReader {
	return s.state
}

func (s *sandboxview) Utils() coretypes.Utils {
	return sandbox_utils.NewUtils()
}
