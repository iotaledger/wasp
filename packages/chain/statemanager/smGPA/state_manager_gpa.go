//
//
//
//
//
//

package smGPA

import (
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smMessages"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smUtils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type blockRequestsWithCommitment struct {
	l1Commitment  *state.L1Commitment
	blockRequests []blockRequest
}

type stateManagerGPA struct {
	log                     *logger.Logger
	chainID                 *isc.ChainID
	blockCache              smGPAUtils.BlockCache
	blockRequests           map[state.BlockHash]*blockRequestsWithCommitment
	nodeRandomiser          smUtils.NodeRandomiser
	store                   state.Store
	timers                  StateManagerTimers
	lastGetBlocksTime       time.Time
	lastCleanBlockCacheTime time.Time
	lastCleanRequestsTime   time.Time
	currentStateIndex       uint32              // Not used in the algorithm; only stores status of the state manager
	currentL1Commitment     *state.L1Commitment // Not used in the algorithm; only stores status of the state manager
	stateOutputSeq          blockRequestID
	lastStateOutputSeq      blockRequestID // Must be strictly monotonic; see handleChainReceiveConfirmedAliasOutput
	lastBlockRequestID      blockRequestID
}

var _ gpa.GPA = &stateManagerGPA{}

const (
	numberOfNodesToRequestBlockFromConst = 5
)

func New(chainID *isc.ChainID, nr smUtils.NodeRandomiser, wal smGPAUtils.BlockWAL, store state.Store, log *logger.Logger, timersOpt ...StateManagerTimers) (gpa.GPA, error) {
	var err error
	var timers StateManagerTimers
	smLog := log.Named("gpa")
	if len(timersOpt) > 0 {
		timers = timersOpt[0]
	} else {
		timers = NewStateManagerTimers()
	}
	blockCache, err := smGPAUtils.NewBlockCache(timers.TimeProvider, wal, smLog)
	if err != nil {
		smLog.Errorf("Error creating block cache: %v", err)
		return nil, err
	}
	result := &stateManagerGPA{
		log:                     smLog,
		chainID:                 chainID,
		blockCache:              blockCache,
		blockRequests:           make(map[state.BlockHash]*blockRequestsWithCommitment),
		nodeRandomiser:          nr,
		store:                   store,
		timers:                  timers,
		lastGetBlocksTime:       time.Time{},
		lastCleanBlockCacheTime: time.Time{},
		currentStateIndex:       0,
		currentL1Commitment:     state.OriginL1Commitment(),
		stateOutputSeq:          0,
		lastStateOutputSeq:      0,
		lastBlockRequestID:      0,
	}

	return result, nil
}

// -------------------------------------
// Implementation for gpa.GPA interface
// -------------------------------------

func (smT *stateManagerGPA) Input(input gpa.Input) gpa.OutMessages {
	switch inputCasted := input.(type) {
	case *smInputs.ChainReceiveConfirmedAliasOutput: // From chain
		return smT.handleChainReceiveConfirmedAliasOutput(inputCasted.GetStateOutput())
	case *smInputs.ConsensusStateProposal: // From consensus
		return smT.handleConsensusStateProposal(inputCasted)
	case *smInputs.ConsensusDecidedState: // From consensus
		return smT.handleConsensusDecidedState(inputCasted)
	case *smInputs.ConsensusBlockProduced: // From consensus
		return smT.handleConsensusBlockProduced(inputCasted)
	case *smInputs.MempoolStateRequest: // From mempool
		return smT.handleMempoolStateRequest(inputCasted)
	case *smInputs.StateManagerTimerTick: // From state manager go routine
		return smT.handleStateManagerTimerTick(inputCasted.GetTime())
	default:
		smT.log.Warnf("Unknown input received, ignoring it: type=%T, message=%v", input, input)
		return nil // No messages to send
	}
}

