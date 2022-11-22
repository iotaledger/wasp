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
	case *smInputs.ConsensusBlockProduced: // From chain
		return smT.handleConsensusBlockProduced(inputCasted)
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
	smT.log.Debugf("Message received from peer %s: request to get block %s", from, commitment)
	block := smT.getBlock(commitment)
	if block == nil {
		smT.log.Debugf("Block %s not found, request ignored", commitment)
		return nil // No messages to send
	}
	smT.log.Debugf("Block %s found, sending it to peer %s", commitment, from)
	return gpa.NoMessages().Add(smMessages.NewBlockMessage(block, from))
}

func (smT *stateManagerGPA) handlePeerBlock(from gpa.NodeID, block state.Block) gpa.OutMessages {
	blockCommitment := block.L1Commitment()
	smT.log.Debugf("Message received from peer %s: block %s", from, blockCommitment)
	requestsWC, ok := smT.blockRequests[blockCommitment.GetBlockHash()]
	if !ok {
		smT.log.Debugf("Block %s is not needed, ignoring it", blockCommitment)
		return nil // No messages to send
	}
	requests := requestsWC.blockRequests
	smT.log.Debugf("%v requests are waiting for block %s", len(requests), blockCommitment)
	err := smT.blockCache.AddBlock(block)
	if err != nil {
		return nil // No messages to send
	}
	request := smT.createStateBlockRequestLocal(blockCommitment, func(_ obtainStateFun) {})
	newRequests := append(requests, request)
	delete(smT.blockRequests, blockCommitment.GetBlockHash())
	for _, request := range newRequests {
		request.blockAvailable(block)
	}
	messages, err := smT.traceBlockChain(block.PreviousL1Commitment(), newRequests)
	if err != nil {
		return nil // No messages to send
	}
	return messages
}

func (smT *stateManagerGPA) handleChainReceiveConfirmedAliasOutput(aliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	aliasOutputID := isc.OID(aliasOutput.ID())
	smT.log.Debugf("Input received: chain confirmed alias output %s", aliasOutputID)
	stateCommitment, err := state.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		smT.log.Errorf("Error retrieving state commitment from alias output %s: %v", aliasOutputID, err)
		return nil // No messages to send
	}
	// `lastBlockRequestSLocalID` is used tu ensure that older state outputs don't
	// overwrite the newer ones. It is assumed that this method receives alias outputs
	// in strictly the same order as they are approved by L1. Moreover, it serves
	// as an ID of the request.
	smT.lastStateOutputSeq++
	seq := smT.lastStateOutputSeq
	smT.log.Debugf("Alias output %s is %v-th in state manager", aliasOutputID, seq)
	request := smT.createStateBlockRequestLocal(stateCommitment, func(_ obtainStateFun) {
		smT.log.Debugf("State for alias output %s (%v-th in state manager) is ready", aliasOutputID, seq)
		if seq <= smT.stateOutputSeq {
			smT.log.Warnf("Alias output %s (%v-th in state manager) is outdated", aliasOutputID, seq)
		} else {
			smT.log.Debugf("STATE CHANGE: state manager is at state index %v, commitment %s, alias output %s (%v-th in state manager)",
				aliasOutput.GetStateIndex(), stateCommitment, aliasOutputID, seq)
			smT.stateOutputSeq = seq
		}
	})
	messages, err := smT.traceBlockChainByRequest(request)
	if err != nil {
		smT.log.Errorf("Error tracing")
		return nil // No messages to send
	}
	return messages
}

func (smT *stateManagerGPA) handleConsensusStateProposal(csp *smInputs.ConsensusStateProposal) gpa.OutMessages {
	smT.log.Debugf("Input received: consensus state proposal for commitment %s", csp.GetL1Commitment())
	smT.lastBlockRequestID++
	request := newStateBlockRequestFromConsensusStateProposal(csp, smT.log, smT.lastBlockRequestID)
	messages, err := smT.traceBlockChainByRequest(request)
	if err != nil {
		smT.log.Errorf("Error handleConsensusStateProposal")
		return nil // No messages to send
	}
	return messages
}

func (smT *stateManagerGPA) handleConsensusDecidedState(cds *smInputs.ConsensusDecidedState) gpa.OutMessages {
	smT.log.Debugf("Input received: consensus request for decided state for commitment %s", cds.GetL1Commitment())
	smT.lastBlockRequestID++
	messages, err := smT.traceBlockChainByRequest(newStateBlockRequestFromConsensusDecidedState(cds, smT.log, smT.lastBlockRequestID))
	if err != nil {
		smT.log.Errorf("Error handleConsensusDecidedState")
		return nil // No messages to send
	}
	return messages
}

