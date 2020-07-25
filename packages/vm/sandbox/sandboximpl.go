package sandbox

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/plugins/publisher"
)

// TODO implement mocked sandbox for testing

type sandbox struct {
	*vm.VMContext
	saveTxBuilder  *txbuilder.Builder // for rollback
	requestWrapper *requestWrapper
	stateWrapper   *stateWrapper
}

func NewSandbox(vctx *vm.VMContext) vmtypes.Sandbox {
	return &sandbox{
		VMContext:      vctx,
		saveTxBuilder:  vctx.TxBuilder.Clone(),
		requestWrapper: &requestWrapper{&vctx.RequestRef},
		stateWrapper:   &stateWrapper{vctx.VirtualState, vctx.StateUpdate},
	}
}

// Sandbox interface

func (vctx *sandbox) IsOriginState() bool {
	return vctx.VirtualState.StateIndex() == 0
}

func (vctx *sandbox) Panic(v interface{}) {
	panic(v)
}

func (vctx *sandbox) Rollback() {
	vctx.TxBuilder = vctx.saveTxBuilder
	vctx.StateUpdate.Clear()
}

func (vctx *sandbox) GetSCAddress() *address.Address {
	return &vctx.Address
}

func (vctx *sandbox) GetOwnerAddress() *address.Address {
	return &vctx.OwnerAddress
}

func (vctx *sandbox) GetTimestamp() int64 {
	return vctx.Timestamp
}

func (vctx *sandbox) GetEntropy() hashing.HashValue {
	return vctx.VMContext.Entropy
}

func (vctx *sandbox) GetWaspLog() *logger.Logger {
	return vctx.Log
}

func (vctx *sandbox) DumpAccount() string {
	return vctx.TxBuilder.Dump()
}

// request arguments

func (vctx *sandbox) AccessRequest() vmtypes.RequestAccess {
	return vctx.requestWrapper
}

func (vctx *sandbox) AccessState() kv.MustCodec {
	return vctx.stateWrapper.MustCodec()
}

func (vctx *sandbox) AccessOwnAccount() vmtypes.AccountAccess {
	return vctx
}

func (vctx *sandbox) SendRequest(par vmtypes.NewRequestParams) bool {
	if par.IncludeReward > 0 {
		availableIotas := vctx.TxBuilder.GetInputBalance(balance.ColorIOTA)
		if par.IncludeReward+1 > availableIotas {
			return false
		}
		err := vctx.TxBuilder.MoveToAddress(*par.TargetAddress, balance.ColorIOTA, par.IncludeReward)
		if err != nil {
			return false
		}
	}
	reqBlock := sctransaction.NewRequestBlock(*par.TargetAddress, par.RequestCode)
	reqBlock.WithTimelock(par.Timelock)
	reqBlock.SetArgs(par.Args)

	if err := vctx.TxBuilder.AddRequestBlock(reqBlock); err != nil {
		return false
	}
	return true
}

func (vctx *sandbox) SendRequestToSelf(reqCode sctransaction.RequestCode, args kv.Map) bool {
	return vctx.SendRequest(vmtypes.NewRequestParams{
		TargetAddress: &vctx.Address,
		RequestCode:   reqCode,
		Args:          args,
		IncludeReward: 0,
	})
}

func (vctx *sandbox) SendRequestToSelfWithDelay(reqCode sctransaction.RequestCode, args kv.Map, delaySec uint32) bool {
	timelock := util.NanoSecToUnixSec(vctx.Timestamp) + delaySec

	return vctx.SendRequest(vmtypes.NewRequestParams{
		TargetAddress: &vctx.Address,
		RequestCode:   reqCode,
		Args:          args,
		Timelock:      timelock,
		IncludeReward: 0,
	})
}

func (vctx *sandbox) Publish(msg string) {
	vctx.Log.Infof("VMMSG: %s %s", vctx.ProgramHash.String(), msg)
	publisher.Publish("vmmsg", vctx.ProgramHash.String(), msg)
}

func (vctx *sandbox) Publishf(format string, args ...interface{}) {
	vctx.Log.Infof("VMMSG: "+format, args...)
	publisher.Publish("vmmsg", vctx.ProgramHash.String(), fmt.Sprintf(format, args...))
}
