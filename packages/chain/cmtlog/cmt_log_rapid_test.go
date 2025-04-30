package cmtlog_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// 	"pgregory.net/rapid"

// 	"github.com/iotaledger/wasp/clients/iota-go/iotago"
// 	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
// 	"github.com/iotaledger/wasp/packages/chain/cmtlog"
// 	"github.com/iotaledger/wasp/packages/cryptolib"
// 	"github.com/iotaledger/wasp/packages/gpa"
// 	"github.com/iotaledger/wasp/packages/isc"
// 	"github.com/iotaledger/wasp/packages/isc/isctest"
// 	"github.com/iotaledger/wasp/packages/testutil"
// 	"github.com/iotaledger/wasp/packages/testutil/testlogger"
// 	"github.com/iotaledger/wasp/packages/testutil/testpeers"
// )

// type cmtLogTestRapidSM struct {
// 	anchorRef       iotago.ObjectRef
// 	chainID         isc.ChainID
// 	governorAddress *cryptolib.Address
// 	stateAddress    *cryptolib.Address
// 	tc              *gpa.TestContext
// 	l1Chain         []*isc.StateAnchor // The actual chain.
// 	l1Delivered     map[gpa.NodeID]int // Position of the last element from l1Chain to delivered for the corresponding node (-1 means none).
// 	genAOSerial     uint32
// 	genNodeID       *rapid.Generator[gpa.NodeID]
// }

// var _ rapid.StateMachine = &cmtLogTestRapidSM{}

// func newCmtLogTestRapidSM(t *rapid.T) *cmtLogTestRapidSM {
// 	sm := new(cmtLogTestRapidSM)
// 	n := 4
// 	f := 1
// 	log := testlogger.NewLogger(t)
// 	//
// 	// Chain identifiers.
// 	sm.anchorRef = *iotatest.RandomObjectRef()
// 	sm.chainID = isc.ChainIDFromObjectID(*sm.anchorRef.ObjectID)
// 	sm.governorAddress = cryptolib.NewKeyPair().Address()
// 	//
// 	// Node identities.
// 	_, peerIdentities := testpeers.SetupKeys(uint16(n))
// 	peerPubKeys := testpeers.PublicKeys(peerIdentities)
// 	//
// 	// Committee.
// 	committeeAddress, committeeKeyShares := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
// 	sm.stateAddress = committeeAddress
// 	//
// 	// Construct the algorithm nodes.
// 	gpaNodeIDs := gpa.NodeIDsFromPublicKeys(peerPubKeys)
// 	gpaNodes := map[gpa.NodeID]gpa.GPA{}
// 	for i := range gpaNodeIDs {
// 		dkShare, err := committeeKeyShares[i].LoadDKShare(committeeAddress)
// 		require.NoError(t, err)
// 		consensusStateRegistry := testutil.NewConsensusStateRegistry() // Empty store in this case.
// 		cmtLogInst, err := cmtlog.New(gpaNodeIDs[i], sm.chainID, dkShare, consensusStateRegistry, gpa.NodeIDFromPublicKey, true, -1, 1, nil, log.NewChildLogger(fmt.Sprintf("N%v", i)))
// 		require.NoError(t, err)
// 		gpaNodes[gpaNodeIDs[i]] = cmtLogInst.AsGPA()
// 	}
// 	sm.tc = gpa.NewTestContext(gpaNodes)
// 	sm.l1Chain = []*isc.StateAnchor{}
// 	sm.l1Delivered = map[gpa.NodeID]int{}
// 	//
// 	// Generators.
// 	sm.genAOSerial = 0
// 	sm.genNodeID = rapid.SampledFrom(gpaNodeIDs)
// 	//
// 	// Start it.
// 	sm.l1Chain = append(sm.l1Chain, sm.nextStateAnchorWithStateIndex(0))
// 	for _, nid := range gpaNodeIDs {
// 		sm.l1Delivered[nid] = -1
// 	}
// 	return sm
// }

// // simulate StateAnchor to state transition
// func (sm *cmtLogTestRapidSM) nextStateAnchorWithStateIndex(stateIndex uint32) *isc.StateAnchor {
// 	newAnchor := iotago.ObjectRef{
// 		ObjectID: sm.anchorRef.ObjectID,
// 		Version:  sm.anchorRef.Version + 1,
// 		Digest:   iotatest.RandomDigest(), // This should change update time ObjectRef has changed
// 	}
// 	stateAnchor := isctest.RandomStateAnchor(isctest.RandomAnchorOption{
// 		StateIndex: &stateIndex,
// 		ObjectRef:  &newAnchor,
// 		Owner:      sm.stateAddress.AsIotaAddress(),
// 	})
// 	return &stateAnchor
// }

