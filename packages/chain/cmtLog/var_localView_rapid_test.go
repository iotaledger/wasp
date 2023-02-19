// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

// A State Machine for for the property based test.
// It models the chain (confirmed, pending, rejected, rejSync fields)
// and contains the actual instance to test (lv).
type varLocalViewSM struct {
	//
	// The actual instance to test.
	lv cmtLog.VarLocalView
	//
	// Following stands for the model.
	confirmed []*isc.AliasOutputWithID // A chain of confirmed AOs.
	pending   []*isc.AliasOutputWithID // A list of AOs proposed by the chain, not confirmed yet.
	rejected  []*isc.AliasOutputWithID // Rejected AOs, that should not impact the output anymore.
	rejSync   bool                     // True, if reject was done and pending was not made empty yet.
	//
	// Helpers.
	utxoIDCounter int // To have unique UTXO IDs.
}

func (sm *varLocalViewSM) Init(t *rapid.T) {
	sm.lv = cmtLog.NewVarLocalView(testlogger.NewLogger(t))
	sm.confirmed = []*isc.AliasOutputWithID{}
	sm.pending = []*isc.AliasOutputWithID{}
	sm.rejected = []*isc.AliasOutputWithID{}
	sm.rejSync = false
}

// E.g. external rotation of a TX by other chain.
//
// If some external entity produced an AO and it was confirmed,
// all the TX'es proposed by us and not yet confirmed will be rejected.
func (sm *varLocalViewSM) L1ExternalAOConfirmed(t *rapid.T) {
	//
	// The AO from L1 is always respected as the correct one.
	newAO := sm.nextAO()
	tipAO, tipChanged := sm.lv.AliasOutputConfirmed(newAO)
	require.True(t, tipChanged)            // BaseAO is replaced or set.
	require.Equal(t, newAO, tipAO)         // BaseAO is replaced or set.
	require.Equal(t, newAO, sm.lv.Value()) // BaseAO is replaced or set.
	//
	// Update the model (add confirmed, move pending to rejected).
	sm.confirmed = append(sm.confirmed, newAO)
	sm.rejected = append(sm.rejected, sm.pending...)
	sm.rejSync = false
	sm.pending = []*isc.AliasOutputWithID{}
}

// E.g. A TX proposed by the consensus was approved.
//
// Take single TX from the pending log and approve it.
func (sm *varLocalViewSM) L1PendingApproved(t *rapid.T) {
	//
	// Check the preconditions.
	if len(sm.pending) == 0 {
		t.Skip()
	}
	//
	// Notify the LocalView on the CNF.
	cnfAO := sm.pending[0]
	prevAO := sm.lv.Value()
	_, tipChanged := sm.lv.AliasOutputConfirmed(cnfAO)
	//
	// Update the model.
	sm.confirmed = append(sm.confirmed, cnfAO)
	sm.pending = sm.pending[1:]
	sm.rejSync = sm.rejSync && len(sm.pending) != 0
	//
	// Post-condition: If there was no rejection, then the BaseAO has to be left unchanged.
	if !sm.rejSync && prevAO != nil {
		require.False(t, tipChanged)            // BaseAO is not replaced.
		require.Equal(t, prevAO, sm.lv.Value()) // BaseAO is not replaced.
	}
}

// E.g. Consensus TX was rejected.
//
// All the pending TXes are marked as rejected.
func (sm *varLocalViewSM) L1PendingRejected(t *rapid.T) {
	//
	// Check the preconditions.
	if len(sm.pending) == 0 {
		t.Skip()
	}
	//
	// Notify the LocalView on the rejection.
	rejectFrom := rapid.IntRange(0, len(sm.pending)-1).Draw(t, "reject.idx")
	newTip, _ := sm.lv.AliasOutputRejected(sm.pending[rejectFrom])
	require.Equal(t, rejectFrom != 0, newTip == nil, "If that't not the first of the pending, then there are pending left, so the new tip is undefined.")
	require.Equal(t, rejectFrom == 0, newTip != nil, "In this case, all the pending are marked as rejected, so we have the tip (the confirmed one).")
	//
	// Update the model.
	sm.rejected = append(sm.rejected, sm.pending[rejectFrom+1:]...)
	sm.pending = sm.pending[:rejectFrom]
	sm.rejSync = len(sm.pending) != 0
}

