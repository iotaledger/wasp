// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/chain/mempool/distsync"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestBasic(t *testing.T) {
	t.Run("N=20, cmtN=10, cmtF=3", func(t *testing.T) { testBasic(t, 20, 10, 3) })
}

func testBasic(t *testing.T, n, cmtN, cmtF int) {
	require.GreaterOrEqual(t, n, cmtN)
	log := testlogger.NewLogger(t)
	kp := cryptolib.NewKeyPair()

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
		requestReceivedCB := func(req isc.Request) bool {
			_, have := recv[thisNodeID]
			recv[thisNodeID] = req
			return !have
		}
		nodes[nid] = distsync.New(thisNodeID, requestNeededCB, requestReceivedCB, 100, func(count int) {}, log)
	}

	req := isc.NewOffLedgerRequest(
		isctest.RandomChainID(),
		isc.NewMessage(isc.Hn("foo"), isc.Hn("bar"), nil),
		0,
		gas.LimitsDefault.MaxGasPerRequest,
	).Sign(kp)
	reqRef := isc.RequestRefFromRequest(req)
	cmtNodes := []gpa.NodeID{} // Random subset of all nodes.
	for pos, idx := range rand.Perm(cmtN) {
		if pos >= cmtN {
			break
		}
		cmtNodes = append(cmtNodes, nodeIDs[idx])
	}

	tc := gpa.NewTestContext(nodes)
	//
	// Setup the committee for all nodes.
	for _, nid := range nodeIDs {
		tc.WithInput(nid, distsync.NewInputServerNodes(cmtNodes, cmtNodes))
	}
	//
	// Send a request to a single node.
	tc.WithInput(nodeIDs[rand.Intn(n)], distsync.NewInputPublishRequest(req))
	tc.RunAll()
	require.GreaterOrEqual(t, len(recv), cmtF+1)
	//
	// All nodes asks for the req.
	ctx := context.Background()
	for _, nid := range nodeIDs {
		tc.WithInput(nid, distsync.NewInputRequestNeeded(ctx, reqRef))
	}
	tc.RunAll()
	require.Equal(t, len(recv), n)
	//
	// Some time ticks (just to check if not crashes.)
	for _, nid := range nodeIDs {
		tc.WithInput(nid, distsync.NewInputTimeTick())
	}
	tc.RunAll()
}