func (smT *stateManagerGPA) Message(msg gpa.Message) gpa.OutMessages {
	switch msgCasted := msg.(type) {
	case *smMessages.GetBlockMessage:
		return smT.handlePeerGetBlock(msgCasted.Sender(), msgCasted.GetL1Commitment())
	case *smMessages.BlockMessage:
		return smT.handlePeerBlock(msgCasted.Sender(), msgCasted.GetBlock())
	default:
		smT.log.Warnf("Unknown message received, ignoring it: type=%T, message=%v", msg, msg)
		return nil // No messages to send
	}
}

func (smT *stateManagerGPA) Output() gpa.Output {
	return nil
}

func (smT *stateManagerGPA) StatusString() string {
	return "" // TODO
}

func (smT *stateManagerGPA) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("Error unmarshalling message: slice of length %v is too short", data)
	}
	var message gpa.Message
	switch data[0] {
	case smMessages.MsgTypeBlockMessage:
		message = smMessages.NewEmptyBlockMessage()
	case smMessages.MsgTypeGetBlockMessage:
		message = smMessages.NewEmptyGetBlockMessage()
	default:
		return nil, fmt.Errorf("Error unmarshalling message: message type %v unknown", data[0])
	}
	err := message.UnmarshalBinary(data)
	return message, err
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smT *stateManagerGPA) handlePeerGetBlock(from gpa.NodeID, commitment *state.L1Commitment) gpa.OutMessages {
	smT.log.Debugf("GetBlock %s message received from peer %s", commitment, from)
	block := smT.getBlock(commitment)
	if block == nil {
		smT.log.Debugf("GetBlock %s: block not found, peer %s request ignored", commitment, from)
		return nil // No messages to send
	}
	smT.log.Debugf("GetBlock %s: block found, sending it to peer %s", commitment, from)
	return gpa.NoMessages().Add(smMessages.NewBlockMessage(block, from))
}

func (smT *stateManagerGPA) handlePeerBlock(from gpa.NodeID, block state.Block) gpa.OutMessages {
	blockCommitment := block.L1Commitment()
	blockHash := blockCommitment.GetBlockHash()
	smT.log.Debugf("Block %s message received from peer %s", blockCommitment, from)
	requestsWC, ok := smT.blockRequests[blockHash]
	if !ok {
		smT.log.Debugf("Block %s: block is not needed, ignoring it", blockCommitment)
		return nil // No messages to send
	}
	requests := requestsWC.blockRequests
	smT.log.Debugf("Block %s: %v requests are waiting for the block", blockCommitment, len(requests))
	err := smT.blockCache.AddBlock(block)
	if err != nil {
		return nil // No messages to send
	}
	request := smT.createStateBlockRequestLocal(blockCommitment, func(_ obtainStateFun) {})
	newRequests := append(requests, request)
	delete(smT.blockRequests, blockHash)
	for _, request := range newRequests {
		request.blockAvailable(block)
	}
	previousCommitment := block.PreviousL1Commitment()
	smT.log.Debugf("Block %s: tracing previous block %s", blockCommitment, previousCommitment)
	messages, err := smT.traceBlockChain(previousCommitment, newRequests)
	if err != nil {
		return nil // No messages to send
	}
	smT.log.Debugf("Block %s message from peer %s handled", blockCommitment, from)
	return messages
}

