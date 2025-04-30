package statemanager

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	consGR "github.com/iotaledger/wasp/packages/chain/cons/cons_gr"
	smgpa "github.com/iotaledger/wasp/packages/chain/statemanager/gpa"
	"github.com/iotaledger/wasp/packages/chain/statemanager/gpa/inputs"
	gpautils "github.com/iotaledger/wasp/packages/chain/statemanager/gpa/utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/snapshots"
	"github.com/iotaledger/wasp/packages/chain/statemanager/utils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

type StateMgr interface {
	consGR.StateMgr
	// The StateMgr has to find a common ancestor for the prevAO and nextAO, then return
	// the state for Next ao and reject blocks in range (commonAO, prevAO]. The StateMgr
	// can determine relative positions of the corresponding blocks based on their state
	// indexes.
	ChainFetchStateDiff(
		ctx context.Context,
		prevAnchor, nextAnchor *isc.StateAnchor,
	) <-chan *inputs.ChainFetchStateDiffResults
	// Invoked by the chain when a set of server (access⁻¹) nodes has changed.
	// These nodes should be used to perform block replication.
	ChainNodesUpdated(serverNodes, accessNodes, committeeNodes []*cryptolib.PublicKey)
	// This is called to save a prelim block, received from other nodes.
	// That should happen on the access nodes to receive the active state faster.
	// This function should save the block (in the WAL) synchronously.
	PreliminaryBlock(block state.Block) error
}

type reqChainNodesUpdated struct {
	serverNodes    []*cryptolib.PublicKey
	accessNodes    []*cryptolib.PublicKey
	committeeNodes []*cryptolib.PublicKey
}

func (r *reqChainNodesUpdated) String() string {
	short := func(pkList []*cryptolib.PublicKey) string {
		return lo.Reduce(pkList, func(acc string, item *cryptolib.PublicKey, _ int) string {
			return acc + " " + gpa.NodeIDFromPublicKey(item).ShortString()
		}, "")
	}
	return fmt.Sprintf("{reqChainNodesUpdated, serverNodes=%s, accessNodes=%s, committeeNodes=%s",
		short(r.serverNodes),
		short(r.accessNodes),
		short(r.committeeNodes),
	)
}

type reqPreliminaryBlock struct {
	block state.Block
	reply chan error
}

func (r *reqPreliminaryBlock) String() string {
	return fmt.Sprintf("{reqPreliminaryBlock, block.L1Commitment=%v", r.block.L1Commitment())
}

func (r *reqPreliminaryBlock) Respond(err error) {
	r.reply <- err
}

type stateManager struct {
	log                  log.Logger
	chainID              isc.ChainID
	stateManagerGPA      gpa.GPA
	nodeRandomiser       utils.NodeRandomiser
	nodeIDToPubKey       map[gpa.NodeID]*cryptolib.PublicKey
	inputPipe            pipe.Pipe[gpa.Input]
	messagePipe          pipe.Pipe[*peering.PeerMessageIn]
	nodePubKeysPipe      pipe.Pipe[*reqChainNodesUpdated]
	preliminaryBlockPipe pipe.Pipe[*reqPreliminaryBlock]
	snapshotManager      snapshots.SnapshotManager
	wal                  gpautils.BlockWAL
	net                  peering.NetworkProvider
	netPeeringID         peering.PeeringID
	parameters           smgpa.StateManagerParameters
	ctx                  context.Context
	cleanupFun           func()
	shutdownCoordinator  *shutdown.Coordinator
}

var (
	_ StateMgr        = &stateManager{}
	_ consGR.StateMgr = &stateManager{}
)

const (
	constMsgTypeStm      byte = iota
	constStatusTimerTime      = 10 * time.Second
)

