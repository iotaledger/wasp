// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/mempool/distSync"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestBasic(t *testing.T) {
	t.Run("N=20, cmtN=10, cmtF=3", func(t *testing.T) { testBasic(t, 20, 10, 3) })
}

func testBasic(t *testing.T, n, cmtN, cmtF int) {
	require.GreaterOrEqual(t, n, cmtN)
	rand.Seed(time.Now().UnixNano())
	log := testlogger.NewLogger(t)
	kp := cryptolib.NewKeyPair()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	recv := map[gpa.NodeID]isc.Request{}
	nodeIDs := gpa.MakeTestNodeIDs(n)
	nodes := map[gpa.NodeID]gpa.GPA{}
	for _, nid := range nodeIDs {
		thisNodeID := nid
		requestNeededCB := func(*isc.RequestRef) isc.Request {
			if req, ok := recv[thisNodeID]; ok {
				return req
			}
			return nil
		}
		requestReceivedCB := func(req isc.Request) {
			recv[thisNodeID] = req
		}
		nodes[nid] = distSync.New(thisNodeID, requestNeededCB, requestReceivedCB, 100, log)
	}

	req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), nil, 0).Sign(kp)
	reqRef := isc.RequestRefFromRequest(req)
	cmtNodes := []gpa.NodeID{} // Random subset of all nodes.
	for pos, idx := range rnd.Perm(cmtN) {
		if pos >= cmtN {
			break
		}
		cmtNodes = append(cmtNodes, nodeIDs[idx])
	}

	tc := gpa.NewTestContext(nodes)
	//
	// Setup the committee for all nodes.
	for _, nid := range nodeIDs {
		tc.WithInput(nid, distSync.NewInputServerNodes(cmtNodes, cmtNodes))
	}
	//
	// Send a request to a single node.
	tc.WithInput(nodeIDs[rand.Intn(n)], distSync.NewInputPublishRequest(req))
	tc.RunAll()
	require.GreaterOrEqual(t, len(recv), cmtF+1)
	//
	// All nodes asks for the req.
	for _, nid := range nodeIDs {
		tc.WithInput(nid, distSync.NewInputRequestNeeded(reqRef, true))
	}
	tc.RunAll()
	require.Equal(t, len(recv), n)
	//
	// Some time ticks (just to check if not crashes.)
	for _, nid := range nodeIDs {
		tc.WithInput(nid, distSync.NewInputTimeTick())
	}
	tc.RunAll()
}
