package testchain

import (
	"testing"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
)

type mockedVMRunner struct {
	stateTransition *MockedStateTransition
	nextState       state.VirtualState
	tx              *ledgerstate.TransactionEssence
	log             *logger.Logger
	t               *testing.T
}

func NewMockedVMRunner(t *testing.T, log *logger.Logger) *mockedVMRunner {
	ret := &mockedVMRunner{
		stateTransition: NewMockedStateTransition(t, nil),
		log:             log,
		t:               t,
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
	newOut := transaction.GetAliasOutputFromEssence(task.ResultTransactionEssence, task.ChainInput.GetAliasAddress())
	require.NotNil(r.t, newOut)
	require.EqualValues(r.t, task.ChainInput.GetStateIndex()+1, newOut.GetStateIndex())
	r.log.Debugf("mockedVMRunner: new state produced: stateIndex: #%d state hash: %s, stateOutput: %s",
		r.nextState.BlockIndex(), r.nextState.Hash().String(), coretypes.OID(newOut.ID()))
	task.OnFinish(nil, nil, nil)
}