func New(
	ctx context.Context,
	chainID isc.ChainID,
	me *cryptolib.PublicKey,
	peerPubKeys []*cryptolib.PublicKey,
	net peering.NetworkProvider,
	wal gpautils.BlockWAL,
	snapshotManager snapshots.SnapshotManager,
	store state.Store,
	shutdownCoordinator *shutdown.Coordinator,
	metrics *metrics.ChainStateManagerMetrics,
	pipeMetrics *metrics.ChainPipeMetrics,
	log log.Logger,
	parameters smgpa.StateManagerParameters,
) (StateMgr, error) {
	smLog := log.NewChildLogger("SM")
	nr := utils.NewNodeRandomiserNoInit(gpa.NodeIDFromPublicKey(me), smLog)
	stateManagerGPA, err := smgpa.New(chainID, snapshotManager.GetLoadedSnapshotStateIndex(), nr, wal, store, metrics, smLog, parameters)
	if err != nil {
		smLog.LogErrorf("failed to create state manager GPA: %w", err)
		return nil, err
	}
	result := &stateManager{
		log:                  smLog,
		chainID:              chainID,
		stateManagerGPA:      stateManagerGPA,
		nodeRandomiser:       nr,
		inputPipe:            pipe.NewInfinitePipe[gpa.Input](),
		messagePipe:          pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		nodePubKeysPipe:      pipe.NewInfinitePipe[*reqChainNodesUpdated](),
		preliminaryBlockPipe: pipe.NewInfinitePipe[*reqPreliminaryBlock](),
		snapshotManager:      snapshotManager,
		wal:                  wal,
		net:                  net,
		netPeeringID:         peering.HashPeeringIDFromBytes(chainID.Bytes(), []byte("StateManager")), // ChainID × StateManager
		parameters:           parameters,
		ctx:                  ctx,
		shutdownCoordinator:  shutdownCoordinator,
	}

	pipeMetrics.TrackPipeLen("sm-inputPipe", result.inputPipe.Len)
	pipeMetrics.TrackPipeLen("sm-messagePipe", result.messagePipe.Len)
	pipeMetrics.TrackPipeLen("sm-nodePubKeysPipe", result.nodePubKeysPipe.Len)
	pipeMetrics.TrackPipeLen("sm-preliminaryBlockPipe", result.preliminaryBlockPipe.Len)

	result.handleNodePublicKeys(&reqChainNodesUpdated{
		serverNodes:    peerPubKeys,
		accessNodes:    []*cryptolib.PublicKey{},
		committeeNodes: []*cryptolib.PublicKey{},
	})

	unhook := result.net.Attach(&result.netPeeringID, peering.ReceiverStateManager, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != constMsgTypeStm {
			result.log.LogWarnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		result.messagePipe.In() <- recv
	})

	result.cleanupFun = func() {
		// The following lines cause this error:
		//	2024-03-20T11:53:22.932Z	DEBUG	TestNodeBasic/N=1,F=0,Reliable=true.8188.N#0.C-0x431e910b.LI-5	cons/cons.go:454	uponDSSIndexProposalReady
		//	panic: send on closed channel
		// TODO: Find a way to close the channels avoiding the error.
		// result.inputPipe.Close()
		// result.messagePipe.Close()
		// result.nodePubKeysPipe.Close()
		// result.preliminaryBlockPipe.Close()
		util.ExecuteIfNotNil(unhook)
	}

	go result.run()
	return result, nil
}

// -------------------------------------
// Implementations for chain package
// -------------------------------------

func (smT *stateManager) ChainFetchStateDiff(ctx context.Context, prevAnchor, nextAnchor *isc.StateAnchor) <-chan *inputs.ChainFetchStateDiffResults {
	input, resultCh := inputs.NewChainFetchStateDiff(ctx, prevAnchor, nextAnchor)
	smT.addInput(input)
	return resultCh
}

func (smT *stateManager) ChainNodesUpdated(serverNodes, accessNodes, committeeNodes []*cryptolib.PublicKey) {
	smT.nodePubKeysPipe.In() <- &reqChainNodesUpdated{
		serverNodes:    serverNodes,
		accessNodes:    accessNodes,
		committeeNodes: committeeNodes,
	}
}

func (smT *stateManager) PreliminaryBlock(block state.Block) error {
	reply := make(chan error, 1)
	smT.preliminaryBlockPipe.In() <- &reqPreliminaryBlock{
		block: block,
		reply: reply,
	}
	return <-reply
}