func (smT *stateManagerGPA) handleChainReceiveConfirmedAliasOutput(aliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	aliasOutputID := isc.OID(aliasOutput.ID())
	smT.log.Debugf("Chain receive confirmed alias output %s input received...", aliasOutputID)
	stateCommitment, err := state.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		smT.log.Errorf("Chain receive confirmed alias output %s: error retrieving state commitment from alias output: %v",
			aliasOutputID, err)
		return nil // No messages to send
	}
	// `lastStateOutputSeq` is used tu ensure that older state outputs don't
	// overwrite the newer ones. It is assumed that this method receives alias outputs
	// in strictly the same order as they are approved by L1.
	smT.lastStateOutputSeq++
	seq := smT.lastStateOutputSeq
	smT.log.Debugf("Chain receive confirmed alias output %s: alias output is %v-th in state manager", aliasOutputID, seq)
	request := smT.createStateBlockRequestLocal(stateCommitment, func(_ obtainStateFun) {
		smT.log.Debugf("Chain receive confirmed alias output %s (%v-th in state manager): all blocks for the state are in store", aliasOutputID, seq)
		if seq <= smT.stateOutputSeq {
			smT.log.Warnf("Chain receive confirmed alias output %s (%v-th in state manager): alias output is outdated", aliasOutputID, seq)
		} else {
			smT.currentStateIndex = aliasOutput.GetStateIndex()
			smT.currentL1Commitment = stateCommitment
			smT.log.Debugf("Chain receive confirmed alias output %s (%v-th in state manager): STATE CHANGE: state manager is at state index %v, commitment %s",
				aliasOutputID, seq, smT.currentStateIndex, smT.currentL1Commitment)
			smT.stateOutputSeq = seq
		}
	})
	messages, err := smT.traceBlockChainByRequest(request)
	if err != nil {
		smT.log.Errorf("Chain receive confirmed alias output %s (%v-th in state manager): error tracing block chain: %v",
			aliasOutputID, seq, err)
		return nil // No messages to send
	}
	smT.log.Debugf("Chain receive confirmed alias output %s (%v-th in state manager) input handled", aliasOutputID, seq)
	return messages
}

func (smT *stateManagerGPA) handleConsensusStateProposal(csp *smInputs.ConsensusStateProposal) gpa.OutMessages {
	smT.log.Debugf("Consensus state proposal %s input received...", csp.GetL1Commitment())
	smT.lastBlockRequestID++
	request := newStateBlockRequestFromConsensusStateProposal(csp, smT.log, smT.lastBlockRequestID)
	messages, err := smT.traceBlockChainByRequest(request)
	if err != nil {
		smT.log.Errorf("Consensus state proposal %s: error tracing block chain: %v", csp.GetL1Commitment(), err)
		return nil // No messages to send
	}
	smT.log.Debugf("Consensus state proposal %s input handled", csp.GetL1Commitment())
	return messages
}

func (smT *stateManagerGPA) handleConsensusDecidedState(cds *smInputs.ConsensusDecidedState) gpa.OutMessages {
	smT.log.Debugf("Consensus decided state %s input received...", cds.GetL1Commitment())
	smT.lastBlockRequestID++
	messages, err := smT.traceBlockChainByRequest(newStateBlockRequestFromConsensusDecidedState(cds, smT.log, smT.lastBlockRequestID))
	if err != nil {
		smT.log.Errorf("Consensus decided state %s: error tracing block chain: %v", cds.GetL1Commitment(), err)
		return nil // No messages to send
	}
	smT.log.Debugf("Consensus decided state %s input handled", cds.GetL1Commitment())
	return messages
}

func (smT *stateManagerGPA) handleConsensusBlockProduced(input *smInputs.ConsensusBlockProduced) gpa.OutMessages {
	commitment := input.GetStateDraft().BaseL1Commitment()
	smT.log.Debugf("Consensus block produced on state %s input received...", commitment)
	request := smT.createStateBlockRequestLocal(commitment, func(_ obtainStateFun) {
		smT.log.Debugf("Consensus block produced on state %s: all blocks to commit state draft are in store, committing it", commitment)
		block := smT.store.Commit(input.GetStateDraft())
		smT.log.Debugf("Consensus block produced on state %s: state draft has been committed to the store; block %s received",
			commitment, block.L1Commitment())
		smT.blockCache.AddBlock(block)
		requestsWC, ok := smT.blockRequests[block.Hash()]
		if ok {
			// It seems, there should never be requests, waiting for this block.
			// If there were some, they should have been marked completed at the
			// same time, as this request. Leaving this code just in case.
			smT.log.Debugf("Consensus block produced on state %s: marking %v requests, waiting for block %s, completed",
				commitment, len(requestsWC.blockRequests), block.L1Commitment())
			for _, request := range requestsWC.blockRequests {
				smT.markRequestCompleted(request)
			}
		}
	})
	messages, err := smT.traceBlockChainByRequest(request)
	input.Respond(err)
	smT.log.Debugf("Consensus block produced on state %s input handled; error=%v", commitment, err)
	return messages
}