func (smT *stateManagerGPA) handleConsensusBlockProduced(input *smInputs.ConsensusBlockProduced) gpa.OutMessages {
	commitment := input.GetStateDraft().BaseL1Commitment()
	smT.log.Debugf("Input received: consensus block produced on state %s", commitment)
	request := smT.createStateBlockRequestLocal(commitment, func(_ obtainStateFun) {
		block := smT.store.Commit(input.GetStateDraft())
		smT.log.Debugf("State draft on %s has been committed to the store; block %s received",
			input.GetStateDraft().BaseL1Commitment(), block.L1Commitment())
		smT.blockCache.AddBlock(block)
		requestsWC, ok := smT.blockRequests[block.Hash()]
		if ok {
			for _, request := range requestsWC.blockRequests {
				smT.markRequestCompleted(request)
			}
		}
	})
	messages, err := smT.traceBlockChainByRequest(request)
	input.Respond(err)
	return messages
}

func (smT *stateManagerGPA) createStateBlockRequestLocal(commitment *state.L1Commitment, respondFun respondFun) *stateBlockRequest {
	smT.lastBlockRequestID++
	return newStateBlockRequestLocal(commitment, respondFun, smT.log, smT.lastBlockRequestID)
}

func (smT *stateManagerGPA) getBlock(commitment *state.L1Commitment) state.Block {
	block := smT.getBlock(commitment)
	if block != nil {
		smT.log.Debugf("Block %s retrieved from cache", commitment)
		return block
	}
	smT.log.Debugf("Block %s is not in cache", commitment)

	// Check in store (DB).
	var err error
	block, err = smT.store.BlockByTrieRoot(commitment.GetTrieRoot())
	if err != nil {
		smT.log.Errorf("Loading block %s from the DB failed: %v", commitment, err)
		return nil
	}
	if block == nil {
		smT.log.Debugf("Block %s not found in the DB", commitment)
		return nil
	}
	if !commitment.GetBlockHash().Equals(block.Hash()) {
		smT.log.Errorf("Block %s loaded from the database has hash %s",
			commitment, block.Hash())
		return nil
	}
	if !state.EqualCommitments(commitment.GetTrieRoot(), block.TrieRoot()) {
		smT.log.Errorf("Block %s loaded from the database has trie root %s",
			commitment.GetTrieRoot(), block.TrieRoot())
		return nil
	}
	smT.log.Debugf("Block %s retrieved from the DB", commitment)
	smT.blockCache.AddBlock(block)
	return block
}

func (smT *stateManagerGPA) traceBlockChainByRequest(request blockRequest) (gpa.OutMessages, error) {
	lastCommitment := request.getLastL1Commitment()
	smT.log.Debugf("Tracing the chain of blocks ending with block %s", lastCommitment)
	if smT.store.HasTrieRoot(lastCommitment.GetTrieRoot()) {
		smT.log.Debugf("Block %s is already in the store, marking request completed", lastCommitment)
		smT.markRequestCompleted(request)
		return nil, nil // No messages to send
	}
	requestsWC, ok := smT.blockRequests[lastCommitment.GetBlockHash()]
	requests := requestsWC.blockRequests
	if ok {
		smT.log.Debugf("~s request(s) are already waiting for block %s; adding this request to the list", len(requests), lastCommitment)
		requestsWC.blockRequests = append(requests, request)
		return nil, nil // No messages to send
	}
	return smT.traceBlockChain(lastCommitment, []blockRequest{request})
}

