// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acss"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

// In this test all the nodes are actually fair.
func TestBasic(t *testing.T) {
	t.Parallel()
	t.Run("n=1,f=0", func(tt *testing.T) { genericTest(tt, 1, 0, 0, 0) })
	t.Run("n=2,f=0", func(tt *testing.T) { genericTest(tt, 2, 0, 0, 0) })
	t.Run("n=3,f=0", func(tt *testing.T) { genericTest(tt, 3, 0, 0, 0) })
	t.Run("n=4,f=1", func(tt *testing.T) { genericTest(tt, 4, 1, 0, 0) })
	t.Run("n=10,f=3", func(tt *testing.T) { genericTest(tt, 10, 3, 0, 0) })
	t.Run("n=31,f=10", func(tt *testing.T) { genericTest(tt, 31, 10, 0, 0) })
}

func TestSilentPeers(t *testing.T) {
	t.Parallel()
	t.Run("n=4,f=1", func(tt *testing.T) { genericTest(tt, 4, 1, 1, 0) })
	t.Run("n=10,f=3", func(tt *testing.T) { genericTest(tt, 10, 3, 3, 0) })
	t.Run("n=31,f=10", func(tt *testing.T) { genericTest(tt, 31, 10, 10, 0) })
}

func TestFaultyDealer(t *testing.T) {
	t.Parallel()
	t.Run("n=4,f=1,F=0,D=1", func(tt *testing.T) { genericTest(tt, 4, 1, 0, 1) })
	t.Run("n=10,f=3,F=2,D=1", func(tt *testing.T) { genericTest(tt, 10, 3, 2, 1) })
	t.Run("n=31,f=10,F=5,D=5", func(tt *testing.T) { genericTest(tt, 31, 10, 5, 5) })
	t.Run("n=31,f=10,F=0,D=10", func(tt *testing.T) { genericTest(tt, 31, 10, 0, 10) })
}

func genericTest(
	t *testing.T,
	n int, // Number of nodes.
	f int, // Max number of faulty nodes.
	silentNodes int, // Number of actually faulty nodes (by not responding to anything).
	faultyDeals int, // How many faulty deals the dealer produces?
) {
	require.True(t, silentNodes+faultyDeals <= f) // Assert tests are within assumptions.
	log := testlogger.WithLevel(testlogger.NewLogger(t), logger.LevelWarn, false)
	defer log.Sync()
	suite := tcrypto.DefaultEd25519Suite()
	secretToShare := suite.Scalar().Pick(suite.RandomStream())
	//
	// Setup keys and node names.
	nodeIDs := gpa.MakeTestNodeIDs("node", n)
	nodeSKs := map[gpa.NodeID]kyber.Scalar{}
	nodePKs := map[gpa.NodeID]kyber.Point{}
	for i := range nodeIDs {
		nodeSKs[nodeIDs[i]] = suite.Scalar().Pick(suite.RandomStream())
		nodePKs[nodeIDs[i]] = suite.Point().Mul(nodeSKs[nodeIDs[i]], nil)
	}
	dealer := nodeIDs[rand.Intn(len(nodeIDs))]
	dealCB := func(i int, e []byte) []byte {
		if silentNodes <= i && i < silentNodes+faultyDeals {
			log.Infof("Corrupting the deal for node[%v]=%v", i, nodeIDs[i])
			e[0] ^= 0xff // Corrupt the data.
		}
		return e
	}
	faulty := nodeIDs[:silentNodes]
	nodes := map[gpa.NodeID]gpa.GPA{}
	for _, nid := range nodeIDs {
		nodes[nid] = acss.New(suite, nodeIDs, nodePKs, f, nid, nodeSKs[nid], dealer, dealCB, log.Named(string(nid)))
		if isNodeInList(nid, faulty) {
			nodes[nid] = &silentNode{nested: nodes[nid]}
		}
	}
	gpa.NewTestContext(nodes).WithInputs(map[gpa.NodeID]gpa.Input{dealer: secretToShare}).RunAll()
	outPriShares := []*share.PriShare{}
	for i, n := range nodes {
		o := n.Output()
		if !isNodeInList(i, faulty) {
			require.NotNil(t, o)
			require.NotNil(t, o.(*acss.Output).PriShare)
			require.NotNil(t, o.(*acss.Output).Commits)
			outPriShares = append(outPriShares, o.(*acss.Output).PriShare)
		}
	}
	outSecret, err := share.RecoverSecret(suite, outPriShares, f+1, n)
	require.NoError(t, err)
	require.True(t, outSecret.Equal(secretToShare))
}

func isNodeInList(n gpa.NodeID, list []gpa.NodeID) bool {
	for i := range list {
		if list[i] == n {
			return true
		}
	}
	return false
}

// silent node don't respond to any messages.
// If it is the dealer, if performs the initial share.
type silentNode struct {
	nested gpa.GPA
}

var _ gpa.GPA = &silentNode{}

func (s *silentNode) Input(input gpa.Input) gpa.OutMessages {
	// Return the messages, if that's a dealer, otherwise the execution is not meaningful.
	return s.nested.Input(input)
}

func (s *silentNode) Message(msg gpa.Message) gpa.OutMessages {
	// Just drop all the received messages.
	return nil
}

func (s *silentNode) Output() gpa.Output {
	return s.nested.Output()
}

func (s *silentNode) StatusString() string {
	return "{silentNode}"
}

func (s *silentNode) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return s.nested.UnmarshalMessage(data)
}
