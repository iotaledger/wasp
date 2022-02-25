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
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// ViewContext implements the needed infrastructure to run external view calls, its more lightweight than vmcontext
type ViewContext struct {
	processors  *processors.Cache
	stateReader state.OptimisticStateReader
	chainID     *iscp.ChainID
	log         *logger.Logger
	chainInfo   *governance.ChainInfo
	gasBurnLog  *gas.BurnLog
	gasBudget   uint64
	callStack   []*callContext
}

var _ execution.WaspContext = &ViewContext{}

func New(ch chain.ChainCore) *ViewContext {
	return &ViewContext{
		processors:  ch.Processors(),
		stateReader: ch.GetStateReader(),
		chainID:     ch.ID(),
		log:         ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
}

func (ctx *ViewContext) contractStateReader(contract iscp.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(ctx.stateReader.KVStoreReader(), kv.Key(contract.Bytes()))
}

func (ctx *ViewContext) LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	return blob.LocateProgram(ctx.contractStateReader(blob.Contract.Hname()), programHash)
}

func (ctx *ViewContext) GetContractRecord(contractHname iscp.Hname) (ret *root.ContractRecord) {
	return root.FindContract(ctx.contractStateReader(root.Contract.Hname()), contractHname)
}

func (ctx *ViewContext) GasBurn(burnCode gas.BurnCode, par ...uint64) {
	g := burnCode.Cost(par...)
	ctx.gasBurnLog.Record(burnCode, g)
	if g > ctx.gasBudget {
		panic(vm.ErrGasBudgetExceeded)
	}
	ctx.gasBudget -= g
}

func (ctx *ViewContext) AccountID() *iscp.AgentID {
	hname := ctx.CurrentContractHname()
	if commonaccount.IsCoreHname(hname) {
		return commonaccount.Get(ctx.ChainID())
	}
	return iscp.NewAgentID(ctx.ChainID().AsAddress(), hname)
}

func (ctx *ViewContext) Processors() *processors.Cache {
	return ctx.processors
}

func (ctx *ViewContext) GetAssets(agentID *iscp.AgentID) *iscp.Assets {
	return accounts.GetAssets(ctx.contractStateReader(accounts.Contract.Hname()), agentID)
}

func (ctx *ViewContext) Timestamp() int64 {
	t, err := ctx.stateReader.Timestamp()
	if err != nil {
		ctx.log.Panicf("%v", err)
	}
	return t.UnixNano()
}

func (ctx *ViewContext) GetIotaBalance(agentID *iscp.AgentID) uint64 {
	return accounts.GetIotaBalance(ctx.contractStateReader(accounts.Contract.Hname()), agentID)
}

func (ctx *ViewContext) GetNativeTokenBalance(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID) *big.Int {
	return accounts.GetNativeTokenBalance(
		ctx.contractStateReader(accounts.Contract.Hname()),
		agentID,
		tokenID)
}

func (ctx *ViewContext) Call(targetContract, epCode iscp.Hname, params dict.Dict, _ *iscp.Assets) dict.Dict {
	ctx.log.Debugf("Call. TargetContract: %s entry point: %s", targetContract, epCode)
	return ctx.callView(targetContract, epCode, params)
}

func (ctx *ViewContext) ChainID() *iscp.ChainID {
	return ctx.chainInfo.ChainID
}

func (ctx *ViewContext) ChainOwnerID() *iscp.AgentID {
	return ctx.chainInfo.ChainOwnerID
}

func (ctx *ViewContext) ContractCreator() *iscp.AgentID {
	rec := ctx.GetContractRecord(ctx.CurrentContractHname())
	if rec == nil {
		panic("can't find current contract")
	}
	return rec.Creator
}

func (ctx *ViewContext) CurrentContractHname() iscp.Hname {
	return ctx.getCallContext().contract
}

func (ctx *ViewContext) Params() *iscp.Params {
	return &ctx.getCallContext().params
}

func (ctx *ViewContext) StateReader() kv.KVStoreReader {
	return ctx.contractStateReader(ctx.CurrentContractHname())
}

func (ctx *ViewContext) GasBudgetLeft() uint64 {
	return ctx.gasBudget
}

func (ctx *ViewContext) Infof(format string, params ...interface{}) {
	ctx.log.Infof(format, params...)
}

func (ctx *ViewContext) Debugf(format string, params ...interface{}) {
	ctx.log.Debugf(format, params...)
}

func (ctx *ViewContext) Panicf(format string, params ...interface{}) {
	ctx.log.Panicf(format, params...)
}

// only for debugging
func (ctx *ViewContext) GasBurnLog() *gas.BurnLog {
	return ctx.gasBurnLog
}

func (ctx *ViewContext) callView(targetContract, entryPoint iscp.Hname, params dict.Dict) (ret dict.Dict) {
	contractRecord := ctx.GetContractRecord(targetContract)
	ep := execution.GetEntryPointByProgHash(ctx, targetContract, entryPoint, contractRecord.ProgramHash)

	if !ep.IsView() {
		panic("target entrypoint is not a view")
	}

	ctx.pushCallContext(targetContract, params)
	defer ctx.popCallContext()

	return ep.Call(sandbox.NewSandboxView(ctx))
}

func (ctx *ViewContext) initCallView(targetContract, entryPoint iscp.Hname, params dict.Dict) (ret dict.Dict) {
	ctx.gasBurnLog = gas.NewGasBurnLog()
	ctx.gasBudget = gas.MaxGasExternalViewCall

	ctx.chainInfo = governance.MustGetChainInfo(ctx.contractStateReader(governance.Contract.Hname()))

	return ctx.callView(targetContract, entryPoint, params)
}

func (ctx *ViewContext) CallViewExternal(targetContract, epCode iscp.Hname, params dict.Dict) (ret dict.Dict, err error) {
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
