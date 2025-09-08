// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist_test

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/chainrunner/accessmanager/dist"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/v2/packages/util"
)

type accessMgrSM struct {
	//
	// These are initialized once.
	initialized     bool
	log             log.Logger
	nodeKeys        []*cryptolib.KeyPair
	nodePubs        []*cryptolib.PublicKey
	nodeIDs         []gpa.NodeID
	genNodeID       *rapid.Generator[gpa.NodeID]
	genNodePub      *rapid.Generator[*cryptolib.PublicKey]
	genNodePubSlice *rapid.Generator[[]*cryptolib.PublicKey]
	//
	// These are set up for each scenario.
	tc      *gpa.TestContext
	nodes   map[gpa.NodeID]gpa.GPA
	servers map[gpa.NodeID][]*cryptolib.PublicKey
	//
	// Model.
	mTrusted map[gpa.NodeID][]*cryptolib.PublicKey
	mActive  map[gpa.NodeID]bool
	mAccess  map[gpa.NodeID][]*cryptolib.PublicKey
}

var _ rapid.StateMachine = &accessMgrSM{}

func newAccessMgrSM(t *rapid.T, nodeCount int) *accessMgrSM {
	sm := new(accessMgrSM)
	if !sm.initialized {
		sm.log = testlogger.NewLogger(t)
		_, sm.nodeKeys = testpeers.SetupKeys(uint16(nodeCount))
		sm.nodePubs = testpeers.PublicKeys(sm.nodeKeys)
		sm.nodeIDs = gpa.NodeIDsFromPublicKeys(sm.nodePubs)
		sm.genNodeID = rapid.SampledFrom(sm.nodeIDs)
		sm.genNodePub = rapid.SampledFrom(sm.nodePubs)
		sm.genNodePubSlice = rapid.SliceOfDistinct(
			sm.genNodePub,
			func(pub *cryptolib.PublicKey) cryptolib.PublicKeyKey { return pub.AsKey() },
		)
		sm.initialized = true
	}

	sm.servers = map[gpa.NodeID][]*cryptolib.PublicKey{}
	sm.nodes = map[gpa.NodeID]gpa.GPA{}
	for _, nid := range sm.nodeIDs {
		sm.servers[nid] = []*cryptolib.PublicKey{}
		nidCopy := nid
		sm.nodes[nid] = dist.NewAccessMgr(
			gpa.NodeIDFromPublicKey,
			func(servers []*cryptolib.PublicKey) {
				t.Logf("serversUpdatedCB: nodeID=%v, servers=%v", nidCopy, servers)
				sm.servers[nidCopy] = servers
			},
			func(pk *cryptolib.PublicKey) {},
			sm.log.NewChildLogger(nid.ShortString()),
		).AsGPA()
	}
	sm.tc = gpa.NewTestContext(sm.nodes)

	sm.mTrusted = map[gpa.NodeID][]*cryptolib.PublicKey{}
	sm.mActive = map[gpa.NodeID]bool{}
	sm.mAccess = map[gpa.NodeID][]*cryptolib.PublicKey{}
	for _, nid := range sm.nodeIDs {
		sm.mTrusted[nid] = []*cryptolib.PublicKey{}
		sm.mAccess[nid] = []*cryptolib.PublicKey{}
		sm.mAccess[nid] = []*cryptolib.PublicKey{}
		sm.mActive[nid] = false
	}
	return sm
}

func (sm *accessMgrSM) InputTrustedNodes(t *rapid.T) {
	nodeID := sm.genNodeID.Draw(t, "nodeID")
	trustedNodes := sm.genNodePubSlice.Draw(t, "trustedNodes")
	sm.tc.WithInput(nodeID, dist.NewInputTrustedNodes(trustedNodes)).RunAll()
	sm.mTrusted[nodeID] = trustedNodes
}

func (sm *accessMgrSM) InputAccessNodes(t *rapid.T) {
	nodeID := sm.genNodeID.Draw(t, "nodeID")
	accessNodes := sm.genNodePubSlice.Draw(t, "accessNodes")
	sm.tc.WithInput(nodeID, dist.NewInputAccessNodes(accessNodes)).RunAll()
	sm.mActive[nodeID] = true
	sm.mAccess[nodeID] = accessNodes
}

func (sm *accessMgrSM) InputChainDisabled(t *rapid.T) {
	nodeID := sm.genNodeID.Draw(t, "nodeID")
	sm.tc.WithInput(nodeID, dist.NewInputChainDisabled()).RunAll()
	sm.mActive[nodeID] = false
}

func (sm *accessMgrSM) Reboot(t *rapid.T) {
	nodeID := sm.genNodeID.Draw(t, "nodeID")
	//
	// Just recreate a node.
	sm.nodes[nodeID] = dist.NewAccessMgr(
		gpa.NodeIDFromPublicKey,
		func(servers []*cryptolib.PublicKey) {
			t.Logf("serversUpdatedCB: nodeID=%v, servers=%v", nodeID, servers)
			sm.servers[nodeID] = servers
		},
		func(pk *cryptolib.PublicKey) {},
		sm.log.NewChildLogger(nodeID.ShortString()),
	).AsGPA()
	//
	// Re-initialize all the persistent info: access information, active chains, trusted nodes.
	// But the servers are not restored here. The algorithm has to restore that.
	sm.tc.WithInput(nodeID, dist.NewInputTrustedNodes(sm.mTrusted[nodeID]))
	if sm.mActive[nodeID] {
		sm.tc.WithInput(nodeID, dist.NewInputAccessNodes(sm.mAccess[nodeID]))
	}
	sm.tc.RunAll()
}

func (sm *accessMgrSM) Check(t *rapid.T) {
	for _, nodePub := range sm.nodePubs {
		nodeID := gpa.NodeIDFromPublicKey(nodePub)
		shouldBeServers := []*cryptolib.PublicKey{}
		for _, peerPub := range sm.nodePubs {
			peerID := gpa.NodeIDFromPublicKey(peerPub)
			if sm.mActive[nodeID] &&
				sm.mActive[peerID] &&
				lo.Contains(sm.mTrusted[peerID], nodePub) &&
				lo.Contains(sm.mTrusted[nodeID], peerPub) &&
				lo.Contains(sm.mAccess[peerID], nodePub) {
				shouldBeServers = append(shouldBeServers, peerPub)
			}
		}
		require.True(t,
			util.Same(sm.servers[nodeID], shouldBeServers),
			"nodeID=%v, chainID=%v, have=%v, expect=%v",
			nodeID, sm.servers[nodeID], shouldBeServers,
		)
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestAccessMgrRapid(t *testing.T) {
	tests := []struct {
		n int
	}{
		{n: 1},
		{n: 2},
		{n: 4},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("N%d", test.n), rapid.MakeCheck(func(t *rapid.T) {
			sm := newAccessMgrSM(t, test.n)
			t.Repeat(rapid.StateMachineActions(sm))
		}))
	}
}
