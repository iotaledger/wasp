package committeelog

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func makeAnchor() *isc.StateAnchor {
	a := isc.NewStateAnchor(&iscmove.AnchorWithRef{}, *iotago.MustAddressFromHex("0x0"))
	return &a
}

type vciTestEnv struct {
	vci       *VarConsInsts
	persisted []LogIndex
	cbLIs     []LogIndex // LogIndexes passed to the onLIInc callback.
}

func newVCITestEnv(t *testing.T) *vciTestEnv {
	t.Helper()
	env := &vciTestEnv{}
	env.vci = NewVarConsInsts(
		0,
		func(li LogIndex) {
			env.persisted = append(env.persisted, li)
		},
		func(Output) {},
		testlogger.NewLogger(t),
	)
	return env
}

func (env *vciTestEnv) cb(li LogIndex) gpa.OutMessages {
	env.cbLIs = append(env.cbLIs, li)
	return gpa.NoMessages()
}

func (env *vciTestEnv) lastCBLI() LogIndex {
	if len(env.cbLIs) == 0 {
		return NilLogIndex()
	}
	return env.cbLIs[len(env.cbLIs)-1]
}

// TestVarConsInsts_ConsOutputSkipDefersUntilTick verifies that after a SKIP
// decision, advancing to the next log index is deferred until Tick is called.
func TestVarConsInsts_ConsOutputSkipDefersUntilTick(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci
	vci.lastAnchor = makeAnchor()

	// Before SKIP, only minLI should be present.
	require.Len(t, vci.lis, 1)
	require.Contains(t, vci.lis, LogIndex(0))

	// Call ConsOutputSkip at LI==0. This should NOT immediately advance to LI==1.
	out := vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.Nil(t, out)

	// Still only minLI should be present; next LI should not be proposed yet.
	require.Len(t, vci.lis, 1)
	require.Contains(t, vci.lis, LogIndex(0))

	// Now simulate a Tick; this should apply the pending advancement to LI==1.
	out = vci.Tick(env.cb)
	require.NotNil(t, out)

	// LI==1 should now be in the map and persisted, and the callback should see LI==1.
	require.Contains(t, vci.lis, LogIndex(1))
	require.Equal(t, LogIndex(1), env.lastCBLI())
	require.Contains(t, env.persisted, LogIndex(1))
}

// TestVarConsInsts_ConsecutiveSkipsEachDeferred verifies that multiple
// consecutive SKIP decisions each require their own Tick to advance.
func TestVarConsInsts_ConsecutiveSkipsEachDeferred(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci
	vci.lastAnchor = makeAnchor()

	// First skip at LI==0: should not advance immediately.
	out := vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.Nil(t, out)
	require.Len(t, vci.lis, 1) // Still only LI==0.

	// Tick applies the pending skip → advances to LI==1.
	vci.Tick(env.cb)
	require.Contains(t, vci.lis, LogIndex(1))
	require.Equal(t, LogIndex(1), env.lastCBLI())

	// Second skip at LI==1: should not advance immediately.
	cbCountBefore := len(env.cbLIs)
	out = vci.ConsOutputSkip(LogIndex(1), env.cb)
	require.Nil(t, out)
	require.Equal(t, cbCountBefore, len(env.cbLIs)) // No new callback.
	require.NotContains(t, vci.lis, LogIndex(2))    // LI==2 not yet proposed.

	// Tick applies the second pending skip → advances to LI==2.
	vci.Tick(env.cb)
	require.Contains(t, vci.lis, LogIndex(2))
	require.Equal(t, LogIndex(2), env.lastCBLI())

	// Third skip at LI==2.
	cbCountBefore = len(env.cbLIs)
	out = vci.ConsOutputSkip(LogIndex(2), env.cb)
	require.Nil(t, out)
	require.Equal(t, cbCountBefore, len(env.cbLIs))

	vci.Tick(env.cb)
	require.Contains(t, vci.lis, LogIndex(3))
	require.Equal(t, LogIndex(3), env.lastCBLI())
}

// TestVarConsInsts_SkipWithNoAnchorThenAnchorArrives verifies the path where
// ConsOutputSkip is called when lastAnchor is nil. The advancement should be
// deferred until LatestL1Anchor provides the anchor.
func TestVarConsInsts_SkipWithNoAnchorThenAnchorArrives(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci

	// lastAnchor is nil (default after construction).
	require.Nil(t, vci.lastAnchor)

	// Skip at LI==0 with no anchor: sets lastLI but does not advance.
	out := vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.Nil(t, out)
	require.Len(t, vci.lis, 1)
	require.Empty(t, env.cbLIs)

	// Tick should NOT advance because lastAnchor is still nil.
	vci.Tick(env.cb)
	require.Len(t, vci.lis, 1)
	require.Empty(t, env.cbLIs)

	// Now an L1 anchor arrives via LatestL1Anchor; this should complete the
	// deferred skip and advance to LI==1.
	anchor := makeAnchor()
	vci.LatestL1Anchor(anchor, env.cb)
	require.Contains(t, vci.lis, LogIndex(1))
	require.Equal(t, LogIndex(1), env.lastCBLI())
}

// TestVarConsInsts_SkipWithNoAnchorTickDoesNotAdvance verifies that Tick alone
// cannot resolve a pending skip when lastAnchor is nil.
func TestVarConsInsts_SkipWithNoAnchorTickDoesNotAdvance(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci

	// Skip with no anchor.
	vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.Empty(t, env.cbLIs)

	// Multiple ticks with no anchor should not advance.
	for i := 0; i < 5; i++ {
		vci.Tick(env.cb)
	}
	require.Empty(t, env.cbLIs)
	require.Len(t, vci.lis, 1) // Only LI==0.
}

