// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/aba/mostefaoui"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	// Basic tests
	t.Run("N=1,F=0,I=rand", func(tt *testing.T) { testBasic(tt, 1, 0, "rand", 0) })
	t.Run("N=2,F=0,I=rand", func(tt *testing.T) { testBasic(tt, 2, 0, "rand", 0) })
	t.Run("N=3,F=0,I=rand", func(tt *testing.T) { testBasic(tt, 3, 0, "rand", 0) })
	t.Run("N=4,F=1,I=rand", func(tt *testing.T) { testBasic(tt, 4, 1, "rand", 0) })
	t.Run("N=10,F=3,I=rand", func(tt *testing.T) { testBasic(tt, 10, 3, "rand", 0) })
	t.Run("N=31,F=10,I=rand", func(tt *testing.T) { testBasic(tt, 31, 10, "rand", 0) })
	//
	// Uniform inputs.
	t.Run("N=1,F=0,I=true", func(tt *testing.T) { testBasic(tt, 1, 0, "true", 0) })
	t.Run("N=1,F=0,I=false", func(tt *testing.T) { testBasic(tt, 1, 0, "false", 0) })
	t.Run("N=2,F=0,I=true", func(tt *testing.T) { testBasic(tt, 2, 0, "true", 0) })
	t.Run("N=2,F=0,I=false", func(tt *testing.T) { testBasic(tt, 2, 0, "false", 0) })
	t.Run("N=3,F=0,I=true", func(tt *testing.T) { testBasic(tt, 3, 0, "true", 0) })
	t.Run("N=3,F=0,I=false", func(tt *testing.T) { testBasic(tt, 3, 0, "false", 0) })
	t.Run("N=4,F=1,I=true", func(tt *testing.T) { testBasic(tt, 4, 1, "true", 0) })
	t.Run("N=4,F=1,I=false", func(tt *testing.T) { testBasic(tt, 4, 1, "false", 0) })
	//
	// Silent nodes.
	t.Run("N=4,F=1,I=rand,S=1", func(tt *testing.T) { testBasic(tt, 4, 1, "rand", 1) })
	t.Run("N=10,F=3,I=rand,S=3", func(tt *testing.T) { testBasic(tt, 10, 3, "rand", 3) })
	t.Run("N=31,F=10,I=rand,S=10", func(tt *testing.T) { testBasic(tt, 31, 10, "rand", 10) })
}

func testBasic(t *testing.T, n, f int, inpType string, silent int) {
	t.Parallel()
	threshold := f + 1
	// Infra and stuff for CC.
	log := testlogger.NewLogger(t)
	suite := tcrypto.DefaultBLSSuite()
	_, commits, priShares := testpeers.MakeSharedSecret(suite, n, threshold)
	//
	// Create the nodes.
	nodeIDs := gpa.MakeTestNodeIDs(n)
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		if i >= n-silent {
			nodes[nid] = gpa.MakeTestSilentNode()
		} else {
			nodeLog := log.Named(nid.ShortString())
			ii := i
			makeCCInst := func(round int) gpa.GPA {
				realCC := blssig.New(
					suite, nodeIDs, commits, priShares[ii], threshold,
					nodeIDs[ii], []byte{1, 2, 3, byte(round)}, nodeLog,
				)
				return semi.New(round, realCC)
			}
			nodes[nid] = mostefaoui.New(nodeIDs, nid, f, makeCCInst, nodeLog).AsGPA()
		}
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Choose inputs.
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		switch inpType {
		case "rand":
			inputs[nid] = rand.Int()%2 == 1
		case "true":
			inputs[nid] = true
		case "false":
			inputs[nid] = false
		default:
			panic("unexpected input type")
		}
	}
	t.Logf("Inputs: %v", inputs)
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("Done,", t.Logf)
	//
	out0 := nodes[nodeIDs[0]].Output().(*mostefaoui.Output)
	for i, nid := range nodeIDs {
		if i >= n-silent {
			continue
		}
		out := nodes[nid].Output().(*mostefaoui.Output)
		require.NotNil(t, out)
		require.True(t, out.Terminated)
		switch inpType {
		case "rand":
			require.Equal(t, out0.Value, out.Value)
		case "true":
			require.Equal(t, true, out.Value)
		case "false":
			require.Equal(t, false, out.Value)
		default:
			panic("unexpected input type")
		}
	}
}