// -------------------------------------
// Implementations of consGR.StateMgr
// -------------------------------------

// ConsensusStateProposal asks State manager to ensure that all the blocks for aliasOutput are available.
// `nil` is sent via the returned channel upon successful retrieval of every block for aliasOutput.
func (smT *stateManager) ConsensusStateProposal(ctx context.Context, anchor *isc.StateAnchor) <-chan interface{} {
	input, resultCh := inputs.NewConsensusStateProposal(ctx, anchor)
	smT.addInput(input)
	return resultCh
}

// ConsensusDecidedState asks State manager to return a virtual state with stateCommitment as its state commitment
func (smT *stateManager) ConsensusDecidedState(ctx context.Context, anchor *isc.StateAnchor) <-chan state.State {
	input, resultCh := inputs.NewConsensusDecidedState(ctx, anchor)
	smT.addInput(input)
	return resultCh
}

func (smT *stateManager) ConsensusProducedBlock(ctx context.Context, stateDraft state.StateDraft) <-chan state.Block {
	input, resultCh := inputs.NewConsensusBlockProduced(ctx, stateDraft)
	smT.addInput(input)
	return resultCh
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smT *stateManager) addInput(input gpa.Input) {
	smT.inputPipe.In() <- input
}

func (smT *stateManager) run() {
	defer smT.cleanupFun()
	inputPipeCh := smT.inputPipe.Out()
	messagePipeCh := smT.messagePipe.Out()
	nodePubKeysPipeCh := smT.nodePubKeysPipe.Out()
	preliminaryBlockPipeCh := smT.preliminaryBlockPipe.Out()
	timerTickCh := smT.parameters.TimeProvider.After(smT.parameters.StateManagerTimerTickPeriod)
	statusTimerCh := smT.parameters.TimeProvider.After(constStatusTimerTime)
	for {
		if smT.ctx.Err() != nil {
			if smT.shutdownCoordinator == nil {
				return
			}
			// TODO what should the statemgr wait for?
			smT.shutdownCoordinator.WaitNestedWithLogging(1 * time.Second)
			smT.log.LogDebugf("Stopping state manager, because context was closed")
			smT.shutdownCoordinator.Done()
			return
		}
		select {
		case input, ok := <-inputPipeCh:
			if ok {
				smT.handleInput(input)
			} else {
				inputPipeCh = nil
			}
		case msg, ok := <-messagePipeCh:
			if ok {
				smT.handleMessage(msg)
			} else {
				messagePipeCh = nil
			}
		case msg, ok := <-nodePubKeysPipeCh:
			if ok {
				smT.handleNodePublicKeys(msg)
			} else {
				nodePubKeysPipeCh = nil
			}
		case msg, ok := <-preliminaryBlockPipeCh:
			if ok {
				smT.handlePreliminaryBlock(msg)
			} else {
				preliminaryBlockPipeCh = nil
			}
		case now, ok := <-timerTickCh:
			if ok {
				smT.handleTimerTick(now)
				timerTickCh = smT.parameters.TimeProvider.After(smT.parameters.StateManagerTimerTickPeriod)
			} else {
				timerTickCh = nil
			}
		case <-statusTimerCh:
			statusTimerCh = smT.parameters.TimeProvider.After(constStatusTimerTime)
			smT.log.LogDebugf("State manager loop iteration; there are %v inputs, %v messages, %v public key changes waiting to be handled",
				smT.inputPipe.Len(), smT.messagePipe.Len(), smT.nodePubKeysPipe.Len())
		case <-smT.ctx.Done():
			continue
		}
	}
}

func (smT *stateManager) handleInput(input gpa.Input) {
	outMsgs := smT.stateManagerGPA.Input(input)
	smT.sendMessages(outMsgs)
	smT.handleOutput()
}

