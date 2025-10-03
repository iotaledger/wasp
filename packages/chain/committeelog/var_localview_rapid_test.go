// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog_test

// TODO: Re-enable this test.

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// 	"pgregory.net/rapid"

// 	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

// 	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
// 	"github.com/iotaledger/wasp/v2/clients/iscmove"
// 	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmovetest"
// 	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
// 	"github.com/iotaledger/wasp/v2/packages/cryptolib"
// 	"github.com/iotaledger/wasp/v2/packages/isc"
// 	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
// )

// // A State Machine for for the property based test.
// // It models the chain (confirmed, pending, rejected, rejSync fields)
// // and contains the actual instance to test (lv).
// type localViewVariableSM struct {
// 	//
// 	// The actual instance to test.
// 	lv committeelog.LocalViewVariable
// 	//
// 	// Following stands for the model.
// 	confirmed []*isc.StateAnchor // A chain of confirmed Anchors.
// 	pending   []*isc.StateAnchor // A list of Anchors proposed by the chain, not confirmed yet.
// 	rejected  []*isc.StateAnchor // Rejected Anchors, that should not impact the output anymore.
// 	rejSync   bool               // True, if reject was done and pending was not made empty yet.
// 	//
// 	// Helpers.
// 	utxoIDCounter int // To have unique UTXO IDs.
// }

// var _ rapid.StateMachine = &localViewVariableSM{}

// func newLocalViewVariableSM(t *rapid.T) *localViewVariableSM {
// 	sm := new(localViewVariableSM)
// 	sm.lv = committeelog.NewLocalViewVariable(-1, func(ao *isc.StateAnchor) {}, testlogger.NewLogger(t))
// 	sm.confirmed = []*isc.StateAnchor{}
// 	sm.pending = []*isc.StateAnchor{}
// 	sm.rejected = []*isc.StateAnchor{}
// 	sm.rejSync = false
// 	return sm
// }

// // E.g. external rotation of a TX by other chain.
// //
// // If some external entity produced an Anchor and it was confirmed,
// // all the TX'es proposed by us and not yet confirmed will be rejected.
// func (sm *localViewVariableSM) L1ExternalAnchorConfirmed(t *rapid.T) {
// 	//
// 	// The Anchor from L1 is always respected as the correct one.
// 	newAnchor := sm.nextAnchor()
// 	tipAnchor, tipChanged, _ := sm.lv.AnchorConfirmed(newAnchor)
// 	require.True(t, tipChanged)            // BaseAnchor is replaced or set.
// 	require.Equal(t, newAnchor, tipAnchor)         // BaseAnchor is replaced or set.
// 	require.Equal(t, newAnchor, sm.lv.Value()) // BaseAnchor is replaced or set.
// 	//
// 	// Update the model (add confirmed, move pending to rejected).
// 	sm.confirmed = append(sm.confirmed, newAnchor)
// 	sm.rejected = append(sm.rejected, sm.pending...)
// 	sm.rejSync = false
// 	sm.pending = []*isc.StateAnchor{}
// }

// // E.g. A TX proposed by the consensus was approved.
// //
// // Take single TX from the pending log and approve it.
// func (sm *localViewVariableSM) L1PendingApproved(t *rapid.T) {
// 	//
// 	// Check the preconditions.
// 	if len(sm.pending) == 0 {
// 		t.Skip()
// 	}
// 	//
// 	// Notify the LocalView on the CNF.
// 	cnfAnchor := sm.pending[0]
// 	prevAnchor := sm.lv.Value()
// 	_, tipChanged, _ := sm.lv.AnchorConfirmed(cnfAnchor)
// 	//
// 	// Update the model.
// 	sm.confirmed = append(sm.confirmed, cnfAnchor)
// 	sm.pending = sm.pending[1:]
// 	sm.rejSync = sm.rejSync && len(sm.pending) != 0
// 	//
// 	// Post-condition: If there was no rejection, then the BaseAnchor has to be left unchanged.
// 	if !sm.rejSync && prevAnchor != nil {
// 		require.False(t, tipChanged)            // BaseAnchor is not replaced.
// 		require.Equal(t, prevAnchor, sm.lv.Value()) // BaseAnchor is not replaced.
// 	}
// }

// // E.g. Consensus TX was rejected.
// //
// // All the pending TXes are marked as rejected.
// func (sm *localViewVariableSM) L1PendingRejected(t *rapid.T) {
// 	//
// 	// Check the preconditions.
// 	if len(sm.pending) == 0 {
// 		t.Skip()
// 	}
// 	//
// 	// Notify the LocalView on the rejection.
// 	rejectFrom := rapid.IntRange(0, len(sm.pending)-1).Draw(t, "reject.idx")
// 	newTip, _ := sm.lv.AnchorRejected(sm.pending[rejectFrom])
// 	require.Equal(t, rejectFrom != 0, newTip == nil, "If that't not the first of the pending, then there are pending left, so the new tip is undefined.")
// 	require.Equal(t, rejectFrom == 0, newTip != nil, "In this case, all the pending are marked as rejected, so we have the tip (the confirmed one).")
// 	//
// 	// Update the model.
// 	sm.rejected = append(sm.rejected, sm.pending[rejectFrom+1:]...)
// 	sm.pending = sm.pending[:rejectFrom]
// 	sm.rejSync = len(sm.pending) != 0
// }

