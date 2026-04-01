package batchproposal_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/packages/chain/consensus/batchproposal"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/util"
)

func makeSharedGasCoin() *coin.CoinWithRef {
	return &coin.CoinWithRef{
		Type:  coin.BaseTokenType,
		Value: coin.Value(100),
		Ref:   iotatest.RandomObjectRef(),
	}
}

// makeValidProposal creates a fully populated BatchProposal for the given node
// index. All proposals must share the same gas coin for F+1 agreement.
func makeValidProposal(
	nodeIndex uint16,
	n int, //nolint:unparam
	anchor *isc.StateAnchor,
	reqRefs []*isc.RequestRef,
	gasCoin *coin.CoinWithRef,
) *batchproposal.BatchProposal {
	return batchproposal.NewBatchProposal(
		nodeIndex,
		anchor,
		util.NewFixedSizeBitVector(uint16(n)).SetBits([]int{int(nodeIndex)}),
		nil,
		time.Now(),
		isctest.NewRandomAgentID(),
		reqRefs,
		[]*coin.CoinWithRef{gasCoin},
		parameterstest.L1Mock,
	)
}

func makeRequestRefs(count int) []*isc.RequestRef {
	refs := make([]*isc.RequestRef, count)
	for i := range refs {
		req := isc.NewOffLedgerRequest(
			isctest.RandomChainID(),
			isc.NewMessage(3, 14),
			uint64(i),
			200,
		).Sign(cryptolib.NewKeyPair())
		refs[i] = &isc.RequestRef{
			ID:   req.ID(),
			Hash: hashing.PseudoRandomHash(nil),
		}
	}
	return refs
}

