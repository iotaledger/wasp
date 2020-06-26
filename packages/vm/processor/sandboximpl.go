package processor

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"

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
	saveTxBuilder  *txbuilder.Builder // for rollback
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

func (s *stateWrapper) Get(name string) ([]byte, bool) {
	// FIXME: this is O(N) with N = amount of accumulated mutations
	// it could be improved by caching the latest mutation for evey key
	muts := s.stateUpdate.Mutations()
	for i := muts.Len() - 1; i >= 0; i-- {
		m := muts.At(i)
		if (*m).Key() == name {
			// The key-value pair has been modified during the current request
			// return the latest assigned value
			return (*m).Value()
		}
	}

	// The key-value pair has not been modified
	// Fetch its value from the virtual state
	return s.virtualState.Variables().Get(name)
}

func (s *stateWrapper) GetInt64(name string) (int64, bool, error) {
	v, ok := s.Get(name)
	if !ok {
		return 0, false, nil
	}
	if len(v) != 8 {
		return 0, false, fmt.Errorf("variable %s: %v is not an int64", name, v)
	}
	return int64(util.Uint64From8Bytes(v)), true, nil
}

func (s *stateWrapper) Del(name string) {
	s.stateUpdate.Mutations().Add(variables.NewMutationDel(name))
}

func (s *stateWrapper) Set(name string, value []byte) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, value))
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