func (smT *stateManagerGPA) handleMempoolStateRequest(input *smInputs.MempoolStateRequest) gpa.OutMessages {
	oldNewContainer := &struct {
		oldStateBlockRequest          *stateBlockRequest
		newStateBlockRequest          *stateBlockRequest
		oldStateBlockRequestCompleted bool
		newStateBlockRequestCompleted bool
		obtainNewStateFun             obtainStateFun
	}{
		oldStateBlockRequestCompleted: false,
		newStateBlockRequestCompleted: false,
	}
	isValidFun := func() bool { return input.IsValid() }
	obtainCommittedBlockFun := func(commitment *state.L1Commitment) state.Block {
		result, err := smT.store.BlockByTrieRoot(commitment.GetTrieRoot())
		if err != nil {
			smT.log.Panicf("Cannot obtain block %s: %v", commitment, err)
		}
		return result
	}
	respondFun := func() {
		oldBaseIndex := input.GetOldStateIndex()
		newBaseIndex := input.GetNewStateIndex()
		var commonIndex uint32
		if newBaseIndex > oldBaseIndex {
			commonIndex = oldBaseIndex
		} else {
			commonIndex = newBaseIndex
		}

		oldBC := oldNewContainer.oldStateBlockRequest.getBlockChain()
		newBC := oldNewContainer.newStateBlockRequest.getBlockChain()
		oldCommitment := input.GetOldL1Commitment()
		newCommitment := input.GetNewL1Commitment()
		oldCOB := newChainOfBlocks(oldBC, oldCommitment, oldBaseIndex, obtainCommittedBlockFun)
		newCOB := newChainOfBlocks(newBC, newCommitment, newBaseIndex, obtainCommittedBlockFun)

		respondToMempoolFun := func(index uint32) {
			newState, err := oldNewContainer.obtainNewStateFun()
			if err != nil {
				smT.log.Errorf("Unable to obtain new state: %v", err)
				return
			}

			input.Respond(smInputs.NewMempoolStateRequestResults(
				newState,
				newCOB.getBlocksFrom(index),
				oldCOB.getBlocksFrom(index),
			))
		}

		for commonIndex > 0 {
			if oldCOB.getL1Commitment(commonIndex).Equals(newCOB.getL1Commitment(commonIndex)) {
				respondToMempoolFun(commonIndex)
				return
			}
			commonIndex--
		}
		respondToMempoolFun(0)
	}
	respondIfNeededFun := func() {
		if oldNewContainer.oldStateBlockRequestCompleted && oldNewContainer.newStateBlockRequestCompleted {
			respondFun()
		}
	}
	respondFromOldFun := func(_ obtainStateFun) {
		oldNewContainer.oldStateBlockRequestCompleted = true
		respondIfNeededFun()
	}
	respondFromNewFun := func(obtainStateFun obtainStateFun) {
		oldNewContainer.newStateBlockRequestCompleted = true
		oldNewContainer.obtainNewStateFun = obtainStateFun
		respondIfNeededFun()
	}
	id := blockRequestID(5) //TODO
	oldNewContainer.oldStateBlockRequest = newStateBlockRequestFromMempool("old", input.GetOldL1Commitment(), isValidFun, respondFromOldFun, smT.log, id)
	oldNewContainer.newStateBlockRequest = newStateBlockRequestFromMempool("new", input.GetNewL1Commitment(), isValidFun, respondFromNewFun, smT.log, id)

	result := gpa.NoMessages()
	messages, err := smT.traceBlockChainByRequest(oldNewContainer.oldStateBlockRequest)
	if err != nil {
		// TODO
	}
	result.AddAll(messages)
	messages, err = smT.traceBlockChainByRequest(oldNewContainer.newStateBlockRequest)
	if err != nil {
		// TODO
	}
	result.AddAll(messages)
	return result
}

