package vm

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
)

type sandboximpl struct {
	*VMContext
	saveTxBuilder *TransactionBuilder // for rollback
}

func NewSandbox(vctx *VMContext) Sandbox {
	return &sandboximpl{
		VMContext:     vctx,
		saveTxBuilder: vctx.TxBuilder.Clone(),
	}
}

// Sandbox interface

// clear all updates, restore same context as in the beginning
func (vctx *sandboximpl) Rollback() {
	vctx.TxBuilder = vctx.saveTxBuilder
	vctx.StateUpdate.Clear()
}

func (vctx *sandboximpl) GetAddress() address.Address {
	return vctx.Address
}

func (vctx *sandboximpl) GetTimestamp() int64 {
	return vctx.Timestamp
}

func (vctx *sandboximpl) Index() uint32 {
	return vctx.VariableState.StateIndex()
}

func (vctx *sandboximpl) ID() sctransaction.RequestId {
	return *vctx.RequestRef.RequestId()
}

func (vctx *sandboximpl) GetLog() *logger.Logger {
	return vctx.Log
}

func (vctx *sandboximpl) Code() sctransaction.RequestCode {
	return vctx.RequestRef.RequestBlock().RequestCode()
}

// request arguments

func (vctx *sandboximpl) Request() Request {
	return vctx
}

func (vctx *sandboximpl) GetInt64(name string) (int64, bool) {
	return vctx.RequestRef.RequestBlock().Params().MustGetInt64(name)
}

func (vctx *sandboximpl) GetString(name string) (string, bool) {
	return vctx.RequestRef.RequestBlock().Params().GetString(name)
}

func (vctx *sandboximpl) State() State {
	return vctx
}

func (vctx *sandboximpl) SetInt64(name string, value int64) {
	vctx.StateUpdate.Mutations().Add(variables.NewMutationSet(name, util.Uint64To8Bytes(uint64(value))))
}

func (vctx *sandboximpl) SetString(name string, value string) {
	vctx.StateUpdate.Mutations().Add(variables.NewMutationSet(name, []byte(value)))
}
