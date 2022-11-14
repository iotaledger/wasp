package mempoolgpa

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

type testPool struct {
	reqs map[isc.RequestID]isc.Request
}

func newTestPool() *testPool {
	return &testPool{
		reqs: make(map[isc.RequestID]isc.Request),
	}
}

func (m *testPool) receiveRequests(reqs ...isc.Request) []bool {
	// mock implementation, only receives 1 request
	req := reqs[0]
	if _, ok := m.reqs[req.ID()]; ok {
		return []bool{false}
	}
	m.reqs[req.ID()] = req
	return []bool{true}
}

func (m *testPool) getRequest(id isc.RequestID) isc.Request {
	return m.reqs[id]
}

func TestGPA(t *testing.T) {
	kp := cryptolib.NewKeyPair()

	log := testlogger.NewLogger(t)

	// Construct the nodes.
	nodes := make(map[gpa.NodeID]gpa.GPA)
	nodeIDs := []gpa.NodeID{gpa.NodeID("1"), gpa.NodeID("2"), gpa.NodeID("3"), gpa.NodeID("4"), gpa.NodeID("5")}
	pools := make(map[gpa.NodeID]*testPool)
	for _, nid := range nodeIDs {
		pool := newTestPool()
		node := New(
			pool.receiveRequests,
			pool.getRequest,
			log,
		)
		pools[nid] = pool
		nodes[nid] = node
		// add everyone as peers
		node.SetPeers(nodeIDs, nil) // no access nodes in this scenario
	}
	tc := gpa.NewTestContext(nodes)

	t.Run("request share", func(t *testing.T) {
		requestA := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), nil, 0).Sign(kp)
		// send requestA to 1 node
		pools[gpa.NodeID("1")].receiveRequests(requestA)
		inputs := make(map[gpa.NodeID]gpa.Input)
		inputs[gpa.NodeID("1")] = requestA
		tc.WithInputs(inputs)
		tc.RunAll()
		// request A should be on 1+shareReqNNodes
		count := 0
		for _, pool := range pools {
			if pool.getRequest(requestA.ID()) != nil {
				count++
			}
		}
		require.Equal(t, 1+shareReqNNodes, count)

		// after a while, all nodes must have the request
		timeTickInputs := make(map[gpa.NodeID]gpa.Input)
		tick := time.Now().Add(1 * time.Second)
		for _, nid := range nodeIDs {
			timeTickInputs[nid] = tick
		}
		tc.WithInputs(timeTickInputs)
		tc.RunAll()
		for nid, pool := range pools {
			require.NotNil(t, pool.getRequest(requestA.ID()), "req not found in node %s", nid)
		}
	})
	t.Run("missing req", func(t *testing.T) {
		requestB := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("foo"), isc.Hn("bar"), nil, 1).Sign(kp)
		ref := isc.RequestRefFromRequest(requestB)
		// add requestB to 2 nodes, send "missing ref" to the other nodes
		pools[gpa.NodeID("1")].receiveRequests(requestB)
		pools[gpa.NodeID("2")].receiveRequests(requestB)

		inputs := make(map[gpa.NodeID]gpa.Input)
		inputs[gpa.NodeID("3")] = ref
		inputs[gpa.NodeID("4")] = ref
		inputs[gpa.NodeID("5")] = ref
		tc.WithInputs(inputs)
		tc.RunAll()

		// after a while, all nodes must have the request
		timeTickInputs := make(map[gpa.NodeID]gpa.Input)
		tick := time.Now().Add(1 * time.Second)
		for _, nid := range nodeIDs {
			timeTickInputs[nid] = tick
		}
		tc.WithInputs(timeTickInputs)
		tc.RunAll()
	})
}
