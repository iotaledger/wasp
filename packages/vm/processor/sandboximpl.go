package processor

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"github.com/iotaledger/wasp/packages/vm"
)

type sandbox struct {
	*vm.VMContext
	saveTxBuilder *vm.TransactionBuilder // for rollback
}

func NewSandbox(vctx *vm.VMContext) Sandbox {
	return &sandbox{
		VMContext:     vctx,
		saveTxBuilder: vctx.TxBuilder.Clone(),
	}
}

// Sandbox interface

// clear all updates, restore same context as in the beginning
func (vctx *sandbox) Rollback() {
	vctx.TxBuilder = vctx.saveTxBuilder
	vctx.StateUpdate.Clear()
}

func (vctx *sandbox) GetAddress() address.Address {
	return vctx.Address
}

func (vctx *sandbox) GetTimestamp() int64 {
	return vctx.Timestamp
}

func (vctx *sandbox) Index() uint32 {
	return vctx.VirtualState.StateIndex()
}

func (vctx *sandbox) ID() sctransaction.RequestId {
	return *vctx.RequestRef.RequestId()
}

func (vctx *sandbox) GetLog() *logger.Logger {
	return vctx.Log
}

func (vctx *sandbox) Code() sctransaction.RequestCode {
	return vctx.RequestRef.RequestBlock().RequestCode()
}

// request arguments

func (vctx *sandbox) Request() Request {
	return vctx
}

func (vctx *sandbox) GetInt64(name string) (int64, bool) {
	return vctx.RequestRef.RequestBlock().Args().MustGetInt64(name)
}

func (vctx *sandbox) GetString(name string) (string, bool) {
	return vctx.RequestRef.RequestBlock().Args().GetString(name)
}

func (vctx *sandbox) GetAddressValue(name string) (address.Address, bool) {
	return vctx.RequestRef.RequestBlock().Args().MustGetAddress(name)
}

func (vctx *sandbox) GetHashValue(name string) (hashing.HashValue, bool) {
	return vctx.RequestRef.RequestBlock().Args().MustGetHashValue(name)
}

func (vctx *sandbox) State() State {
	return vctx
}

func (vctx *sandbox) SetInt64(name string, value int64) {
	vctx.StateUpdate.Mutations().Add(variables.NewMutationSet(name, util.Uint64To8Bytes(uint64(value))))
}

func (vctx *sandbox) SetString(name string, value string) {
	vctx.StateUpdate.Mutations().Add(variables.NewMutationSet(name, []byte(value)))
}

func (vctx *sandbox) SetAddressValue(name string, addr address.Address) {
	vctx.StateUpdate.Mutations().Add(variables.NewMutationSet(name, addr[:]))
}

func (vctx *sandbox) SetHashValue(name string, h hashing.HashValue) {
	vctx.StateUpdate.Mutations().Add(variables.NewMutationSet(name, h[:]))
}
