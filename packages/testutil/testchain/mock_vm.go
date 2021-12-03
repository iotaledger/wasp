package testchain

import (
	"strings"
	"testing"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/stretchr/testify/require"
)

type MockedVMRunner struct {
	stateTransition *MockedStateTransition
	nextState       state.VirtualStateAccess
	tx              *iotago.TransactionEssence
	log             *logger.Logger
	t               *testing.T
}

func NewMockedVMRunner(t *testing.T, log *logger.Logger) *MockedVMRunner {
	ret := &MockedVMRunner{
		stateTransition: NewMockedStateTransition(t, nil),
		log:             log,
		t:               t,
	}
	ret.stateTransition.OnVMResult(func(vs state.VirtualStateAccess, tx *iotago.TransactionEssence) {
		ret.nextState = vs
		ret.tx = tx
	})
	return ret
}

func (r *MockedVMRunner) Run(task *vm.VMTask) {
	reqstr := strings.Join(iscp.ShortRequestIDs(iscp.TakeRequestIDs(task.Requests...)), ",")

	r.log.Debugf("VM input: state hash: %s, chain input: %s, requests: [%s]",
		task.VirtualStateAccess.StateCommitment(), iscp.OID(task.AnchorOutput.ID()), reqstr)

	r.stateTransition.NextState(task.VirtualStateAccess, task.AnchorOutput, task.Timestamp, task.Requests...)
	task.ResultTransactionEssence = r.tx
	task.VirtualStateAccess = r.nextState
	newOut := transaction.GetAliasOutputFromEssence(task.ResultTransactionEssence, task.AnchorOutput.GetAliasAddress())
	require.NotNil(r.t, newOut)
	require.EqualValues(r.t, task.AnchorOutput.GetStateIndex()+1, newOut.GetStateIndex())
	// essenceHash := hashing.HashData(task.ResultTransactionEssence.Bytes())
	// r.log.Debugf("mockedVMRunner: new state produced: stateIndex: #%d state hash: %s, essence hash: %s stateOutput: %s\n essence : %s",
	//	r.nextState.BlockIndex(), r.nextState.Hash().String(), essenceHash.String(), iscp.OID(newOut.ID()), task.ResultTransactionEssence.String())
	task.OnFinish(nil, nil, nil)
}
