package viewcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/sandbox/sandbox_utils"
)

type sandboxview struct {
	contractHname iscp.Hname
	params        dict.Dict
	state         kv.KVStoreReader
	vctx          *Viewcontext
}

var _ iscp.SandboxView = &sandboxview{}

var getChainInfoHname = root.FuncGetChainConfig.Hname()

func newSandboxView(vctx *Viewcontext, contractHname iscp.Hname, params dict.Dict) *sandboxview {
	return &sandboxview{
		vctx:          vctx,
		contractHname: contractHname,
		params:        params,
		state:         contractStateSubpartition(vctx.stateReader.KVStoreReader(), contractHname),
	}
}

func (s *sandboxview) AccountID() *iscp.AgentID {
	hname := s.contractHname
	switch hname {
	case root.Contract.Hname(), accounts.Contract.Hname(), blob.Contract.Hname(), blocklog.Contract.Hname():
		hname = 0
	}
	return iscp.NewAgentID(s.vctx.chainID.AsAddress(), hname)
}

func (s *sandboxview) Balances() colored.Balances {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractHname, entryPoint iscp.Hname, params dict.Dict) (dict.Dict, error) {
	return s.vctx.CallView(contractHname, entryPoint, params)
}

func (s *sandboxview) ChainID() *iscp.ChainID {
	return &s.vctx.chainID
}

func (s *sandboxview) ChainOwnerID() *iscp.AgentID {
	r, err := s.Call(root.Contract.Hname(), getChainInfoHname, nil)
	a := assert.NewAssert(s.Log())
	a.RequireNoError(err)
	res := kvdecoder.New(r, s.Log())
	return res.MustGetAgentID(root.VarChainOwnerID)
}

func (s *sandboxview) Contract() iscp.Hname {
	return s.contractHname
}

func (s *sandboxview) ContractCreator() *iscp.AgentID {
	contractRecord, found := root.FindContract(contractStateSubpartition(s.vctx.stateReader.KVStoreReader(), root.Contract.Hname()), s.contractHname)
	assert.NewAssert(s.Log()).Require(found, "failed to find contract %s", s.contractHname)
	return contractRecord.Creator
}

func (s *sandboxview) GetTimestamp() int64 {
	ret, err := s.vctx.stateReader.Timestamp()
	if err != nil {
		s.Log().Panicf("%v", err)
	}
	return ret.UnixNano()
}

func (s *sandboxview) Log() iscp.LogInterface {
	return s.vctx
}

func (s *sandboxview) Params() dict.Dict {
	return s.params
}

func (s *sandboxview) State() kv.KVStoreReader {
	return s.state
}

func (s *sandboxview) Utils() iscp.Utils {
	return sandbox_utils.NewUtils()
}
