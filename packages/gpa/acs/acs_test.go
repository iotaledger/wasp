// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acs_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acs"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	// Basic tests
	t.Run("N=1,F=0", func(tt *testing.T) { testBasic(tt, 1, 0, 0) })
	t.Run("N=2,F=0", func(tt *testing.T) { testBasic(tt, 2, 0, 0) })
	t.Run("N=3,F=0", func(tt *testing.T) { testBasic(tt, 3, 0, 0) })
	t.Run("N=4,F=1", func(tt *testing.T) { testBasic(tt, 4, 1, 0) })
	t.Run("N=10,F=3", func(tt *testing.T) { testBasic(tt, 10, 3, 0) })
	t.Run("N=31,F=10", func(tt *testing.T) { testBasic(tt, 31, 10, 0) })
	//
	// Silent nodes.
	t.Run("N=4,F=1,S=1", func(tt *testing.T) { testBasic(tt, 4, 1, 1) })
	t.Run("N=10,F=3,S=3", func(tt *testing.T) { testBasic(tt, 10, 3, 3) })
	t.Run("N=31,F=10,S=10", func(tt *testing.T) { testBasic(tt, 31, 10, 10) })
}

func testBasic(t *testing.T, n, f, silent int) {
	t.Parallel()
	ccThreshold := f + 1
	//
	// Infra and stuff for CC.
	log := testlogger.NewLogger(t)
	suite := tcrypto.DefaultBLSSuite()
	_, commits, priShares := testpeers.MakeSharedSecret(suite, n, ccThreshold)
	//
	// Create the nodes.
	nodeIDs := gpa.MakeTestNodeIDs("acs", n)
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		if i >= n-silent {
			nodes[nid] = gpa.MakeTestSilentNode()
		} else {
			nodeLog := log.Named(string(nid))
			ii := i
			makeCCInstFun := func(nodeID gpa.NodeID, round int) gpa.GPA {
				sid := fmt.Sprintf("%s-%v", nodeID, round)
				realCC := blssig.New(
					suite, nodeIDs, commits, priShares[ii], ccThreshold,
					nodeIDs[ii], []byte(sid), nodeLog,
				)
				return semi.New(round, realCC)
			}
			nodes[nid] = acs.New(nodeIDs, nid, f, makeCCInstFun, nodeLog).AsGPA()
		}
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Choose inputs.
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		inputs[nid] = []byte(fmt.Sprintf("%v-input", nid))
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("Done,", t.Logf)
	//
	out0 := nodes[nodeIDs[0]].Output().(*acs.Output)
	for i, nid := range nodeIDs {
		if i >= n-silent {
			continue
		}
		out := nodes[nid].Output().(*acs.Output)
		require.NotNil(t, out)
		require.True(t, out.Terminated)
		require.Equal(t, out0.Values, out.Values)
	}
}
