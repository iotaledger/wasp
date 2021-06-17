package vmcontext

import (
	"errors"
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/root"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var ErrContractNotFound = errors.New("contract not found")

// Call
func (vmctx *VMContext) Call(targetContract coretypes.Hname, epCode coretypes.Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances) (dict.Dict, error) {
	vmctx.log.Debugw("Call", "targetContract", targetContract, "epCode", epCode.String())
	rec, ok := vmctx.findContractByHname(targetContract)
	if !ok {
		return nil, ErrContractNotFound
	}
	return vmctx.callByProgramHash(targetContract, epCode, params, transfer, rec.ProgramHash)
}

func (vmctx *VMContext) callByProgramHash(targetContract coretypes.Hname, epCode coretypes.Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances, progHash hashing.HashValue) (dict.Dict, error) {
	proc, err := vmctx.processors.GetOrCreateProcessorByProgramHash(progHash, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		ep = proc.GetDefaultEntryPoint()
	}
	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		if epCode == coretypes.EntryPointInit {
			return nil, fmt.Errorf("'init' entry point can't be a view")
		}
		// passing nil as transfer: calling the view should not have effect on chain ledger
		if err := vmctx.pushCallContextWithTransfer(targetContract, params, nil); err != nil {
			return nil, err
		}
		defer vmctx.popCallContext()

		return ep.Call(NewSandboxView(vmctx))
	}
	if err := vmctx.pushCallContextWithTransfer(targetContract, params, transfer); err != nil {
		return nil, err
	}
	defer vmctx.popCallContext()

	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == coretypes.EntryPointInit && targetContract != root.Interface.Hname() {
		if !vmctx.callerIsRoot() {
			return nil, fmt.Errorf("attempt to callByProgramHash init not from the root contract")
		}
	}
	return ep.Call(NewSandbox(vmctx))
}

func (vmctx *VMContext) callNonViewByProgramHash(targetContract coretypes.Hname, epCode coretypes.Hname, params dict.Dict, transfer *ledgerstate.ColoredBalances, progHash hashing.HashValue) (dict.Dict, error) {
	proc, err := vmctx.processors.GetOrCreateProcessorByProgramHash(progHash, vmctx.getBinary)
	if err != nil {
		return nil, err
	}
	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		ep = proc.GetDefaultEntryPoint()
	}
	// distinguishing between two types of entry points. Passing different types of sandboxes
	if ep.IsView() {
		return nil, fmt.Errorf("non-view entry point expected")
	}
	if err := vmctx.pushCallContextWithTransfer(targetContract, params, transfer); err != nil {
		return nil, err
	}
	defer vmctx.popCallContext()

	// prevent calling 'init' not from root contract or not while initializing root
	if epCode == coretypes.EntryPointInit && targetContract != root.Interface.Hname() {
		if !vmctx.callerIsRoot() {
			return nil, fmt.Errorf("attempt to callByProgramHash init not from the root contract")
		}
	}
	return ep.Call(NewSandbox(vmctx))
}

func (vmctx *VMContext) callerIsRoot() bool {
	caller := vmctx.Caller()
	if !caller.Address().Equals(vmctx.chainID.AsAddress()) {
		return false
	}
	return caller.Hname() == root.Interface.Hname()
}

func (vmctx *VMContext) requesterIsLocal() bool {
	return vmctx.chainOwnerID.Equals(vmctx.req.SenderAccount()) ||
		vmctx.chainID.AsAddress().Equals(vmctx.req.SenderAccount().Address())
}

func (vmctx *VMContext) Params() dict.Dict {
	return vmctx.getCallContext().params
}