// TODO: state manager may ask for several requests at once: the request can be formulated
//		 as "give me blocks from some commitment till some index". If the requested
//		 node has the required block committed into the store, it certainly has
//		 all the blocks before it.
func (smT *stateManagerGPA) traceBlockChain(initCommitment *state.L1Commitment, requests []blockRequest) (gpa.OutMessages, error) {
	smT.log.Debugf("Tracing block %s chain...", initCommitment)
	commitment := initCommitment
	for !smT.store.HasTrieRoot(commitment.GetTrieRoot()) && !commitment.Equals(state.OriginL1Commitment()) {
		smT.log.Debugf("Tracing block %s chain: block %s is not stored and is not an origin block",
			initCommitment, commitment)
		block := smT.getBlock(commitment)
		if block == nil {
			// Mark that the requests are waiting for `blockHash` block
			blockHash := commitment.GetBlockHash()
			currrentRequestsWC, ok := smT.blockRequests[blockHash]
			if !ok {
				smT.log.Debugf("Block %s is missing, it is the first request waiting for it", commitment)
				smT.blockRequests[blockHash] = &blockRequestsWithCommitment{
					l1Commitment:  commitment,
					blockRequests: requests,
				}
			} else {
				currrentRequests := currrentRequestsWC.blockRequests
				smT.log.Debugf("Tracing block %s chain completed: Block %s is missing, %v requests are waiting for it in addition to this one", initCommitment, commitment, len(currrentRequests))
				currrentRequestsWC.blockRequests = append(currrentRequests, requests...)
			}
			return smT.makeGetBlockRequestMessages(commitment), nil
		}
		for _, request := range requests {
			request.blockAvailable(block)
		}
		commitment = block.PreviousL1Commitment()
	}

	smT.log.Debugf("Tracing block %s chain completed, committing the blocks", initCommitment)
	committedBlocks := make(map[state.BlockHash]bool)
	for _, request := range requests {
		blockChain := request.getBlockChain()
		for i := len(blockChain) - 1; i >= 0; i-- {
			block := blockChain[i]
			blockCommitment := block.L1Commitment()
			_, ok := committedBlocks[blockCommitment.GetBlockHash()]
			if !ok {
				var stateType string
				var stateDraft state.StateDraft
				previousCommitment := block.PreviousL1Commitment()
				if previousCommitment.Equals(state.OriginL1Commitment()) {
					stateDraft = smT.store.NewOriginStateDraft()
					stateType = "origin state"
				} else {
					var err error
					stateDraft, err = smT.store.NewEmptyStateDraft(previousCommitment)
					if err != nil {
						return nil, fmt.Errorf("Error creating empty state draft to store block %s: %v", blockCommitment, err)
					}
					stateType = fmt.Sprintf("state (%s, index %v)", previousCommitment, stateDraft.BlockIndex())
				}
				block.Mutations().ApplyTo(stateDraft)
				committedBlock := smT.store.Commit(stateDraft)
				committedBlockHash := committedBlock.Hash()
				if !committedBlockHash.Equals(blockCommitment.GetBlockHash()) {
					smT.log.Panicf("Block, received after committing (%s) differs from the block, which was committed (%s)",
						committedBlockHash, blockCommitment.GetBlockHash())
				}
				smT.log.Debugf("Block %s on %s has been committed to the store", blockCommitment, stateType)
				committedBlocks[blockCommitment.GetBlockHash()] = true
			}
		}
	}

	smT.log.Debugf("Committing blocks of block %s chain completed, marking all the requests as completed", initCommitment)
	for _, request := range requests {
		smT.markRequestCompleted(request)
	}
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
	smT.log.Debugf("Input received: timer tick %v", now)
	nextGetBlocksTime := smT.lastGetBlocksTime.Add(smT.timers.StateManagerGetBlockRetry)
	if now.After(nextGetBlocksTime) {
		smT.log.Debugf("Timer tick: resending get block messages...")
		for _, blockRequestsWC := range smT.blockRequests {
			result.AddAll(smT.makeGetBlockRequestMessages(blockRequestsWC.l1Commitment))
		}
		smT.lastGetBlocksTime = now
	} else {
		smT.log.Debugf("Timer tick: no need to resend get block messages; next resend at %v", nextGetBlocksTime)
	}
	nextCleanBlockCacheTime := smT.lastCleanBlockCacheTime.Add(smT.timers.BlockCacheBlockCleaningPeriod)
	if now.After(nextCleanBlockCacheTime) {
		smT.log.Debugf("Timer tick: cleaning block cache...")
		smT.blockCache.CleanOlderThan(now.Add(-smT.timers.BlockCacheBlocksInCacheDuration))
		smT.lastCleanBlockCacheTime = now
	} else {
		smT.log.Debugf("Timer tick: no need to clean block cache; next clean at %v", nextCleanBlockCacheTime)
	}
	nextCleanRequestsTime := smT.lastCleanRequestsTime.Add(smT.timers.StateManagerRequestCleaningPeriod)
	if now.After(nextCleanRequestsTime) {
		smT.log.Debugf("Timer tick: cleaning requests...")
		newBlockRequestsMap := make(map[state.BlockHash](*blockRequestsWithCommitment)) //nolint:gocritic
		for blockHash, blockRequestsWC := range smT.blockRequests {
			blockRequests := blockRequestsWC.blockRequests
			smT.log.Debugf("Timer tick: checking %v requests waiting for block %v", len(blockRequests), blockRequestsWC.l1Commitment)
			outI := 0
			for _, blockRequest := range blockRequests {
				if blockRequest.isValid() { // Request is valid - keeping it
					blockRequests[outI] = blockRequest
					outI++
				} else {
					smT.log.Debugf("Timer tick: deleting %s request %v as it is no longer valid", blockRequest.getType(), blockRequest.getID())
				}
			}
			for i := outI; i < len(blockRequests); i++ {
				blockRequests[i] = nil // Not needed requests at the end - freeing memory
			}
			blockRequests = blockRequests[:outI]
			if len(blockRequests) > 0 {
				smT.log.Debugf("Timer tick: %v requests remaining waiting for block %v", len(blockRequests), blockHash)
				newBlockRequestsMap[blockHash] = &blockRequestsWithCommitment{
					l1Commitment:  blockRequestsWC.l1Commitment,
					blockRequests: blockRequests,
				}
			} else {
				smT.log.Debugf("Timer tick: no more requests waiting for block %v", blockHash)
			}
		}
		smT.blockRequests = newBlockRequestsMap
		smT.lastCleanRequestsTime = now
	} else {
		smT.log.Debugf("Timer tick: no need to clean requests; next clean at %v", nextCleanRequestsTime)
	}
	smT.log.Debugf("Timer tick %v handled", now)
	return result
}