// Handle those outdated rejections.
func (sm *varLocalViewSM) OutdatedRejectHandled(t *rapid.T) {
	//
	// Check the preconditions.
	if len(sm.rejected) == 0 {
		t.Skip()
	}
	selectedIdx := rapid.IntRange(0, len(sm.rejected)-1).Draw(t, "reject.idx")
	selectedAO := sm.rejected[selectedIdx]
	//
	// Perform the action.
	_, tipChanged := sm.lv.AliasOutputRejected(selectedAO)
	require.False(t, tipChanged)
	//
	// Update the model.
	sm.rejected = append(sm.rejected[:selectedIdx], sm.rejected[selectedIdx+1:]...)
}

// Consensus produced a new output.
func (sm *varLocalViewSM) ConsensusOutput(t *rapid.T) {
	//
	// Check the preconditions.
	if !sm.nextChainStepPossible() {
		t.Skip()
	}
	//
	// Perform the action.
	prevAO := sm.lv.Value()
	require.NotNil(t, prevAO)
	newAO := sm.nextAO(prevAO)
	tipAO, tipChanged := sm.lv.ConsensusOutputDone(prevAO.OutputID(), newAO)
	require.True(t, tipChanged)
	require.Equal(t, newAO, tipAO)
	require.Equal(t, newAO, sm.lv.Value())
	//
	// Update the model.
	sm.pending = append(sm.pending, newAO)
}

// Here we check the invariants.
func (sm *varLocalViewSM) Check(t *rapid.T) {
	t.Logf("Check, ModelStatus: %v", sm.modelStatus())
	t.Logf("Check, %v", sm.lv.StatusString())
	sm.propBaseAOProposedIfPossible(t)
	sm.propBaseAOProposedCorrect(t)
}

// We don't use randomness to generate AOs because they have to be unique.
func (sm *varLocalViewSM) nextAO(prevAO ...*isc.AliasOutputWithID) *isc.AliasOutputWithID {
	sm.utxoIDCounter++
	txIDBytes := []byte(fmt.Sprintf("%v", sm.utxoIDCounter))
	utxoInput := iotago.UTXOInput{}
	copy(utxoInput.TransactionID[:], txIDBytes)
	utxoInput.TransactionOutputIndex = 0
	if len(prevAO) > 1 {
		panic("0/1 prevAO can be provided")
	}
	var stateIndex uint32
	if len(prevAO) == 1 {
		stateIndex = prevAO[0].GetStateIndex() + 1
	} else {
		stateIndex = uint32(sm.utxoIDCounter)
	}
	return isc.NewAliasOutputWithID(&iotago.AliasOutput{StateIndex: stateIndex}, utxoInput.ID())
}

// Alias output can be proposed, if there is at least one AO confirmed and there is no
// ongoing resync because of rejections.
func (sm *varLocalViewSM) nextChainStepPossible() bool {
	return len(sm.confirmed) != 0 && !sm.rejSync
}

// The LocalView proposes next BaseAO if there is received at least 1 confirmed output
// and there is no rejections, that are not reported to the LocalView yet.
func (sm *varLocalViewSM) propBaseAOProposedIfPossible(t *rapid.T) {
	require.Equal(t,
		sm.nextChainStepPossible(),
		sm.lv.Value() != nil,
	)
}

// If an BaseAO is proposed, it matches the last pending, or last confirmed, if there are no pending.
func (sm *varLocalViewSM) propBaseAOProposedCorrect(t *rapid.T) {
	if sm.nextChainStepPossible() {
		if len(sm.pending) != 0 {
			require.Equal(t, sm.pending[len(sm.pending)-1], sm.lv.Value())
		} else {
			require.Equal(t, sm.confirmed[len(sm.confirmed)-1], sm.lv.Value())
		}
	}
}

// Just for debugging.
func (sm *varLocalViewSM) modelStatus() string {
	str := fmt.Sprintf("Rejected[sync=%v]", sm.rejSync)
	for _, e := range sm.rejected {
		oid := e.OutputID()
		str += fmt.Sprintf(" %v", oid[0:4])
	}
	str += "; Pending"
	for _, e := range sm.pending {
		oid := e.OutputID()
		str += fmt.Sprintf(" %v", oid[0:4])
	}
	return str
}

var _ rapid.StateMachine = &varLocalViewSM{}

// E.g. for special parameters for reproducibility, etc.
// `go test ./packages/chain/cmtLog/ --run TestPropsRapid -v -rapid.seed=13061922091840831492 -rapid.checks=100`
func TestVarLocalViewRapid(t *testing.T) {
	rapid.Check(t, rapid.Run[*varLocalViewSM]())
}