// // Handle those outdated rejections.
// func (sm *localViewVariableSM) OutdatedRejectHandled(t *rapid.T) {
// 	//
// 	// Check the preconditions.
// 	if len(sm.rejected) == 0 {
// 		t.Skip()
// 	}
// 	selectedIdx := rapid.IntRange(0, len(sm.rejected)-1).Draw(t, "reject.idx")
// 	selectedAnchor := sm.rejected[selectedIdx]
// 	//
// 	// Perform the action.
// 	_, tipChanged := sm.lv.AnchorRejected(selectedAnchor)
// 	require.False(t, tipChanged)
// 	//
// 	// Update the model.
// 	sm.rejected = append(sm.rejected[:selectedIdx], sm.rejected[selectedIdx+1:]...)
// }

// // Consensus produced a new output.
// func (sm *localViewVariableSM) ConsensusOutput(t *rapid.T) {
// 	//
// 	// Check the preconditions.
// 	if !sm.nextChainStepPossible() {
// 		t.Skip()
// 	}
// 	//
// 	// Perform the action.
// 	prevAnchor := sm.lv.Value()
// 	require.NotNil(t, prevAnchor)
// 	newAnchor := sm.nextAnchor(prevAnchor)
// 	tipAnchor, tipChanged := sm.lv.ConsensusOutputDone(committeelog.NilLogIndex(), prevAnchor.GetObjectRef()) // TODO: LogIndex.
// 	require.True(t, tipChanged)
// 	require.Equal(t, newAnchor, tipAnchor)
// 	require.Equal(t, newAnchor, sm.lv.Value())
// 	//
// 	// Update the model.
// 	sm.pending = append(sm.pending, newAnchor)
// }

// // Here we check the invariants.
// func (sm *localViewVariableSM) Check(t *rapid.T) {
// 	t.Logf("Check, ModelStatus: %v", sm.modelStatus())
// 	t.Logf("Check, %v", sm.lv.StatusString())
// 	sm.propBaseAnchorProposedIfPossible(t)
// 	sm.propBaseAnchorProposedCorrect(t)
// }

// // We don't use randomness to generate Anchors because they have to be unique.
// func (sm *localViewVariableSM) nextAnchor(prevAnchor ...*isc.StateAnchor) *isc.StateAnchor {
// 	sm.utxoIDCounter++
// 	txIDBytes := []byte(fmt.Sprintf("%v", sm.utxoIDCounter))
// 	utxoInput := iotago.UTXOInput{}
// 	copy(utxoInput.TransactionID[:], txIDBytes)
// 	utxoInput.TransactionOutputIndex = 0
// 	if len(prevAnchor) > 1 {
// 		panic("0/1 prevAnchor can be provided")
// 	}
// 	var stateIndex uint32
// 	if len(prevAnchor) == 1 {
// 		stateIndex = prevAnchor[0].GetStateIndex() + 1
// 	} else {
// 		stateIndex = uint32(sm.utxoIDCounter)
// 	}

// 	anchor := iscmovetest.RandomAnchor(iscmovetest.RandomAnchorOption{StateMetadata: &[]byte{}, StateIndex: &stateIndex})
// 	stateAnchor := isc.NewStateAnchor(
// 		&iscmove.AnchorWithRef{
// 			Object:    &anchor,
// 			ObjectRef: *iotatest.RandomObjectRef(),
// 			Owner:     iotatest.RandomAddress(),
// 		}, *cryptolib.NewRandomAddress().AsIotaAddress())
// 	return &stateAnchor
// }

// // Alias output can be proposed, if there is at least one Anchor confirmed and there is no
// // ongoing resync because of rejections.
// func (sm *localViewVariableSM) nextChainStepPossible() bool {
// 	return len(sm.confirmed) != 0 && !sm.rejSync
// }

// // The LocalView proposes next BaseAnchor if there is received at least 1 confirmed output
// // and there is no rejections, that are not reported to the LocalView yet.
// func (sm *localViewVariableSM) propBaseAnchorProposedIfPossible(t *rapid.T) {
// 	require.Equal(t,
// 		sm.nextChainStepPossible(),
// 		sm.lv.Value() != nil,
// 	)
// }

// // If an BaseAnchor is proposed, it matches the last pending, or last confirmed, if there are no pending.
// func (sm *localViewVariableSM) propBaseAnchorProposedCorrect(t *rapid.T) {
// 	if sm.nextChainStepPossible() {
// 		if len(sm.pending) != 0 {
// 			require.Equal(t, sm.pending[len(sm.pending)-1], sm.lv.Value())
// 		} else {
// 			require.Equal(t, sm.confirmed[len(sm.confirmed)-1], sm.lv.Value())
// 		}
// 	}
// }

// // Just for debugging.
// func (sm *localViewVariableSM) modelStatus() string {
// 	str := fmt.Sprintf("Rejected[sync=%v]", sm.rejSync)
// 	for _, e := range sm.rejected {
// 		oid := e.GetObjectID()
// 		str += fmt.Sprintf(" %v", oid[0:4])
// 	}
// 	str += "; Pending"
// 	for _, e := range sm.pending {
// 		oid := e.GetObjectID()
// 		str += fmt.Sprintf(" %v", oid[0:4])
// 	}
// 	return str
// }

// var _ rapid.StateMachine = &localViewVariableSM{}

// // E.g. for special parameters for reproducibility, etc.
// // `go test ./packages/chain/committeeLog/ --run TestPropsRapid -v -rapid.seed=13061922091840831492 -rapid.checks=100`
// func TestLocalViewVariableRapid(t *testing.T) {
// 	rapid.Check(t, func(t *rapid.T) {
// 		sm := newLocalViewVariableSM(t)
// 		t.Repeat(rapid.StateMachineActions(sm))
// 	})
// }
