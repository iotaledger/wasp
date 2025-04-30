// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

// TODO: Test should involve suspend/resume.

func TestCmtLogBasic(t *testing.T) {
	type test struct {
		n int
		f int
	}
	tests := []test{
		{n: 4, f: 1},
	}
	for _, tst := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v", tst.n, tst.f),
			func(tt *testing.T) { testCmtLogBasic(tt, tst.n, tst.f) })
	}
}

func testCmtLogBasic(t *testing.T, n, f int) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	//
	// Chain identifiers.
	aliasRef := iotatest.RandomObjectRef()
	chainID := isc.ChainIDFromObjectID(*aliasRef.ObjectID)
	//
	// Node identities.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	peerPubKeys := testpeers.PublicKeys(peerIdentities)
	//
	// Committee.
	committeeAddress, committeeKeyShares := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Construct the algorithm nodes.
	gpaNodeIDs := gpa.NodeIDsFromPublicKeys(peerPubKeys)
	gpaNodes := map[gpa.NodeID]gpa.GPA{}
	for i := range gpaNodeIDs {
		dkShare, err := committeeKeyShares[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		consensusStateRegistry := testutil.NewConsensusStateRegistry() // Empty store in this case.
		cmtLogInst, err := cmtlog.New(gpaNodeIDs[i], chainID, dkShare, consensusStateRegistry, gpa.NodeIDFromPublicKey, true, -1, nil, log.NewChildLogger(fmt.Sprintf("N%v", i)))
		require.NoError(t, err)
		gpaNodes[gpaNodeIDs[i]] = cmtLogInst.AsGPA()
	}
	gpaTC := gpa.NewTestContext(gpaNodes)
	//
	// Start the algorithms.
	gpaTC.RunAll()
	gpaTC.PrintAllStatusStrings("Initial", t.Logf)
	//
	// Provide first alias output. Consensus should be sent now.
	// FIXME is should be anchor state transition, instead of random anchor
	ao1 := randomAnchorWithID(*aliasRef.ObjectID, committeeAddress, 1)
	t.Logf("AO1=%v", ao1)
	gpaTC.WithInputs(inputAnchorConfirmed(gpaNodes, ao1)).RunAll()
	gpaTC.PrintAllStatusStrings("After AO1Recv", t.Logf)
	cons1 := gpaNodes[gpaNodeIDs[0]].Output().(cmtlog.Output)
	cons1Outs := map[gpa.NodeID]cmtlog.Output{}
	for nid, n := range gpaNodes {
		require.NotNil(t, n.Output())
		require.Equal(t, cons1, n.Output())
		cons1Outs[nid] = n.Output().(cmtlog.Output)
		require.Equal(t, 1, len(cons1Outs[nid]))
		require.Nil(t, cons1Outs[nid][cmtlog.LogIndex(1)])
	}
	//
	// Consensus results received (consumed ao1, produced ao2).
	// FIXME is should be anchor state transition, instead of random anchor
	ao2 := randomAnchorWithID(*aliasRef.ObjectID, committeeAddress, 2)
	t.Logf("AO2=%v", ao2)
	gpaTC.WithInputs(inputConsensusOutput(cons1Outs, ao2)).RunAll()
	gpaTC.PrintAllStatusStrings("After gpaMsgsAO2Cons", t.Logf)
	cons2 := gpaNodes[gpaNodeIDs[0]].Output().(cmtlog.Output)
	t.Logf("cons2=%v", cons2)
	for _, n := range gpaNodes {
		out := n.Output().(cmtlog.Output)
		require.NotNil(t, out)
		require.Equal(t, cons2, out)
		require.Equal(t, 2, len(out))
		require.Nil(t, out[cmtlog.LogIndex(1)])
		require.Equal(t, ao2, out[cmtlog.LogIndex(2)])
	}
	//
	// AO Confirmed received (nothing changes, we are ahead of it)
	gpaTC.WithInputs(inputAnchorConfirmed(gpaNodes, ao2)).RunAll()
	gpaTC.PrintAllStatusStrings("After gpaMsgsAO2Recv", t.Logf)
	for _, n := range gpaNodes {
		require.NotNil(t, n.Output())
		require.Equal(t, cons2, n.Output())
	}
}

////////////////////////////////////////////////////////////////////////////////
// Helper functions.

func inputAnchorConfirmed(gpaNodes map[gpa.NodeID]gpa.GPA, ao *isc.StateAnchor) map[gpa.NodeID]gpa.Input {
	inputs := map[gpa.NodeID]gpa.Input{}
	for n := range gpaNodes {
		inputs[n] = cmtlog.NewInputAnchorConfirmed(ao)
	}
	return inputs
}

func inputConsensusOutput(consReq map[gpa.NodeID]cmtlog.Output, nextAO *isc.StateAnchor) map[gpa.NodeID]gpa.Input {
	inputs := map[gpa.NodeID]gpa.Input{}
	for nid, outs := range consReq {
		maxLI := cmtlog.NilLogIndex()
		for li := range outs {
			if li <= maxLI {
				break
			}
			maxLI = li
			inputs[nid] = cmtlog.NewInputConsensusOutputConfirmed(nextAO, li)
		}
	}
	return inputs
}

func randomAnchorWithID(anchorID iotago.ObjectID, stateAddress *cryptolib.Address, stateIndex uint32) *isc.StateAnchor {
	anchor := iscmovetest.RandomAnchor(iscmovetest.RandomAnchorOption{StateMetadata: &[]byte{}, StateIndex: &stateIndex, ID: &anchorID})
	stateAnchor := isc.NewStateAnchor(
		&iscmove.AnchorWithRef{
			Object:    &anchor,
			ObjectRef: *iotatest.RandomObjectRef(),
			Owner:     stateAddress.AsIotaAddress(),
		}, *cryptolib.NewRandomAddress().AsIotaAddress())

	return &stateAnchor
}
