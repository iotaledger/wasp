package vm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
)

// context of one VM call (for one request)
type VMContext struct {
	// invariant through the batch
	// address of the smart contract
	Address address.Address
	// programHash
	ProgramHash hashing.HashValue
	// owner address
	OwnerAddress address.Address
	// reward address
	RewardAddress address.Address
	// minimum reward
	MinimumReward int64
	// tx builder to build the final transaction
	TxBuilder *TransactionBuilder
	// timestamp of the batch
	Timestamp int64
	// initial state of the batch
	VariableState state.VariableState
	// set for each call
	Request sctransaction.RequestRef
	// IsEmpty state update upon call, result of the call.
	StateUpdate state.StateUpdate
	// log
	Log *logger.Logger
}

// Sandbox interface

func (vctx *VMContext) GetAddress() address.Address {
	return vctx.Address
}

func (vctx *VMContext) GetTimestamp() int64 {
	return vctx.Timestamp
}

func (vctx *VMContext) GetStateIndex() uint32 {
	return vctx.VariableState.StateIndex()
}

func (vctx *VMContext) GetRequestID() sctransaction.RequestId {
	return *vctx.Request.RequestId()
}

func (vctx *VMContext) GetLog() *logger.Logger {
	return vctx.Log
}

func (vctx *VMContext) GetRequestCode() sctransaction.RequestCode {
	return vctx.Request.RequestBlock().RequestCode()
}

func (vctx *VMContext) GetIntRequest(name string) (int, bool) {
	return vctx.Request.RequestBlock().Variables().GetInt(name)
}

func (vctx *VMContext) SetInt(name string, value int) {
	vctx.StateUpdate.Variables().Set(name, uint32(value))
}

func (vctx *VMContext) GetStringRequest(name string) (string, bool) {
	s, ok := vctx.Request.RequestBlock().Variables().Get(name)
	if !ok {
		return "", false
	}
	ret, ok := s.(string)
	if !ok {
		return "", false
	}
	return ret, true
}

func (vctx *VMContext) SetString(name string, value string) {
	vctx.StateUpdate.Variables().Set(name, value)
}