// TestAggregateBatchProposals_ValidProposals verifies that a fully valid set of
// proposals does NOT get skipped.
func TestAggregateBatchProposals_ValidProposals(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 4
	f := 1
	nodeIDs := gpa.MakeTestNodeIDs(n)
	anchor := isctest.RandomStateAnchor()
	reqRefs := makeRequestRefs(3)
	gasCoin := makeSharedGasCoin()

	inputs := map[gpa.NodeID][]byte{}
	for i, nid := range nodeIDs {
		bp := makeValidProposal(uint16(i), n, &anchor, reqRefs, gasCoin)
		inputs[nid] = bp.Bytes()
	}

	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.False(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipMoreThanFNilProposals verifies that when >F
// proposals have a nil base anchor, aggregation is skipped.
func TestAggregateBatchProposals_SkipMoreThanFNilProposals(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 4
	f := 1
	nodeIDs := gpa.MakeTestNodeIDs(n)
	anchor := isctest.RandomStateAnchor()
	reqRefs := makeRequestRefs(3)
	gasCoin := makeSharedGasCoin()

	inputs := map[gpa.NodeID][]byte{}
	for i, nid := range nodeIDs {
		var bp *batchproposal.BatchProposal
		if i < f+1 {
			// These nodes propose nil anchor (void proposal).
			bp = batchproposal.NewBatchProposal(
				uint16(i),
				nil, // nil baseAnchor
				util.NewFixedSizeBitVector(uint16(n)).SetBits([]int{i}),
				nil,
				time.Now(),
				isctest.NewRandomAgentID(),
				nil,
				nil,
				nil,
			)
		} else {
			bp = makeValidProposal(uint16(i), n, &anchor, reqRefs, gasCoin)
		}
		inputs[nid] = bp.Bytes()
	}

	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipZeroValidProposals verifies that when all
// inputs are empty or garbage, aggregation is skipped.
func TestAggregateBatchProposals_SkipZeroValidProposals(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 4
	f := 1
	nodeIDs := gpa.MakeTestNodeIDs(n)

	inputs := map[gpa.NodeID][]byte{}
	for _, nid := range nodeIDs {
		inputs[nid] = []byte{0xFF, 0xFE, 0xFD} // garbage
	}

	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipEmptyInputs verifies that an empty input map
// results in a skip.
func TestAggregateBatchProposals_SkipEmptyInputs(t *testing.T) {
	log := testlogger.NewLogger(t)
	nodeIDs := gpa.MakeTestNodeIDs(4)

	aggr := batchproposal.AggregateBatchProposals(map[gpa.NodeID][]byte{}, nodeIDs, 1, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipAllEmptyPayloads verifies that when all
// inputs are empty byte slices (not garbage), aggregation is skipped.
func TestAggregateBatchProposals_SkipAllEmptyPayloads(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 4
	f := 1
	nodeIDs := gpa.MakeTestNodeIDs(n)

	inputs := map[gpa.NodeID][]byte{}
	for _, nid := range nodeIDs {
		inputs[nid] = []byte{}
	}

	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipNoRequestRefs verifies that valid proposals
// with no request refs cause a skip (decidedRequestRefs == 0).
func TestAggregateBatchProposals_SkipNoRequestRefs(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 1
	f := 0
	nodeIDs := gpa.MakeTestNodeIDs(n)
	anchor := isctest.RandomStateAnchor()

	bp := batchproposal.NewBatchProposal(
		0,
		&anchor,
		util.NewFixedSizeBitVector(uint16(n)).SetBits([]int{0}),
		nil,
		time.Now(),
		isctest.NewRandomAgentID(),
		[]*isc.RequestRef{}, // empty request refs
		[]*coin.CoinWithRef{makeSharedGasCoin()},
		parameterstest.L1Mock,
	)

	inputs := map[gpa.NodeID][]byte{nodeIDs[0]: bp.Bytes()}
	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipNoGasCoins verifies that proposals without
// gas coins cause a skip.
func TestAggregateBatchProposals_SkipNoGasCoins(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 1
	f := 0
	nodeIDs := gpa.MakeTestNodeIDs(n)
	anchor := isctest.RandomStateAnchor()
	reqRefs := makeRequestRefs(1)

	bp := batchproposal.NewBatchProposal(
		0,
		&anchor,
		util.NewFixedSizeBitVector(uint16(n)).SetBits([]int{0}),
		nil,
		time.Now(),
		isctest.NewRandomAgentID(),
		reqRefs,
		nil, // no gas coins
		parameterstest.L1Mock,
	)

	inputs := map[gpa.NodeID][]byte{nodeIDs[0]: bp.Bytes()}
	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipNoL1Params verifies that proposals without
// L1 params cause a skip.
func TestAggregateBatchProposals_SkipNoL1Params(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 1
	f := 0
	nodeIDs := gpa.MakeTestNodeIDs(n)
	anchor := isctest.RandomStateAnchor()
	reqRefs := makeRequestRefs(1)

	bp := batchproposal.NewBatchProposal(
		0,
		&anchor,
		util.NewFixedSizeBitVector(uint16(n)).SetBits([]int{0}),
		nil,
		time.Now(),
		isctest.NewRandomAgentID(),
		reqRefs,
		[]*coin.CoinWithRef{makeSharedGasCoin()},
		nil, // no L1 params
	)

	inputs := map[gpa.NodeID][]byte{nodeIDs[0]: bp.Bytes()}
	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipDisagreeingAnchors verifies that when nodes
// propose different base anchors and none reaches F+1 agreement, aggregation
// is skipped (decidedBaseAnchor == nil).
func TestAggregateBatchProposals_SkipDisagreeingAnchors(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 4
	f := 1
	nodeIDs := gpa.MakeTestNodeIDs(n)
	reqRefs := makeRequestRefs(1)
	gasCoin := makeSharedGasCoin()

	inputs := map[gpa.NodeID][]byte{}
	for i, nid := range nodeIDs {
		// Each node proposes a different anchor — none reaches F+1.
		anchor := isctest.RandomStateAnchor()
		bp := makeValidProposal(uint16(i), n, &anchor, reqRefs, gasCoin)
		inputs[nid] = bp.Bytes()
	}

	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}

// TestAggregateBatchProposals_SkipDisagreeingRequestRefs verifies that when
// base anchor is agreed upon but request refs are all different and none
// reaches F+1, aggregation is skipped.
func TestAggregateBatchProposals_SkipDisagreeingRequestRefs(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 4
	f := 1
	nodeIDs := gpa.MakeTestNodeIDs(n)
	anchor := isctest.RandomStateAnchor()
	gasCoin := makeSharedGasCoin()

	inputs := map[gpa.NodeID][]byte{}
	for i, nid := range nodeIDs {
		// Each node proposes unique request refs — none reaches F+1.
		uniqueRefs := makeRequestRefs(1)
		bp := makeValidProposal(uint16(i), n, &anchor, uniqueRefs, gasCoin)
		inputs[nid] = bp.Bytes()
	}

	aggr := batchproposal.AggregateBatchProposals(inputs, nodeIDs, f, log)
	require.True(t, aggr.ShouldBeSkipped())
}