func (smT *stateManagerGPA) createStateBlockRequestLocal(commitment *state.L1Commitment, respondFun respondFun) *stateBlockRequest {
	smT.lastBlockRequestID++
	return newStateBlockRequestLocal(commitment, respondFun, smT.log, smT.lastBlockRequestID)
}

func (smT *stateManagerGPA) getBlock(commitment *state.L1Commitment) state.Block {
	block := smT.blockCache.GetBlock(commitment)
	if block != nil {
		smT.log.Debugf("Block %s retrieved from cache", commitment)
		return block
	}
	smT.log.Debugf("Block %s is not in cache", commitment)

	// Check in store (DB).
	if !smT.store.HasTrieRoot(commitment.GetTrieRoot()) {
		smT.log.Debugf("Block %s not found in the DB", commitment)
		return nil
	}
	var err error
	block, err = smT.store.BlockByTrieRoot(commitment.GetTrieRoot())
	if err != nil {
		smT.log.Errorf("Loading block %s from the DB failed: %v", commitment, err)
		return nil
	}
	if !commitment.GetBlockHash().Equals(block.Hash()) {
		smT.log.Errorf("Block %s loaded from the database has hash %s",
			commitment, block.Hash())
		return nil
	}
	if !commitment.GetTrieRoot().Equals(block.TrieRoot()) {
		smT.log.Errorf("Block %s loaded from the database has trie root %s",
			commitment, block.TrieRoot())
		return nil
	}
	smT.log.Debugf("Block %s retrieved from the DB", commitment)
	smT.blockCache.AddBlock(block)
	return block
}

func (smT *stateManagerGPA) traceBlockChainByRequest(request blockRequest) (gpa.OutMessages, error) {
	lastCommitment := request.getLastL1Commitment()
	smT.log.Debugf("Request %s id %v tracing block %s chain...", request.getType(), request.getID(), lastCommitment)
	if smT.store.HasTrieRoot(lastCommitment.GetTrieRoot()) {
		smT.log.Debugf("Request %s id %v tracing block %s chain: the block is already in the store, marking request completed",
			request.getType(), request.getID(), lastCommitment)
		smT.markRequestCompleted(request)
		return nil, nil // No messages to send
	}
	requestsWC, ok := smT.blockRequests[lastCommitment.GetBlockHash()]
	if ok {
		requests := requestsWC.blockRequests
		smT.log.Debugf("Request %s id %v tracing block %s chain: ~s request(s) are already waiting for block %s; adding this request to the list",
			request.getType(), request.getID(), len(requests), lastCommitment)
		requestsWC.blockRequests = append(requests, request)
		return nil, nil // No messages to send
	}
	return smT.traceBlockChain(lastCommitment, []blockRequest{request})
}

