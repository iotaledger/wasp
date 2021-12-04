package vmcontext

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"golang.org/x/xerrors"
)

func (vmctx *VMContext) ChainID() *iscp.ChainID {
	return (*iscp.ChainID)(&vmctx.task.AnchorOutput.AliasID)
}

func (vmctx *VMContext) ChainOwnerID() *iscp.AgentID {
	return vmctx.chainOwnerID
}

func (vmctx *VMContext) ContractCreator() *iscp.AgentID {
	rec, ok := vmctx.findContractByHname(vmctx.CurrentContractHname())
	if !ok {
		panic("can't find current contract")
	}
	return rec.Creator
}

func (vmctx *VMContext) CurrentContractHname() iscp.Hname {
	return vmctx.getCallContext().contract
}

func (vmctx *VMContext) MyAgentID() *iscp.AgentID {
	return iscp.NewAgentID(vmctx.ChainID().AsAddress(), vmctx.CurrentContractHname())
}

func (vmctx *VMContext) Caller() *iscp.AgentID {
	return vmctx.getCallContext().caller
}

func (vmctx *VMContext) Timestamp() int64 {
	return vmctx.virtualState.Timestamp().UnixNano()
}

func (vmctx *VMContext) Entropy() hashing.HashValue {
	return vmctx.entropy
}

func (vmctx *VMContext) Request() iscp.Request {
	return vmctx.req
}

func (vmctx *VMContext) AccountID() *iscp.AgentID {
	hname := vmctx.CurrentContractHname()
	if commonaccount.IsCoreHname(hname) {
		return commonaccount.Get(vmctx.ChainID())
	}
	return iscp.NewAgentID(vmctx.task.AnchorOutput.AliasID.ToAddress(), hname)
}

func (vmctx *VMContext) IncomingTransfer() *iscp.Assets {
	return vmctx.getCallContext().transfer
}

func (vmctx *VMContext) StateAnchor() *iscp.StateAnchor {
	sd, err := iscp.StateDataFromBytes(vmctx.task.AnchorOutput.StateMetadata)
	if err != nil {
		panic(xerrors.Errorf("StateAnchor: %w", err))
	}
	return &iscp.StateAnchor{
		Output:               vmctx.task.AnchorOutput,
		StateController:      vmctx.task.AnchorOutput.StateController,
		GovernanceController: vmctx.task.AnchorOutput.GovernanceController,
		StateIndex:           vmctx.task.AnchorOutput.StateIndex,
		OutputID:             vmctx.task.AnchorOutputID,
		StateData:            sd,
		Deposit:              0,
	}
}