// // func (sm *cmtLogTestRapidSM) ConsDone(t *rapid.T) {
// // 	nodeID := sm.genNodeID.Draw(t, "node")
// // 	var li cmtLog.LogIndex         // TODO: Set it.
// // 	var pAO iotago.ObjectID        // TODO: Set it.
// // 	var bAO iotago.ObjectID        // TODO: Set it.
// // 	var nAO *isc.StateAnchor // TODO: Set it.
// // 	sm.tc.WithInput(nodeID, cmtLog.NewInputConsensusOutputDone(li, pAO, bAO, nAO))
// // 	sm.tc.RunAll()
// // }

// // func (sm *cmtLogTestRapidSM) ConsSkip(t *rapid.T) {
// // 	nodeID := sm.genNodeID.Draw(t, "node")
// // 	var li cmtLog.LogIndex  // TODO: Set it.
// // 	var pAO iotago.ObjectID // TODO: Set it.
// // 	sm.tc.WithInput(nodeID, cmtLog.NewInputConsensusOutputSkip(li, pAO))
// // 	sm.tc.RunAll()
// // }

// // func (sm *cmtLogTestRapidSM) ConsRecover(t *rapid.T) {
// // 	nodeID := sm.genNodeID.Draw(t, "node")
// // 	var li cmtLog.LogIndex // TODO: Set it.
// // 	sm.tc.WithInput(nodeID, cmtLog.NewInputConsensusTimeout(li))
// // 	sm.tc.RunAll()
// // }

// // func (sm *cmtLogTestRapidSM) ConsConfirmed(t *rapid.T) {
// // 	nodeID := sm.genNodeID.Draw(t, "node")
// // 	var ao *isc.StateAnchor // TODO: Set it.
// // 	var li cmtLog.LogIndex        // TODO: Set it.
// // 	sm.tc.WithInput(nodeID, cmtLog.NewInputConsensusOutputConfirmed(ao, li))
// // 	sm.tc.RunAll()
// // }

// // func (sm *cmtLogTestRapidSM) ConsRejected(t *rapid.T) {
// // 	nodeID := sm.genNodeID.Draw(t, "node")
// // 	var ao *isc.StateAnchor // TODO: Set it.
// // 	var li cmtLog.LogIndex        // TODO: Set it.
// // 	sm.tc.WithInput(nodeID, cmtLog.NewInputConsensusOutputRejected(ao, li))
// // 	sm.tc.RunAll()
// // }

// func (sm *cmtLogTestRapidSM) AliasOutputConfirmed(t *rapid.T) {
// 	nodeID := sm.genNodeID.Draw(t, "node")
// 	if len(sm.l1Chain)-sm.l1Delivered[nodeID] <= 1 {
// 		t.SkipNow()
// 	}
// 	deliverIdx := rapid.IntRange(sm.l1Delivered[nodeID]+1, len(sm.l1Chain)-1).Draw(t, "deliverIdx")
// 	ao := sm.l1Chain[deliverIdx]
// 	sm.l1Delivered[nodeID] = deliverIdx
// 	sm.tc.WithInput(nodeID, cmtlog.NewInputAnchorConfirmed(ao))
// 	sm.tc.RunAll()
// }

// // Trim the chain to some length and reset the delivery counters to all the peers to not exceed the trimmed chain.
// func (sm *cmtLogTestRapidSM) L1Reorg(t *rapid.T) {
// 	chainLen := len(sm.l1Chain)
// 	if chainLen <= 1 {
// 		t.SkipNow()
// 	}
// 	newLast := rapid.IntRange(1, chainLen-1).Draw(t, "reorgTo")
// 	sm.l1Chain = sm.l1Chain[:newLast]
// 	for nid, pos := range sm.l1Delivered {
// 		if pos >= newLast {
// 			sm.l1Delivered[nid] = newLast - 1
// 		}
// 	}
// }

// func (sm *cmtLogTestRapidSM) Check(t *rapid.T) {
// 	sm.invHaveConsRunningOrTxConfirming(t)
// }

// // We want here to check liveness -- if the chain never stops. This test framework
// // don't allow to check temporal properties, only state predicates and invariants.
// // So we reformulate the property to the condition, that always, either TX is confirming
// // or consensus is running. Assuming fairness for both, it should imply liveness of
// // this algorithm.
// func (sm *cmtLogTestRapidSM) invHaveConsRunningOrTxConfirming(t *rapid.T) {
// 	// TODO: >...
// }

// var _ rapid.StateMachine = &cmtLogTestRapidSM{}

// func TestCmtLogRapid(t *testing.T) {
// 	rapid.Check(t, func(t *rapid.T) {
// 		sm := newCmtLogTestRapidSM(t)
// 		t.Repeat(rapid.StateMachineActions(sm))
// 	})
// }
