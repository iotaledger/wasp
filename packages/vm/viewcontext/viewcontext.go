package viewcontext

import (
	"math/big"
	"runtime/debug"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"golang.org/x/xerrors"
)

// viewcontext implements the needed infrastucture to run external view calls, its more lightweight than vmcontext
type Viewcontext struct {
	processors  *processors.Cache
	stateReader state.OptimisticStateReader
	chainID     *iscp.ChainID
	log         *logger.Logger
	chainInfo   *governance.ChainInfo
	gasBurnLog  *gas.BurnLog
	gasBudget   uint64

	// target call specific
	targetContractHname   iscp.Hname
	targetEntryPointHname iscp.Hname
	params                iscp.Params
}

var _ execution.WaspContext = &Viewcontext{}

func New(ch chain.ChainCore) *Viewcontext {
	return &Viewcontext{
		processors:          ch.Processors(),
		stateReader:         ch.GetStateReader(),
		chainID:             ch.ID(),
		log:                 ch.Log(),
		targetContractHname: 0,
	}
}

func (ctx *Viewcontext) contractStateReader(contract iscp.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(ctx.stateReader.KVStoreReader(), kv.Key(contract.Bytes()))
}

func (ctx *Viewcontext) LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	return blob.LocateProgram(ctx.contractStateReader(blob.Contract.Hname()), programHash)
}

func (ctx *Viewcontext) GetContractRecord(contractHname iscp.Hname) (ret *root.ContractRecord) {
	return root.FindContract(ctx.contractStateReader(root.Contract.Hname()), contractHname)
}

func (ctx *Viewcontext) GasBurn(burnCode gas.BurnCode, par ...uint64) {
	g := burnCode.Cost(par...)
	ctx.gasBurnLog.Record(burnCode, g)
	if g > ctx.gasBudget {
		panic(xerrors.Errorf("maximum gas exceeded"))
	}
	ctx.gasBudget -= g
}

func (ctx *Viewcontext) AccountID() *iscp.AgentID {
	hname := ctx.CurrentContractHname()
	if commonaccount.IsCoreHname(hname) {
		return commonaccount.Get(ctx.ChainID())
	}
	return iscp.NewAgentID(ctx.ChainID().AsAddress(), hname)
}

func (ctx *Viewcontext) Processors() *processors.Cache {
	return ctx.processors
}

func (ctx *Viewcontext) GetAssets(agentID *iscp.AgentID) *iscp.Assets {
	return accounts.GetAssets(ctx.contractStateReader(accounts.Contract.Hname()), agentID)
}

func (ctx *Viewcontext) Timestamp() int64 {
	t, err := ctx.stateReader.Timestamp()
	if err != nil {
		ctx.log.Panicf("%v", err)
	}
	return t.UnixNano()
}

func (ctx *Viewcontext) GetIotaBalance(agentID *iscp.AgentID) uint64 {
	return accounts.GetIotaBalance(ctx.contractStateReader(accounts.Contract.Hname()), agentID)
}

func (ctx *Viewcontext) GetNativeTokenBalance(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	return accounts.GetNativeTokenBalance(
		ctx.contractStateReader(accounts.Contract.Hname()),
		agentID,
		tokenID)
}

func (ctx *Viewcontext) Call(targetContract, epCode iscp.Hname, params dict.Dict, _ *iscp.Assets) dict.Dict {
	ctx.log.Debugf("Call. TargetContract: %s entry point: %s", targetContract, epCode)
	return ctx.callView(targetContract, epCode, params)
}

func (ctx *Viewcontext) ChainID() *iscp.ChainID {
	return ctx.chainInfo.ChainID
}

func (ctx *Viewcontext) ChainOwnerID() *iscp.AgentID {
	return ctx.chainInfo.ChainOwnerID
}

func (ctx *Viewcontext) ContractCreator() *iscp.AgentID {
	rec := ctx.GetContractRecord(ctx.targetContractHname)
	if rec == nil {
		panic("can't find current contract")
	}
	return rec.Creator
}

func (ctx *Viewcontext) CurrentContractHname() iscp.Hname {
	return ctx.targetContractHname
}

func (ctx *Viewcontext) Params() *iscp.Params {
	return &ctx.params
}

func (ctx *Viewcontext) StateReader() kv.KVStoreReader {
	return ctx.contractStateReader(ctx.targetContractHname)
}

func (ctx *Viewcontext) GasBudgetLeft() uint64 {
	return ctx.gasBudget
}

func (ctx *Viewcontext) Infof(format string, params ...interface{}) {
	ctx.log.Infof(format, params...)
}

func (ctx *Viewcontext) Debugf(format string, params ...interface{}) {
	ctx.log.Debugf(format, params...)
}

func (ctx *Viewcontext) Panicf(format string, params ...interface{}) {
	ctx.log.Panicf(format, params...)
}

// only for debugging
func (ctx *Viewcontext) GasBurnLog() *gas.BurnLog {
	return ctx.gasBurnLog
}

func (ctx *Viewcontext) callView(targetContract, entryPoint iscp.Hname, params dict.Dict) (ret dict.Dict) {
	contractRecord := ctx.GetContractRecord(targetContract)
	ep := execution.GetEntryPointByProgHash(ctx, targetContract, entryPoint, contractRecord.ProgramHash)

	if !ep.IsView() {
		panic("target entrypoint is not a view")
	}

	return ep.Call(sandbox.NewSandboxView(ctx))
}

func (ctx *Viewcontext) initCallView(targetContract, entryPoint iscp.Hname, params dict.Dict) (ret dict.Dict) {
	ctx.gasBurnLog = gas.NewGasBurnLog()
	ctx.gasBudget = gas.ViewCallGasBudget

	ctx.targetContractHname = targetContract
	ctx.targetEntryPointHname = entryPoint
	ctx.params = iscp.Params{
		Dict:      params,
		KVDecoder: kvdecoder.New(params, ctx.log),
	}

	ctx.chainInfo = governance.MustGetChainInfo(ctx.contractStateReader(governance.Contract.Hname()))

	return ctx.callView(targetContract, entryPoint, params)
}

func (ctx *Viewcontext) CallViewExternal(targetContract, epCode iscp.Hname, params dict.Dict) (ret dict.Dict, err error) {
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			ret = nil
			switch err1 := r.(type) {
			case *kv.DBError:
				ctx.log.Panicf("DB error: %v", err1)
			case error:
				err = err1
			default:
				err = xerrors.Errorf("viewcontext: panic in VM: %v", err1)
			}
			ctx.log.Debugf("CallView: %v", err)
			ctx.log.Debugf(string(debug.Stack()))
		}()
		ret = ctx.initCallView(targetContract, epCode, params)
	}()
	return ret, err
}
