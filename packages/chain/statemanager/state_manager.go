package statemanager

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/core/logger"
	consGR "github.com/iotaledger/wasp/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smUtils"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
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
		prevAO, nextAO *isc.AliasOutputWithID,
	) <-chan *smInputs.ChainFetchStateDiffResults
	// Invoked by the chain when a set of server (access⁻¹) nodes has changed.
	// These nodes should be used to perform block replication.
	ChainServerNodesUpdated(serverNodePubKeys []*cryptolib.PublicKey)
}

type stateManager struct {
	log             *logger.Logger
	chainID         isc.ChainID
	stateManagerGPA gpa.GPA
	nodeRandomiser  smUtils.NodeRandomiser
	nodeIDToPubKey  map[gpa.NodeID]*cryptolib.PublicKey
	inputPipe       pipe.Pipe
	messagePipe     pipe.Pipe
	nodePubKeysPipe pipe.Pipe
	net             peering.NetworkProvider
	netPeeringID    peering.PeeringID
	timers          smGPA.StateManagerTimers
	ctx             context.Context
	cleanupFun      func()
}

var (
	_ StateMgr        = &stateManager{}
	_ consGR.StateMgr = &stateManager{}
)

const (
	constMsgTypeStm           byte = iota
	constDefaultTimerTickTime      = 1 * time.Second
)

func New(
	ctx context.Context,
	chainID isc.ChainID,
	me *cryptolib.PublicKey,
	peerPubKeys []*cryptolib.PublicKey,
	net peering.NetworkProvider,
	wal smGPAUtils.BlockWAL,
	store state.Store,
	log *logger.Logger,
	timersOpt ...smGPA.StateManagerTimers,
) (StateMgr, error) {
	nr := smUtils.NewNodeRandomiserNoInit(pubKeyAsNodeID(me), log)
	var timers smGPA.StateManagerTimers
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = smGPA.NewStateManagerTimers()
	}

	stateManagerGPA, err := smGPA.New(chainID, nr, wal, store, log, timers)
	if err != nil {
		log.Errorf("Failed to create state manager GPA: %v", err)
		return nil, err
	}
	result := &stateManager{
		log:             log,
		chainID:         chainID,
		stateManagerGPA: stateManagerGPA,
		nodeRandomiser:  nr,
		inputPipe:       pipe.NewDefaultInfinitePipe(),
		messagePipe:     pipe.NewDefaultInfinitePipe(),
		nodePubKeysPipe: pipe.NewDefaultInfinitePipe(),
		net:             net,
		netPeeringID:    peering.PeeringIDFromBytes(hashing.HashDataBlake2b(chainID.Bytes(), []byte("STM")).Bytes()),
		timers:          timers,
		ctx:             ctx,
	}
	result.handleNodePublicKeys(peerPubKeys)

	attachID := result.net.Attach(&result.netPeeringID, peering.PeerMessageReceiverStateManager, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != constMsgTypeStm {
			result.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		result.messagePipe.In() <- recv
	})

	result.cleanupFun = func() {
		// result.inputPipe.Close() // TODO: Uncomment it.
		// result.messagePipe.Close() // TODO: Uncomment it.
		result.net.Detach(attachID)
	}

	go result.run()
	return result, nil
}

// -------------------------------------
// Implementations for chain package
// -------------------------------------

func (smT *stateManager) ChainFetchStateDiff(ctx context.Context, prevAO, nextAO *isc.AliasOutputWithID) <-chan *smInputs.ChainFetchStateDiffResults {
	input, resultCh := smInputs.NewChainFetchStateDiff(ctx, prevAO, nextAO)
	smT.addInput(input)
	return resultCh
}

func (smT *stateManager) ChainServerNodesUpdated(serverNodePubKeys []*cryptolib.PublicKey) {
	smT.nodePubKeysPipe.In() <- serverNodePubKeys
}

// -------------------------------------
// Implementations of consGR.StateMgr
// -------------------------------------

