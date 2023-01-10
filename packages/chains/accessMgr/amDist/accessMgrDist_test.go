// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package amDist_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chains/accessMgr/amDist"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/util"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	t.Run("N=4", func(t *testing.T) { testBasic(t, 4) })
}

func testBasic(t *testing.T, n int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodePubs := testpeers.PublicKeys(peerIdentities)
	nodeIDs := gpa.NodeIDsFromPublicKeys(nodePubs)
	chainID := isc.RandomChainID()

	servers := map[gpa.NodeID][]*cryptolib.PublicKey{}
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nidCopy := nid
		nodes[nid] = amDist.NewAccessMgr(
			gpa.NodeIDFromPublicKey,
			func(ci isc.ChainID, pks []*cryptolib.PublicKey) {
				servers[nidCopy] = pks
			},
			func(pk *cryptolib.PublicKey) {},
			log.Named(fmt.Sprintf("N%v", i)),
		).AsGPA()
	}

	tc := gpa.NewTestContext(nodes)
	for _, nid := range nodeIDs {
		tc.WithInput(nid, amDist.NewInputTrustedNodes(nodePubs))
		tc.WithInput(nid, amDist.NewInputAccessNodes(chainID, nodePubs))
	}
	tc.RunAll()
	for nid := range nodes {
		require.True(t,
			util.Same(nodePubs, servers[nid]),
			"should be same: %v, %v", nodePubs, servers[nid],
		)
	}
}