// TODO: state manager may ask for several requests at once: the request can be formulated
//
//	as "give me blocks from some commitment till some index". If the requested
//	node has the required block committed into the store, it certainly has
//	all the blocks before it.
func (smT *stateManagerGPA) traceBlockChain(initCommitment *state.L1Commitment, requests []blockRequest) (gpa.OutMessages, error) {
	smT.log.Debugf("Tracing block %s chain...", initCommitment)
	commitment := initCommitment
	for !smT.store.HasTrieRoot(commitment.GetTrieRoot()) && !commitment.Equals(state.OriginL1Commitment()) {
		smT.log.Debugf("Tracing block %s chain: block %s is not in store and is not an origin block",
			initCommitment, commitment)
		block := smT.blockCache.GetBlock(commitment)
		if block == nil {
			smT.log.Debugf("Tracing block %s chain: block %s is missing", initCommitment, commitment)
			// Mark that the requests are waiting for `blockHash` block
			blockHash := commitment.GetBlockHash()
			currrentRequestsWC, ok := smT.blockRequests[blockHash]
			if !ok {
				smT.blockRequests[blockHash] = &blockRequestsWithCommitment{
					l1Commitment:  commitment,
					blockRequests: requests,
				}
				smT.log.Debugf("Tracing block %s chain completed: %v requests waiting for block %s, no requests was waiting before",
					initCommitment, len(requests), commitment)
			} else {
				currentRequests := currrentRequestsWC.blockRequests
				oldLen := len(currentRequests)
				currrentRequestsWC.blockRequests = append(currentRequests, requests...)
				smT.log.Debugf("Tracing block %s chain completed: %v requests waiting for block %s is missing, %v requests were waiting for it before",
					initCommitment, len(currrentRequestsWC.blockRequests), commitment, oldLen)
			}
			return smT.makeGetBlockRequestMessages(commitment), nil
		}
		for _, request := range requests {
			request.blockAvailable(block)
		}
		commitment = block.PreviousL1Commitment()
	}

	smT.log.Debugf("Tracing block %s chain: tracing completed, committing the blocks", initCommitment)
	committedBlocks := make(map[state.BlockHash]bool)
	for _, request := range requests {
		smT.log.Debugf("Tracing block %s chain: commiting blocks of %s request %v", initCommitment, request.getType(), request.getID())
		blockChain := request.getBlockChain()
		for i := len(blockChain) - 1; i >= 0; i-- {
			block := blockChain[i]
			blockCommitment := block.L1Commitment()
			_, ok := committedBlocks[blockCommitment.GetBlockHash()]
			if ok {
				smT.log.Debugf("Tracing block %s chain: block %s is already committed", initCommitment, blockCommitment)
			} else {
				smT.log.Debugf("Tracing block %s chain: committing block %s...", initCommitment, blockCommitment)
				var stateDraft state.StateDraft
				previousCommitment := block.PreviousL1Commitment()
				var err error
				stateDraft, err = smT.store.NewEmptyStateDraft(previousCommitment)
				if err != nil {
					return nil, fmt.Errorf("Tracing block %s chain: error creating empty state draft to store block %s: %v",
						initCommitment, blockCommitment, err)
				}
				block.Mutations().ApplyTo(stateDraft)
				committedBlock := smT.store.Commit(stateDraft)
				committedCommitment := committedBlock.L1Commitment()
				if !committedCommitment.Equals(blockCommitment) {
					smT.log.Panicf("Tracing block %s chain: block, received after committing (%s) differs from the block, which was committed (%s)",
						initCommitment, committedCommitment, blockCommitment)
				}
				smT.log.Debugf("Tracing block %s chain: block %s, index %v has been committed to the store on state %s",
					initCommitment, blockCommitment, stateDraft.BlockIndex(), previousCommitment)
				committedBlocks[blockCommitment.GetBlockHash()] = true
			}
		}
	}

	smT.log.Debugf("Tracing block %s chain: committing blocks completed, marking all the requests as completed", initCommitment)
	for _, request := range requests {
		smT.markRequestCompleted(request)
	}
	smT.log.Debugf("Tracing block %s chain done", initCommitment)
	return nil, nil // No messages to send
}

// Make `numberOfNodesToRequestBlockFromConst` messages to random peers
func (smT *stateManagerGPA) makeGetBlockRequestMessages(commitment *state.L1Commitment) gpa.OutMessages {
	nodeIDs := smT.nodeRandomiser.GetRandomOtherNodeIDs(numberOfNodesToRequestBlockFromConst)
	smT.log.Debugf("Requesting block %s from %v random nodes %v", commitment.GetBlockHash(), numberOfNodesToRequestBlockFromConst, nodeIDs)
	response := gpa.NoMessages()
	for _, nodeID := range nodeIDs {
		response.Add(smMessages.NewGetBlockMessage(commitment, nodeID))
	}
	return response
}