// TestVarConsInsts_SkipThenDoneBeforeTick verifies that if ConsOutputDone
// arrives before the tick resolves a pending skip, the Done takes precedence
// and the pending skip becomes a no-op (because trySet deduplicates).
func TestVarConsInsts_SkipThenDoneBeforeTick(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci
	vci.lastAnchor = makeAnchor()

	// Skip at LI==0: defers to next tick.
	vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.Empty(t, env.cbLIs)

	// Before the tick, ConsOutputDone arrives at LI==0 with a produced anchor.
	// This immediately advances to LI==1.
	produced := makeAnchor()
	vci.ConsOutputDone(LogIndex(0), produced, env.cb)
	require.Equal(t, LogIndex(1), env.lastCBLI())
	require.Contains(t, vci.lis, LogIndex(1))

	// Now the tick fires. The pending skip targets LI==1, but it's already in
	// the map, so trySet should be a no-op — no double-advance to LI==2.
	cbCountBefore := len(env.cbLIs)
	vci.Tick(env.cb)
	require.Equal(t, cbCountBefore, len(env.cbLIs)) // No new callback.
	require.NotContains(t, vci.lis, LogIndex(2))    // LI==2 should not exist.
}

// TestVarConsInsts_SkipPendingOverwrittenByLaterSkip verifies that if two
// skips happen before a tick, the second one overwrites the first and only
// the later target LI is used.
func TestVarConsInsts_SkipPendingOverwrittenByLaterSkip(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci
	vci.lastAnchor = makeAnchor()

	// Advance to LI==1 via ConsOutputDone so we can skip from LI==1.
	vci.ConsOutputDone(LogIndex(0), makeAnchor(), env.cb)
	require.Contains(t, vci.lis, LogIndex(1))

	// Skip at LI==1 (would pend LI==2).
	vci.ConsOutputSkip(LogIndex(1), env.cb)
	require.True(t, vci.hasPendingAfterSkip)
	require.Equal(t, LogIndex(2), vci.pendingAfterSkipLI)

	// Another skip at LI==1 doesn't change the target since it's the same LI.
	// But if somehow a skip at a higher LI came (e.g. from a parallel path),
	// it would overwrite. Verify the field is consistent.
	vci.ConsOutputSkip(LogIndex(1), env.cb)
	require.Equal(t, LogIndex(2), vci.pendingAfterSkipLI)

	// Tick resolves it.
	vci.Tick(env.cb)
	require.Contains(t, vci.lis, LogIndex(2))
	require.False(t, vci.hasPendingAfterSkip)
}

// TestVarConsInsts_TickWithNoPendingSkipIsHarmless verifies that Tick does
// nothing special when there is no pending skip.
func TestVarConsInsts_TickWithNoPendingSkipIsHarmless(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci
	vci.lastAnchor = makeAnchor()

	require.False(t, vci.hasPendingAfterSkip)

	// Multiple ticks with nothing pending should not change state.
	for i := 0; i < 5; i++ {
		vci.Tick(env.cb)
	}
	require.Len(t, vci.lis, 1)
	require.Empty(t, env.cbLIs)
}

// TestVarConsInsts_SkipWithAnchorThenAnchorUpdate verifies that receiving
// LatestL1Anchor while a skip is pending (and lastAnchor was already set)
// does NOT prematurely apply the pending skip — only Tick should do that.
func TestVarConsInsts_SkipWithAnchorThenAnchorUpdate(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci
	vci.lastAnchor = makeAnchor()

	// Skip at LI==0.
	vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.True(t, vci.hasPendingAfterSkip)
	require.Empty(t, env.cbLIs)

	// A new L1 anchor arrives. Since the skip-with-anchor path sets
	// pendingAfterSkipLI (not lastLI), LatestL1Anchor's trySet with
	// lastLI==NilLogIndex should be a no-op.
	newAnchor := makeAnchor()
	vci.LatestL1Anchor(newAnchor, env.cb)
	// The pending skip should still be pending — only Tick resolves it.
	require.True(t, vci.hasPendingAfterSkip)
	require.NotContains(t, vci.lis, LogIndex(1))

	// Tick resolves it, now using the updated anchor.
	vci.Tick(env.cb)
	require.Contains(t, vci.lis, LogIndex(1))
	require.Equal(t, LogIndex(1), env.lastCBLI())
	require.False(t, vci.hasPendingAfterSkip)
}

// TestVarConsInsts_SkipNoAnchorThenSkipAgainAfterAnchor verifies the sequence:
// skip with no anchor → anchor arrives (advances) → skip again with anchor →
// tick advances.
func TestVarConsInsts_SkipNoAnchorThenSkipAgainAfterAnchor(t *testing.T) {
	env := newVCITestEnv(t)
	vci := env.vci

	// Skip at LI==0 with no anchor.
	vci.ConsOutputSkip(LogIndex(0), env.cb)
	require.Empty(t, env.cbLIs)

	// Anchor arrives → resolves to LI==1 via LatestL1Anchor path.
	anchor := makeAnchor()
	vci.LatestL1Anchor(anchor, env.cb)
	require.Equal(t, LogIndex(1), env.lastCBLI())

	// Now skip at LI==1 with anchor known → defers to tick.
	cbCountBefore := len(env.cbLIs)
	vci.ConsOutputSkip(LogIndex(1), env.cb)
	require.Equal(t, cbCountBefore, len(env.cbLIs)) // No immediate advance.
	require.True(t, vci.hasPendingAfterSkip)

	// Tick resolves to LI==2.
	vci.Tick(env.cb)
	require.Contains(t, vci.lis, LogIndex(2))
	require.Equal(t, LogIndex(2), env.lastCBLI())
}
