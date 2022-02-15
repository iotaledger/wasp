package viewcontext

import (
	"github.com/iotaledger/wasp/packages/vm"
	"math/big"

	"github.com/iotaledger/wasp/packages/vm/gas"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

type sandboxview struct {
	contractHname iscp.Hname
	params        iscp.Params
	state         kv.KVStoreReader
	vctx          *Viewcontext
	assertObj     *assert.Assert
	// gas related
	gasBudget uint64
	gasBurned uint64
}

var _ iscp.SandboxView = &sandboxview{}

var getChainInfoHname = governance.FuncGetChainInfo.Hname()

func newSandboxView(vctx *Viewcontext, contractHname iscp.Hname, params dict.Dict) *sandboxview {
	return &sandboxview{
		vctx:          vctx,
		contractHname: contractHname,
		params: iscp.Params{
			Dict:      params,
			KVDecoder: kvdecoder.New(params, vctx.log),
		},
		state:     contractStateSubpartition(vctx.stateReader.KVStoreReader(), contractHname),
		gasBudget: 1_000_000, // TODO sensible upper limit for view calls
	}
}

func (s *sandboxview) assert() *assert.Assert {
	if s.assertObj == nil {
		s.assertObj = assert.NewAssert(s.vctx)
	}
	return s.assertObj
}

func (s *sandboxview) AccountID() *iscp.AgentID {
	hname := s.contractHname
	switch hname {
	case root.Contract.Hname(), accounts.Contract.Hname(), blob.Contract.Hname(), blocklog.Contract.Hname():
		hname = 0
	}
	return iscp.NewAgentID(s.vctx.chainID.AsAddress(), hname)
}

func (s *sandboxview) BalanceIotas() uint64 {
	return accounts.GetIotaBalance(s.state, s.AccountID())
}

func (s *sandboxview) BalanceNativeToken(tokenID *iotago.NativeTokenID) *big.Int {
	return accounts.GetNativeTokenBalance(s.state, s.AccountID(), tokenID)
}

func (s *sandboxview) Assets() *iscp.Assets {
	ret := accounts.GetAssets(s.state, s.AccountID())
	if ret == nil {
		ret = &iscp.Assets{}
	}
	return ret
}

func (s *sandboxview) Call(contractHname, entryPoint iscp.Hname, params dict.Dict) dict.Dict {
	return s.vctx.callView(contractHname, entryPoint, params)
}

func (s *sandboxview) ChainID() *iscp.ChainID {
	return s.vctx.chainID
}

func (s *sandboxview) ChainOwnerID() *iscp.AgentID {
	r := s.Call(governance.Contract.Hname(), getChainInfoHname, nil)
	res := kvdecoder.New(r, s.Log())
	return res.MustGetAgentID(governance.VarChainOwnerID)
}

func (s *sandboxview) Contract() iscp.Hname {
	return s.contractHname
}

func (s *sandboxview) ContractCreator() *iscp.AgentID {
	contractRecord := root.FindContract(contractStateSubpartition(s.vctx.stateReader.KVStoreReader(), root.Contract.Hname()), s.contractHname)
	assert.NewAssert(s.Log()).Requiref(contractRecord != nil, "failed to find contract %s", s.contractHname)
	return contractRecord.Creator
}

func (s *sandboxview) Timestamp() int64 {
	ret, err := s.vctx.stateReader.Timestamp()
	if err != nil {
		s.Log().Panicf("%v", err)
	}
	return ret.UnixNano()
}

func (s *sandboxview) Log() iscp.LogInterface {
	return s.vctx
}

func (s *sandboxview) Params() *iscp.Params {
	return &s.params
}

func (s *sandboxview) State() kv.KVStoreReader {
	return s.state
}

func (s *sandboxview) Utils() iscp.Utils {
	return sandbox.NewUtils(s.Gas())
}

func (s *sandboxview) Gas() iscp.Gas {
	return s
}

func (s *sandboxview) Burn(burnCode gas.BurnCode, par ...uint64) {
	s.gasBurned += burnCode.Cost(par...)
	if s.gasBurned > s.gasBudget {
		panic(vm.ErrGasBudgetExceeded)
	}
}

func (s *sandboxview) Budget() uint64 {
	return s.gasBudgetLeft()
}

func (s *sandboxview) gasBudgetLeft() uint64 {
	if s.gasBudget < s.gasBurned {
		return 0
	}
	return s.gasBudget - s.gasBurned
}

func (s *sandboxview) Requiref(cond bool, format string, args ...interface{}) {
	s.assert().Requiref(cond, format, args...)
}

func (s *sandboxview) RequireNoError(err error, str ...string) {
	s.assert().RequireNoError(err, str...)
}
