package testchain

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
)

type mockedVMRunner struct {
	stateTransition *MockedStateTransition
	nextState       state.VirtualState
	tx              *ledgerstate.TransactionEssence
}

func NewMockedVMRunner(t *testing.T) *mockedVMRunner {
	ret := &mockedVMRunner{
		stateTransition: NewMockedStateTransition(t, nil),
	}
	ret.stateTransition.OnVMResult(func(vs state.VirtualState, tx *ledgerstate.TransactionEssence) {
		ret.nextState = vs
		ret.tx = tx
	})
	return ret
}

func (r *mockedVMRunner) Run(task *vm.VMTask) {
	r.stateTransition.NextState(task.VirtualState, task.ChainInput)
	task.ResultTransactionEssence = r.tx
	task.VirtualState = r.nextState
}