func (smT *stateManager) handleMessage(peerMsg *peering.PeerMessageIn) {
	msg, err := smT.stateManagerGPA.UnmarshalMessage(peerMsg.MsgData)
	if err != nil {
		smT.log.LogWarnf("Parsing message failed: %v", err)
		return
	}
	msg.SetSender(gpa.NodeIDFromPublicKey(peerMsg.SenderPubKey))
	outMsgs := smT.stateManagerGPA.Message(msg)
	smT.sendMessages(outMsgs)
	smT.handleOutput()
}

func (smT *stateManager) handleOutput() {
	output := smT.stateManagerGPA.Output().(smgpa.StateManagerOutput)
	for _, snapshotInfo := range output.TakeBlocksCommitted() {
		smT.snapshotManager.BlockCommittedAsync(snapshotInfo)
	}
	for _, input := range output.TakeNextInputs() {
		smT.addInput(input)
	}
}

func (smT *stateManager) handleNodePublicKeys(req *reqChainNodesUpdated) {
	smT.log.LogDebugf("handleNodePublicKeys: %v", req)
	smT.nodeIDToPubKey = map[gpa.NodeID]*cryptolib.PublicKey{}
	peerNodeIDs := []gpa.NodeID{}
	for _, pubKey := range req.serverNodes {
		nodeID := gpa.NodeIDFromPublicKey(pubKey)
		if _, ok := smT.nodeIDToPubKey[nodeID]; !ok {
			smT.nodeIDToPubKey[nodeID] = pubKey
			peerNodeIDs = append(peerNodeIDs, nodeID)
		}
	}
	for _, pubKey := range req.accessNodes {
		nodeID := gpa.NodeIDFromPublicKey(pubKey)
		if _, ok := smT.nodeIDToPubKey[nodeID]; !ok {
			smT.nodeIDToPubKey[nodeID] = pubKey
			// Don't use access nodes for queries.
		}
	}
	for _, pubKey := range req.committeeNodes {
		nodeID := gpa.NodeIDFromPublicKey(pubKey)
		if _, ok := smT.nodeIDToPubKey[nodeID]; !ok {
			smT.nodeIDToPubKey[nodeID] = pubKey
			peerNodeIDs = append(peerNodeIDs, nodeID)
		}
	}

	smT.log.LogInfof("Updating list of nodeIDs: [%v]",
		lo.Reduce(peerNodeIDs, func(acc string, item gpa.NodeID, _ int) string {
			return acc + " " + item.ShortString()
		}, ""),
	)
	smT.nodeRandomiser.UpdateNodeIDs(peerNodeIDs)
}

func (smT *stateManager) handlePreliminaryBlock(msg *reqPreliminaryBlock) {
	if !smT.wal.Contains(msg.block.Hash()) {
		if err := smT.wal.Write(msg.block); err != nil {
			smT.log.LogWarnf("Preliminary block index %v %s cannot be saved to the WAL: %v",
				msg.block.StateIndex(), msg.block.L1Commitment(), err)
			msg.Respond(err)
			return
		}
		smT.log.LogWarnf("Preliminary block index %v %s saved to the WAL.", msg.block.StateIndex(), msg.block.L1Commitment())
		msg.Respond(nil)
		return
	}
	smT.log.LogWarnf("Preliminary block index %v %s already exist in the WAL.", msg.block.StateIndex(), msg.block.L1Commitment())
	msg.Respond(nil)
}

func (smT *stateManager) handleTimerTick(now time.Time) {
	smT.handleInput(inputs.NewStateManagerTimerTick(now))
}

func (smT *stateManager) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(msg gpa.Message) {
		msgBytes := lo.Must(gpa.MarshalMessage(msg))
		pm := peering.NewPeerMessageData(smT.netPeeringID, peering.ReceiverStateManager, constMsgTypeStm, msgBytes)
		recipientPubKey, ok := smT.nodeIDToPubKey[msg.Recipient()]
		if !ok {
			smT.log.LogDebugf("Dropping outgoing message, because NodeID=%s it is not in the NodeList.", msg.Recipient().ShortString())
			return
		}
		smT.net.SendMsgByPubKey(recipientPubKey, pm)
	})
}
