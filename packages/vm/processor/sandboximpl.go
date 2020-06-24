package processor

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
	"github.com/iotaledger/wasp/packages/vm"
)

type sandbox struct {
	*vm.VMContext
	saveTxBuilder  *vm.TransactionBuilder // for rollback
	requestWrapper *requestWrapper
	stateWrapper   *stateWrapper
}

type requestWrapper struct {
	ref *sctransaction.RequestRef
}

type stateWrapper struct {
	virtualState state.VirtualState
	stateUpdate  state.StateUpdate
}

func NewSandbox(vctx *vm.VMContext) Sandbox {
	return &sandbox{
		VMContext:      vctx,
		saveTxBuilder:  vctx.TxBuilder.Clone(),
		requestWrapper: &requestWrapper{&vctx.RequestRef},
		stateWrapper:   &stateWrapper{vctx.VirtualState, vctx.StateUpdate},
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

func (vctx *sandbox) GetLog() *logger.Logger {
	return vctx.Log
}

// request arguments

func (vctx *sandbox) Request() Request {
	return vctx.requestWrapper
}

func (r *requestWrapper) ID() sctransaction.RequestId {
	return *r.ref.RequestId()
}

func (r *requestWrapper) Code() sctransaction.RequestCode {
	return r.ref.RequestBlock().RequestCode()
}

func (r *requestWrapper) GetInt64(name string) (int64, bool) {
	return r.ref.RequestBlock().Args().MustGetInt64(name)
}

func (r *requestWrapper) GetString(name string) (string, bool) {
	return r.ref.RequestBlock().Args().GetString(name)
}

func (r *requestWrapper) GetAddressValue(name string) (address.Address, bool) {
	return r.ref.RequestBlock().Args().MustGetAddress(name)
}

func (r *requestWrapper) GetHashValue(name string) (hashing.HashValue, bool) {
	return r.ref.RequestBlock().Args().MustGetHashValue(name)
}

func (vctx *sandbox) State() State {
	return vctx.stateWrapper
}

func (s *stateWrapper) Index() uint32 {
	return s.virtualState.StateIndex()
}

func (s *stateWrapper) GetInt64(name string) (int64, bool, error) {
	// TODO: look in state update mutations?
	return s.virtualState.Variables().GetInt64(name)
}

func (s *stateWrapper) SetInt64(name string, value int64) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, util.Uint64To8Bytes(uint64(value))))
}

func (s *stateWrapper) SetString(name string, value string) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, []byte(value)))
}

func (s *stateWrapper) SetAddressValue(name string, addr address.Address) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, addr[:]))
}

func (s *stateWrapper) SetHashValue(name string, h hashing.HashValue) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, h[:]))
}