func (smT *stateManagerGPA) markRequestCompleted(request blockRequest) {
	request.markCompleted(func() (state.State, error) {
		return smT.store.StateByTrieRoot(request.getLastL1Commitment().GetTrieRoot())
	})
}

func (smT *stateManagerGPA) handleStateManagerTimerTick(now time.Time) gpa.OutMessages {
	result := gpa.NoMessages()
	smT.log.Debugf("State manager timer tick %v input received...", now)
	nextGetBlocksTime := smT.lastGetBlocksTime.Add(smT.timers.StateManagerGetBlockRetry)
	if now.After(nextGetBlocksTime) {
		smT.log.Debugf("State manager timer tick %v: resending get block messages...", now)
		for _, blockRequestsWC := range smT.blockRequests {
			result.AddAll(smT.makeGetBlockRequestMessages(blockRequestsWC.l1Commitment))
		}
		smT.lastGetBlocksTime = now
	} else {
		smT.log.Debugf("State manager timer tick %v: no need to resend get block messages; next resend at %v",
			now, nextGetBlocksTime)
	}
	nextCleanBlockCacheTime := smT.lastCleanBlockCacheTime.Add(smT.timers.BlockCacheBlockCleaningPeriod)
	if now.After(nextCleanBlockCacheTime) {
		smT.log.Debugf("State manager timer tick %v: cleaning block cache...", now)
		smT.blockCache.CleanOlderThan(now.Add(-smT.timers.BlockCacheBlocksInCacheDuration))
		smT.lastCleanBlockCacheTime = now
	} else {
		smT.log.Debugf("State manager timer tick %v: no need to clean block cache; next clean at %v", now, nextCleanBlockCacheTime)
	}
	nextCleanRequestsTime := smT.lastCleanRequestsTime.Add(smT.timers.StateManagerRequestCleaningPeriod)
	if now.After(nextCleanRequestsTime) {
		smT.log.Debugf("State manager timer tick %v: cleaning requests...", now)
		newBlockRequestsMap := make(map[state.BlockHash](*blockRequestsWithCommitment)) //nolint:gocritic
		for blockHash, blockRequestsWC := range smT.blockRequests {
			commitment := blockRequestsWC.l1Commitment
			blockRequests := blockRequestsWC.blockRequests
			smT.log.Debugf("State manager timer tick %v: checking %v requests waiting for block %v",
				now, len(blockRequests), commitment)
			outI := 0
			for _, blockRequest := range blockRequests {
				if blockRequest.isValid() { // Request is valid - keeping it
					blockRequests[outI] = blockRequest
					outI++
				} else {
					smT.log.Debugf("State manager timer tick %v: deleting %s request %v as it is no longer valid",
						now, blockRequest.getType(), blockRequest.getID())
				}
			}
			for i := outI; i < len(blockRequests); i++ {
				blockRequests[i] = nil // Not needed requests at the end - freeing memory
			}
			blockRequests = blockRequests[:outI]
			if len(blockRequests) > 0 {
				smT.log.Debugf("State manager timer tick %v: %v requests remaining waiting for block %v", now, len(blockRequests), commitment)
				newBlockRequestsMap[blockHash] = &blockRequestsWithCommitment{
					l1Commitment:  commitment,
					blockRequests: blockRequests,
				}
			} else {
				smT.log.Debugf("State manager timer tick %v: no more requests waiting for block %v", now, commitment)
			}
		}
		smT.blockRequests = newBlockRequestsMap
		smT.lastCleanRequestsTime = now
	} else {
		smT.log.Debugf("State manager timer tick %v: no need to clean requests; next clean at %v", now, nextCleanRequestsTime)
	}
	smT.log.Debugf("State manager timer tick %v input handled", now)
	return result
}
