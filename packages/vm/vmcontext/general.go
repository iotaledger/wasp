package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
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

var _ iscp.StateAnchor = &VMContext{}

func (vmctx *VMContext) StateController() iotago.Address {
	return vmctx.task.AnchorOutput.StateController
}

func (vmctx *VMContext) GovernanceController() iotago.Address {
	return vmctx.task.AnchorOutput.GovernanceController
}

func (vmctx *VMContext) StateIndex() uint32 {
	return vmctx.task.AnchorOutput.StateIndex
}

func (vmctx *VMContext) OutputID() iotago.UTXOInput {
	return vmctx.task.AnchorOutputID
}

func (vmctx *VMContext) StateData() (ret iscp.StateData) {
	var err error
	ret, err = iscp.StateDataFromBytes(vmctx.task.AnchorOutput.StateMetadata)
	if err != nil {
		panic(err)
	}
	return
}
