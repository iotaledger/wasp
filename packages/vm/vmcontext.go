package vm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
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

func (vctx *VMContext) GetInt64RequestParam(name string) (int64, bool) {
	return vctx.Request.RequestBlock().Params().GetInt64(name)
}

func (vctx *VMContext) SetInt64(name string, value int64) {
	vctx.StateUpdate.AddMutation(variables.NewMutationSet(name, util.Uint64To8Bytes(uint64(value))))
}

func (vctx *VMContext) GetStringRequestParam(name string) (string, bool) {
	return vctx.Request.RequestBlock().Params().GetString(name)
}

func (vctx *VMContext) SetString(name string, value string) {
	vctx.StateUpdate.AddMutation(variables.NewMutationSet(name, []byte(value)))
}
