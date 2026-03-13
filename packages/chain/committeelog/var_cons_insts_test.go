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

// TestVarConsInsts_ConsOutputSkipDefersUntilTick verifies that after a SKIP
// decision, advancing to the next log index is deferred until Tick is called.
func TestVarConsInsts_ConsOutputSkipDefersUntilTick(t *testing.T) {
	log := testlogger.NewLogger(t)
	minLI := LogIndex(0)

	persisted := []LogIndex{}
	var lastTickLI LogIndex

	vci := NewVarConsInsts(
		minLI,
		func(li LogIndex) {
			persisted = append(persisted, li)
		},
		func(Output) {},
		log,
	)

	// Simulate that we have an L1 anchor already known.
	anchor := isc.NewStateAnchor(&iscmove.AnchorWithRef{}, *iotago.MustAddressFromHex("0x0"))
	vci.lastAnchor = &anchor

	// Before SKIP, only minLI should be present.
	require.Len(t, vci.lis, 1)
	require.Contains(t, vci.lis, minLI)

	// Call ConsOutputSkip at LI==0. This should NOT immediately advance to LI==1.
	out := vci.ConsOutputSkip(minLI, func(li LogIndex) gpa.OutMessages {
		lastTickLI = li
		return gpa.NoMessages()
	})
	require.Nil(t, out)

	// Still only minLI should be present; next LI should not be proposed yet.
	require.Len(t, vci.lis, 1)
	require.Contains(t, vci.lis, minLI)

	// Now simulate a Tick; this should apply the pending advancement to LI==1.
	out = vci.Tick(func(li LogIndex) gpa.OutMessages {
		lastTickLI = li
		return gpa.NoMessages()
	})
	require.NotNil(t, out)

	// LI==1 should now be in the map and persisted, and the callback should see LI==1.
	require.Contains(t, vci.lis, LogIndex(1))
	require.Equal(t, LogIndex(1), lastTickLI)
	require.Contains(t, persisted, LogIndex(1))
}
