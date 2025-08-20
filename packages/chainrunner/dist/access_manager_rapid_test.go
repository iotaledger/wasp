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
	"github.com/iotaledger/wasp/v2/packages/chains/accessmanager/dist"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
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
	chainIDs        []isc.ChainID
	genNodeID       *rapid.Generator[gpa.NodeID]
	genNodePub      *rapid.Generator[*cryptolib.PublicKey]
	genNodePubSlice *rapid.Generator[[]*cryptolib.PublicKey]
	genChainID      *rapid.Generator[isc.ChainID]
	//
	// These are set up for each scenario.
	tc      *gpa.TestContext
	nodes   map[gpa.NodeID]gpa.GPA
	servers map[gpa.NodeID]map[isc.ChainID][]*cryptolib.PublicKey
	//
	// Model.
	mTrusted map[gpa.NodeID][]*cryptolib.PublicKey
	mActive  map[gpa.NodeID]map[isc.ChainID]bool
	mAccess  map[gpa.NodeID]map[isc.ChainID][]*cryptolib.PublicKey
}

var _ rapid.StateMachine = &accessMgrSM{}

func newAccessMgrSM(t *rapid.T, nodeCount, chainCount int) *accessMgrSM {
	sm := new(accessMgrSM)
	if !sm.initialized {
		sm.log = testlogger.NewLogger(t)
		_, sm.nodeKeys = testpeers.SetupKeys(uint16(nodeCount))
		sm.nodePubs = testpeers.PublicKeys(sm.nodeKeys)
		sm.nodeIDs = gpa.NodeIDsFromPublicKeys(sm.nodePubs)
		sm.chainIDs = make([]isc.ChainID, chainCount)
		for i := range sm.chainIDs {
			sm.chainIDs[i] = isctest.RandomChainID([]byte{byte(i)})
		}
		sm.genNodeID = rapid.SampledFrom(sm.nodeIDs)
		sm.genNodePub = rapid.SampledFrom(sm.nodePubs)
		sm.genNodePubSlice = rapid.SliceOfDistinct(
			sm.genNodePub,
			func(pub *cryptolib.PublicKey) cryptolib.PublicKeyKey { return pub.AsKey() },
		)
		sm.genChainID = rapid.SampledFrom(sm.chainIDs)
		sm.initialized = true
	}

	sm.servers = map[gpa.NodeID]map[isc.ChainID][]*cryptolib.PublicKey{}
	sm.nodes = map[gpa.NodeID]gpa.GPA{}
	for _, nid := range sm.nodeIDs {
		sm.servers[nid] = map[isc.ChainID][]*cryptolib.PublicKey{}
		for _, chainID := range sm.chainIDs {
			sm.servers[nid][chainID] = []*cryptolib.PublicKey{}
		}
		nidCopy := nid
		sm.nodes[nid] = dist.NewAccessMgr(
			gpa.NodeIDFromPublicKey,
			func(chainID isc.ChainID, servers []*cryptolib.PublicKey) {
				t.Logf("serversUpdatedCB: nodeID=%v, chainID=%v, servers=%v", nidCopy, chainID, servers)
				sm.servers[nidCopy][chainID] = servers
			},
			func(pk *cryptolib.PublicKey) {},
			sm.log.NewChildLogger(nid.ShortString()),
		).AsGPA()
	}
	sm.tc = gpa.NewTestContext(sm.nodes)

	sm.mTrusted = map[gpa.NodeID][]*cryptolib.PublicKey{}
	sm.mActive = map[gpa.NodeID]map[isc.ChainID]bool{}
	sm.mAccess = map[gpa.NodeID]map[isc.ChainID][]*cryptolib.PublicKey{}
	for _, nid := range sm.nodeIDs {
		sm.mTrusted[nid] = []*cryptolib.PublicKey{}
		sm.mAccess[nid] = map[isc.ChainID][]*cryptolib.PublicKey{}
		sm.mActive[nid] = map[isc.ChainID]bool{}
		for _, ch := range sm.chainIDs {
			sm.mAccess[nid][ch] = []*cryptolib.PublicKey{}
			sm.mActive[nid][ch] = false
		}
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
	chainID := sm.genChainID.Draw(t, "chainID")
	accessNodes := sm.genNodePubSlice.Draw(t, "accessNodes")
	sm.tc.WithInput(nodeID, dist.NewInputAccessNodes(chainID, accessNodes)).RunAll()
	sm.mActive[nodeID][chainID] = true
	sm.mAccess[nodeID][chainID] = accessNodes
}

func (sm *accessMgrSM) InputChainDisabled(t *rapid.T) {
	nodeID := sm.genNodeID.Draw(t, "nodeID")
	chainID := sm.genChainID.Draw(t, "chainID")
	sm.tc.WithInput(nodeID, dist.NewInputChainDisabled(chainID)).RunAll()
	sm.mActive[nodeID][chainID] = false
}

func (sm *accessMgrSM) Reboot(t *rapid.T) {
	nodeID := sm.genNodeID.Draw(t, "nodeID")
	//
	// Just recreate a node.
	sm.nodes[nodeID] = dist.NewAccessMgr(
		gpa.NodeIDFromPublicKey,
		func(chainID isc.ChainID, servers []*cryptolib.PublicKey) {
			t.Logf("serversUpdatedCB: nodeID=%v, chainID=%v, servers=%v", nodeID, chainID, servers)
			sm.servers[nodeID][chainID] = servers
		},
		func(pk *cryptolib.PublicKey) {},
		sm.log.NewChildLogger(nodeID.ShortString()),
	).AsGPA()
	//
	// Re-initialize all the persistent info: access information, active chains, trusted nodes.
	// But the servers are not restored here. The algorithm has to restore that.
	sm.tc.WithInput(nodeID, dist.NewInputTrustedNodes(sm.mTrusted[nodeID]))
	for _, chainID := range sm.chainIDs {
		if !sm.mActive[nodeID][chainID] {
			continue
		}
		sm.tc.WithInput(nodeID, dist.NewInputAccessNodes(chainID, sm.mAccess[nodeID][chainID]))
	}
	sm.tc.RunAll()
}

func (sm *accessMgrSM) Check(t *rapid.T) {
	for _, nodePub := range sm.nodePubs {
		nodeID := gpa.NodeIDFromPublicKey(nodePub)
		for _, chainID := range sm.chainIDs {
			shouldBeServers := []*cryptolib.PublicKey{}
			for _, peerPub := range sm.nodePubs {
				peerID := gpa.NodeIDFromPublicKey(peerPub)
				if sm.mActive[nodeID][chainID] &&
					sm.mActive[peerID][chainID] &&
					lo.Contains(sm.mTrusted[peerID], nodePub) &&
					lo.Contains(sm.mTrusted[nodeID], peerPub) &&
					lo.Contains(sm.mAccess[peerID][chainID], nodePub) {
					shouldBeServers = append(shouldBeServers, peerPub)
				}
			}
			require.True(t,
				util.Same(sm.servers[nodeID][chainID], shouldBeServers),
				"nodeID=%v, chainID=%v, have=%v, expect=%v",
				nodeID, chainID, sm.servers[nodeID][chainID], shouldBeServers,
			)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestAccessMgrRapid(t *testing.T) {
	tests := []struct {
		n int
		c int
	}{
		{n: 1, c: 1},
		{n: 2, c: 1},
		{n: 4, c: 3},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("N%d C%d", test.n, test.c), rapid.MakeCheck(func(t *rapid.T) {
			sm := newAccessMgrSM(t, test.n, test.c)
			t.Repeat(rapid.StateMachineActions(sm))
		}))
	}
}