// ConsensusStateProposal asks State manager to ensure that all the blocks for aliasOutput are available.
// `nil` is sent via the returned channel upon successful retrieval of every block for aliasOutput.
func (smT *stateManager) ConsensusStateProposal(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan interface{} {
	input, resultCh := smInputs.NewConsensusStateProposal(ctx, aliasOutput)
	smT.addInput(input)
	return resultCh
}

// ConsensusDecidedState asks State manager to return a virtual state with stateCommitment as its state commitment
func (smT *stateManager) ConsensusDecidedState(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan state.State {
	input, resultCh := smInputs.NewConsensusDecidedState(ctx, aliasOutput)
	smT.addInput(input)
	return resultCh
}

func (smT *stateManager) ConsensusProducedBlock(ctx context.Context, stateDraft state.StateDraft) <-chan error {
	input, resultCh := smInputs.NewConsensusBlockProduced(ctx, stateDraft)
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
	ctxCloseCh := smT.ctx.Done()
	inputPipeCh := smT.inputPipe.Out()
	messagePipeCh := smT.messagePipe.Out()
	nodePubKeysPipeCh := smT.nodePubKeysPipe.Out()
	timerTickCh := smT.timers.TimeProvider.After(smT.timers.StateManagerTimerTickPeriod)
	for {
		smT.log.Debugf("State manager loop iteration; there are %v inputs, %v messages, %v public key changes waiting to be handled",
			smT.inputPipe.Len(), smT.messagePipe.Len(), smT.nodePubKeysPipe.Len())
		select {
		case input, ok := <-inputPipeCh:
			if ok {
				smT.handleInput(input.(gpa.Input))
			} else {
				inputPipeCh = nil
			}
		case msg, ok := <-messagePipeCh:
			if ok {
				smT.handleMessage(msg.(*peering.PeerMessageIn))
			} else {
				messagePipeCh = nil
			}
		case msg, ok := <-nodePubKeysPipeCh:
			if ok {
				smT.handleNodePublicKeys(msg.([]*cryptolib.PublicKey))
			} else {
				nodePubKeysPipeCh = nil
			}
		case now, ok := <-timerTickCh:
			if ok {
				smT.handleTimerTick(now)
				timerTickCh = smT.timers.TimeProvider.After(smT.timers.StateManagerTimerTickPeriod)
			} else {
				timerTickCh = nil
			}
		case <-ctxCloseCh:
			smT.log.Debugf("Stopping state manager, because context was closed")
			return
		}
	}
}

func (smT *stateManager) handleInput(input gpa.Input) {
	outMsgs := smT.stateManagerGPA.Input(input)
	smT.sendMessages(outMsgs)
}

func (smT *stateManager) handleMessage(peerMsg *peering.PeerMessageIn) {
	msg, err := smT.stateManagerGPA.UnmarshalMessage(peerMsg.MsgData)
	if err != nil {
		smT.log.Warnf("Parsing message failed: %v", err)
		return
	}
	msg.SetSender(pubKeyAsNodeID(peerMsg.SenderPubKey))
	outMsgs := smT.stateManagerGPA.Message(msg)
	smT.sendMessages(outMsgs)
}

func (smT *stateManager) handleNodePublicKeys(peerPubKeys []*cryptolib.PublicKey) {
	smT.nodeIDToPubKey = make(map[gpa.NodeID]*cryptolib.PublicKey)
	peerNodeIDs := make([]gpa.NodeID, len(peerPubKeys))
	for i := range peerPubKeys {
		peerNodeIDs[i] = pubKeyAsNodeID(peerPubKeys[i])
		smT.nodeIDToPubKey[peerNodeIDs[i]] = peerPubKeys[i]
	}
	smT.log.Infof("Updating list of nodeIDs: [%v]",
		lo.Reduce(peerNodeIDs, func(acc string, item gpa.NodeID, _ int) string {
			return acc + " " + string(item)
		}, ""),
	)
	smT.nodeRandomiser.UpdateNodeIDs(peerNodeIDs)
}

func (smT *stateManager) handleTimerTick(now time.Time) {
	smT.handleInput(smInputs.NewStateManagerTimerTick(now))
}

func (smT *stateManager) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(msg gpa.Message) {
		msgData, err := msg.MarshalBinary()
		if err != nil {
			smT.log.Warnf("Failed to marshal message for sending: %v", err)
			return
		}
		pm := &peering.PeerMessageData{
			PeeringID:   smT.netPeeringID,
			MsgReceiver: peering.PeerMessageReceiverStateManager,
			MsgType:     constMsgTypeStm,
			MsgData:     msgData,
		}
		recipientPubKey, ok := smT.nodeIDToPubKey[msg.Recipient()]
		if !ok {
			smT.log.Debugf("Dropping outgoing message, because NodeID=%v it is not in the NodeList.", msg.Recipient())
			return
		}
		smT.net.SendMsgByPubKey(recipientPubKey, pm)
	})
}

func pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID(pubKey.String())
}
