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

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smMessages"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smUtils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type stateManagerGPA struct {
	log                     *logger.Logger
	chainID                 isc.ChainID
	blockCache              smGPAUtils.BlockCache
	blocksToFetch           blockFetchers
	blocksFetched           blockFetchers
	nodeRandomiser          smUtils.NodeRandomiser
	store                   state.Store
	timers                  StateManagerTimers
	lastGetBlocksTime       time.Time
	lastCleanBlockCacheTime time.Time
	lastCleanRequestsTime   time.Time
}

var _ gpa.GPA = &stateManagerGPA{}

const (
	numberOfNodesToRequestBlockFromConst = 5
)

func New(
	chainID isc.ChainID,
	nr smUtils.NodeRandomiser,
	wal smGPAUtils.BlockWAL,
	store state.Store,
	log *logger.Logger,
	timers StateManagerTimers,
) (gpa.GPA, error) {
	var err error
	smLog := log.Named("gpa")
	blockCache, err := smGPAUtils.NewBlockCache(timers.TimeProvider, timers.BlockCacheMaxSize, wal, smLog)
	if err != nil {
		smLog.Errorf("Error creating block cache: %v", err)
		return nil, err
	}
	result := &stateManagerGPA{
		log:                     smLog,
		chainID:                 chainID,
		blockCache:              blockCache,
		blocksToFetch:           newBlockFetchers(),
		blocksFetched:           newBlockFetchers(),
		nodeRandomiser:          nr,
		store:                   store,
		timers:                  timers,
		lastGetBlocksTime:       time.Time{},
		lastCleanBlockCacheTime: time.Time{},
	}

	return result, nil
}

// -------------------------------------
// Implementation for gpa.GPA interface
// -------------------------------------

func (smT *stateManagerGPA) Input(input gpa.Input) gpa.OutMessages {
	switch inputCasted := input.(type) {
	case *smInputs.ConsensusStateProposal: // From consensus
		return smT.handleConsensusStateProposal(inputCasted)
	case *smInputs.ConsensusDecidedState: // From consensus
		return smT.handleConsensusDecidedState(inputCasted)
	case *smInputs.ConsensusBlockProduced: // From consensus
		return smT.handleConsensusBlockProduced(inputCasted)
	case *smInputs.ChainFetchStateDiff: // From mempool
		return smT.handleChainFetchStateDiff(inputCasted)
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
	return fmt.Sprintf(
		"State manager is waiting for %v blocks from other noded; "+
			"%v blocks are obtained and waiting to be committed; "+
			"last time blocks were requested from peer nodes: %v (every %v); "+
			"last time outdated requests were cleared: %v (every %v); "+
			"last time block cache was cleaned: %v (every %v).",
		smT.blocksToFetch.getSize(),
		smT.blocksFetched.getSize(),
		util.TimeOrNever(smT.lastGetBlocksTime), smT.timers.StateManagerGetBlockRetry,
		util.TimeOrNever(smT.lastCleanRequestsTime), smT.timers.StateManagerRequestCleaningPeriod,
		util.TimeOrNever(smT.lastCleanBlockCacheTime), smT.timers.BlockCacheBlockCleaningPeriod,
	)
}

func (smT *stateManagerGPA) UnmarshalMessage(data []byte) (gpa.Message, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("error unmarshalling message: slice of length %d is too short", len(data))
	}
	var message gpa.Message
	switch data[0] {
	case smMessages.MsgTypeBlockMessage:
		message = smMessages.NewEmptyBlockMessage()
	case smMessages.MsgTypeGetBlockMessage:
		message = smMessages.NewEmptyGetBlockMessage()
	default:
		return nil, fmt.Errorf("error unmarshalling message: message type %v unknown", data[0])
	}
	err := message.UnmarshalBinary(data)
	return message, err
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smT *stateManagerGPA) handlePeerGetBlock(from gpa.NodeID, commitment *state.L1Commitment) gpa.OutMessages {
	// TODO: [KP] Only accept queries from access nodes.
	fromLog := from.ShortString()
	smT.log.Debugf("Message GetBlock %s received from peer %s", commitment, fromLog)
	block := smT.getBlock(commitment)
	if block == nil {
		smT.log.Debugf("Message GetBlock %s: block not found, peer %s request ignored", commitment, fromLog)
		return nil // No messages to send
	}
	smT.log.Debugf("Message GetBlock %s: block found, sending it to peer %s", commitment, fromLog)
	return gpa.NoMessages().Add(smMessages.NewBlockMessage(block, from))
}

func (smT *stateManagerGPA) handlePeerBlock(from gpa.NodeID, block state.Block) gpa.OutMessages {
	blockCommitment := block.L1Commitment()
	fromLog := from.ShortString()
	smT.log.Debugf("Message Block %s received from peer %s", blockCommitment, fromLog)
	fetcher := smT.blocksToFetch.takeFetcher(blockCommitment)
	if fetcher == nil {
		smT.log.Debugf("Message Block %s: block is not needed, ignoring it", blockCommitment)
		return nil // No messages to send
	}
	smT.blockCache.AddBlock(block)
	callback := newBlockRequestCallback(brcAlwaysValidFun, brcIgnoreRequestCompletedFun)
	fetcher.addCallback(callback) // TODO: maybe it is not needed?
	messages, err := smT.traceBlockChain(fetcher)
	if err != nil {
		return nil // No messages to send
	}
	smT.log.Debugf("Message Block %s from peer %s handled", blockCommitment, fromLog)
	return messages
}

func (smT *stateManagerGPA) handleConsensusStateProposal(csp *smInputs.ConsensusStateProposal) gpa.OutMessages {
	smT.log.Debugf("Input consensus state proposal %s received...", csp.GetL1Commitment())
	callback := newBlockRequestCallback(
		func() bool {
			return csp.IsValid()
		},
		func() {
			csp.Respond()
		},
	)
	messages, err := smT.traceBlockChainWithCallback(csp.GetL1Commitment(), callback)
	if err != nil {
		smT.log.Errorf("Input consensus state proposal %s: error tracing block chain: %v", csp.GetL1Commitment(), err)
		return nil // No messages to send
	}
	smT.log.Debugf("Input consensus state proposal %s handled", csp.GetL1Commitment())
	return messages
}

func (smT *stateManagerGPA) handleConsensusDecidedState(cds *smInputs.ConsensusDecidedState) gpa.OutMessages {
	smT.log.Debugf("Input consensus decided state %s received...", cds.GetL1Commitment())
	callback := newBlockRequestCallback(
		func() bool {
			return cds.IsValid()
		},
		func() {
			state, err := smT.store.StateByTrieRoot(cds.GetL1Commitment().TrieRoot())
			if err != nil {
				smT.log.Errorf("error obtaining state: %w", err)
				return
			}
			cds.Respond(state)
		},
	)
	messages, err := smT.traceBlockChainWithCallback(cds.GetL1Commitment(), callback)
	if err != nil {
		smT.log.Errorf("Input consensus decided state %s: error tracing block chain: %v", cds.GetL1Commitment(), err)
		return nil // No messages to send
	}
	smT.log.Debugf("Input consensus decided state %s handled", cds.GetL1Commitment())
	return messages
}

func (smT *stateManagerGPA) handleConsensusBlockProduced(input *smInputs.ConsensusBlockProduced) gpa.OutMessages {
	commitment := input.GetStateDraft().BaseL1Commitment()
	smT.log.Debugf("Input block produced on state %s received...", commitment)
	if !smT.store.HasTrieRoot(commitment.TrieRoot()) {
		smT.log.Panicf("Input block produced on state %s: state, on which this block is produced, is not yet in the store", commitment)
	}
	// NOTE: committing already committed block is allowed (see `TestDoubleCommit` test in `packages/state/state_test.go`)
	block := smT.store.Commit(input.GetStateDraft())
	blockCommitment := block.L1Commitment()
	smT.log.Debugf("Input block produced on state %s: state draft index %v has been committed to the store, resulting block %s",
		commitment, input.GetStateDraft().BlockIndex(), blockCommitment)
	smT.blockCache.AddBlock(block)
	input.Respond(block)
	fetcher := smT.blocksToFetch.takeFetcher(blockCommitment)
	if fetcher != nil {
		result, err := smT.markFetched(fetcher)
		if err != nil {
			smT.log.Errorf("Input block produced on state %s: failed to mark block fetched: %w",
				commitment, err)
		} else {
			smT.log.Debugf("Input block produced on state %s handled", commitment)
			return result
		}
	}
	smT.log.Debugf("Input block produced on state %s handled", commitment)
	return nil // No messages to send
}

func (smT *stateManagerGPA) handleChainFetchStateDiff(input *smInputs.ChainFetchStateDiff) gpa.OutMessages { //nolint:funlen
	smT.log.Debugf("Input mempool state request for state (index %v, %s) is received compared to state (index %v, %s)...",
		input.GetNewStateIndex(), input.GetNewL1Commitment(), input.GetOldStateIndex(), input.GetOldL1Commitment())
	oldBlockRequestCompleted := false
	newBlockRequestCompleted := false
	isValidFun := func() bool { return input.IsValid() }
	obtainCommittedBlockFun := func(commitment *state.L1Commitment) state.Block {
		result := smT.getBlock(commitment)
		if result == nil {
			smT.log.Panicf("Input mempool state request for state (index %v, %s): cannot obtain block %s", input.GetNewStateIndex(), input.GetNewL1Commitment(), commitment)
		}
		return result
	}
	lastBlockFun := func(blocks []state.Block) state.Block {
		return blocks[len(blocks)-1]
	}
	respondFun := func() {
		oldBlock := obtainCommittedBlockFun(input.GetOldL1Commitment())
		newBlock := obtainCommittedBlockFun(input.GetNewL1Commitment())
		oldChainOfBlocks := []state.Block{oldBlock}
		newChainOfBlocks := []state.Block{newBlock}
		for lastBlockFun(oldChainOfBlocks).StateIndex() > lastBlockFun(newChainOfBlocks).StateIndex() {
			oldChainOfBlocks = append(oldChainOfBlocks, obtainCommittedBlockFun(lastBlockFun(oldChainOfBlocks).PreviousL1Commitment()))
		}
		for lastBlockFun(oldChainOfBlocks).StateIndex() < lastBlockFun(newChainOfBlocks).StateIndex() {
			newChainOfBlocks = append(newChainOfBlocks, obtainCommittedBlockFun(lastBlockFun(newChainOfBlocks).PreviousL1Commitment()))
		}
		for lastBlockFun(oldChainOfBlocks).StateIndex() > 0 {
			if lastBlockFun(oldChainOfBlocks).L1Commitment().Equals(lastBlockFun(newChainOfBlocks).L1Commitment()) {
				smT.log.Debugf("Input mempool state request for state (index %v, %s): old and new blocks index %v match: %s",
					input.GetNewStateIndex(), input.GetNewL1Commitment(),
					lastBlockFun(newChainOfBlocks).StateIndex(), lastBlockFun(newChainOfBlocks).L1Commitment())
				break
			}
			smT.log.Debugf("Input mempool state request for state (index %v, %s): old (%s) and new (%s) blocks index %v do not match",
				input.GetNewStateIndex(), input.GetNewL1Commitment(),
				lastBlockFun(oldChainOfBlocks).L1Commitment(), lastBlockFun(newChainOfBlocks).L1Commitment(),
				lastBlockFun(newChainOfBlocks).StateIndex())
			oldChainOfBlocks = append(oldChainOfBlocks, obtainCommittedBlockFun(lastBlockFun(oldChainOfBlocks).PreviousL1Commitment()))
			newChainOfBlocks = append(newChainOfBlocks, obtainCommittedBlockFun(lastBlockFun(newChainOfBlocks).PreviousL1Commitment()))
		}
		oldChainOfBlocks = lo.Reverse(oldChainOfBlocks[:len(oldChainOfBlocks)-1])
		newChainOfBlocks = lo.Reverse(newChainOfBlocks[:len(newChainOfBlocks)-1])
		newState, err := smT.store.StateByTrieRoot(input.GetNewL1Commitment().TrieRoot())
		if err != nil {
			smT.log.Errorf("error obtaining state: %w", err)
			return
		}
		input.Respond(smInputs.NewChainFetchStateDiffResults(newState, newChainOfBlocks, oldChainOfBlocks))
	}
	respondIfNeededFun := func() {
		if oldBlockRequestCompleted && newBlockRequestCompleted {
			smT.log.Debugf("Input mempool state request for state (index %v, %s): both requests are completed, responding",
				input.GetNewStateIndex(), input.GetNewL1Commitment())
			respondFun()
		}
	}
	oldRequestCallback := newBlockRequestCallback(isValidFun, func() {
		oldBlockRequestCompleted = true
		respondIfNeededFun()
	})
	newRequestCallback := newBlockRequestCallback(isValidFun, func() {
		newBlockRequestCompleted = true
		respondIfNeededFun()
	})
	result := gpa.NoMessages()
	smT.log.Debugf("Input mempool state request for state (index %v, %s): tracing chain by old block request",
		input.GetNewStateIndex(), input.GetNewL1Commitment())
	messages, err := smT.traceBlockChainWithCallback(input.GetOldL1Commitment(), oldRequestCallback)
	if err != nil {
		smT.log.Errorf("Input mempool state request for state (index %v, %s): error tracing chain by old block request: %v",
			input.GetNewStateIndex(), input.GetNewL1Commitment(), err)
		return nil // No messages to send
	}
	result.AddAll(messages)
	smT.log.Debugf("Input mempool state request for state (index %v, %s): tracing chain by new block request",
		input.GetNewStateIndex(), input.GetNewL1Commitment())
	messages, err = smT.traceBlockChainWithCallback(input.GetNewL1Commitment(), newRequestCallback)
	if err != nil {
		smT.log.Errorf("Input mempool state request for state (index %v, %s): error tracing chain by new block request: %v",
			input.GetNewStateIndex(), input.GetNewL1Commitment(), err)
		return nil // No messages to send
	}
	result.AddAll(messages)
	smT.log.Debugf("Input mempool state request for state (index %v, %s) handled",
		input.GetNewStateIndex(), input.GetNewL1Commitment())
	return result
}

func (smT *stateManagerGPA) getBlock(commitment *state.L1Commitment) state.Block {
	block := smT.blockCache.GetBlock(commitment)
	if block != nil {
		smT.log.Debugf("Block %s retrieved from cache", commitment)
		return block
	}
	smT.log.Debugf("Block %s is not in cache", commitment)

	// Check in store (DB).
	if !smT.store.HasTrieRoot(commitment.TrieRoot()) {
		smT.log.Debugf("Block %s not found in the DB", commitment)
		return nil
	}
	var err error
	block, err = smT.store.BlockByTrieRoot(commitment.TrieRoot())
	if err != nil {
		smT.log.Errorf("Loading block %s from the DB failed: %v", commitment, err)
		return nil
	}
	if !commitment.BlockHash().Equals(block.Hash()) {
		smT.log.Errorf("Block %s loaded from the database has hash %s",
			commitment, block.Hash())
		return nil
	}
	if !commitment.TrieRoot().Equals(block.TrieRoot()) {
		smT.log.Errorf("Block %s loaded from the database has trie root %s",
			commitment, block.TrieRoot())
		return nil
	}
	smT.log.Debugf("Block %s retrieved from the DB", commitment)
	smT.blockCache.AddBlock(block)
	return block
}

func (smT *stateManagerGPA) traceBlockChainWithCallback(lastCommitment *state.L1Commitment, callback blockRequestCallback) (gpa.OutMessages, error) {
	smT.log.Debugf("Tracing block %s chain...", lastCommitment)
	if smT.store.HasTrieRoot(lastCommitment.TrieRoot()) {
		smT.log.Debugf("Tracing block %s chain: the block is already in the store, calling back", lastCommitment)
		callback.requestCompleted()
		return nil, nil // No messages to send
	}
	if smT.blocksToFetch.addCallback(lastCommitment, callback) {
		smT.log.Debugf("Tracing block %s chain: the block is already being fetched", lastCommitment)
		return nil, nil
	}
	if smT.blocksFetched.addCallback(lastCommitment, callback) {
		smT.log.Debugf("Tracing block %s chain: the block is already fetched, but cannot yet be committed", lastCommitment)
		return nil, nil
	}
	fetcher := newBlockFetcherWithCallback(lastCommitment, callback)
	return smT.traceBlockChain(fetcher)
}

// TODO: state manager may ask for several requests at once: the request can be
// formulated as "give me blocks from some commitment till some index". If the
// requested node has the required block committed into the store, it certainly
// has all the blocks before it.
func (smT *stateManagerGPA) traceBlockChain(fetcher blockFetcher) (gpa.OutMessages, error) {
	commitment := fetcher.getCommitment()
	smT.log.Debugf("Tracing block %s chain...", commitment)
	if !smT.store.HasTrieRoot(commitment.TrieRoot()) {
		smT.log.Debugf("Tracing block %s chain: block is not in store", commitment)
		block := smT.blockCache.GetBlock(commitment)
		if block == nil {
			smT.log.Debugf("Tracing block %s chain: block is missing", commitment)
			smT.blocksToFetch.addFetcher(fetcher)
			return smT.makeGetBlockRequestMessages(commitment), nil
		}
		previousCommitment := block.PreviousL1Commitment()
		if previousCommitment == nil {
			result, err := smT.markFetched(fetcher)
			smT.log.Debugf("Tracing block %s chain done, err=%v", commitment, err)
			return result, err
		}
		smT.blocksFetched.addFetcher(fetcher)
		if smT.blocksToFetch.addRelatedFetcher(previousCommitment, fetcher) {
			smT.log.Debugf("Tracing block %s chain: previous block %s is already being fetched", previousCommitment)
			return nil, nil
		}
		if smT.blocksFetched.addRelatedFetcher(previousCommitment, fetcher) {
			smT.log.Debugf("Tracing block %s chain: previous block %s is already fetched, but cannot yet be committed", previousCommitment)
			return nil, nil
		}
		return smT.traceBlockChain(newBlockFetcherWithRelatedFetcher(previousCommitment, fetcher))
	}
	result, err := smT.markFetched(fetcher)
	smT.log.Debugf("Tracing block %s chain done, err=%v", commitment, err)
	return result, err
}

func (smT *stateManagerGPA) markFetched(fetcher blockFetcher) (gpa.OutMessages, error) {
	result := gpa.NoMessages()
	err := fetcher.notifyFetched(func(bf blockFetcher) (bool, error) {
		commitment := bf.getCommitment()
		block := smT.blockCache.GetBlock(commitment)
		if block == nil {
			// Block was previously received but it is no longer in cache and
			// for some unexpected reasons it is not in WAL: rerequest it
			smT.log.Warnf("Block %s was previously obtained, but it can neither be found in cache nor in WAL. Rerequesting it.", commitment)
			smT.blocksToFetch.addFetcher(bf)
			result.AddAll(smT.makeGetBlockRequestMessages(commitment))
			return false, nil
		}
		// Commit block
		var stateDraft state.StateDraft
		previousCommitment := block.PreviousL1Commitment()
		if previousCommitment == nil {
			// origin block
			stateDraft = smT.store.NewOriginStateDraft()
		} else {
			var err error
			stateDraft, err = smT.store.NewEmptyStateDraft(previousCommitment)
			if err != nil {
				return false, fmt.Errorf("Error creating empty state draft to store block %s: %w", commitment, err)
			}
		}
		block.Mutations().ApplyTo(stateDraft)
		committedBlock := smT.store.Commit(stateDraft)
		committedCommitment := committedBlock.L1Commitment()
		if !committedCommitment.Equals(commitment) {
			smT.log.Panicf("Block, received after committing (%s) differs from the block, which was committed (%s)",
				committedCommitment, commitment)
		}
		smT.log.Debugf("Block index %v %s has been committed to the store on state %s",
			block.StateIndex(), commitment, previousCommitment)
		_ = smT.blocksFetched.takeFetcher(commitment)
		return true, nil
	})
	return result, err
}

// Make `numberOfNodesToRequestBlockFromConst` messages to random peers
func (smT *stateManagerGPA) makeGetBlockRequestMessages(commitment *state.L1Commitment) gpa.OutMessages {
	nodeIDs := smT.nodeRandomiser.GetRandomOtherNodeIDs(numberOfNodesToRequestBlockFromConst)
	smT.log.Debugf("Requesting block %s from %v random nodes %s", commitment, numberOfNodesToRequestBlockFromConst, util.SliceShortString(nodeIDs))
	response := gpa.NoMessages()
	for _, nodeID := range nodeIDs {
		response.Add(smMessages.NewGetBlockMessage(commitment, nodeID))
	}
	return response
}

func (smT *stateManagerGPA) handleStateManagerTimerTick(now time.Time) gpa.OutMessages {
	result := gpa.NoMessages()
	smT.log.Debugf("Input timer tick %v received...", now)
	smT.log.Debugf("Status: %s", smT.StatusString())
	nextGetBlocksTime := smT.lastGetBlocksTime.Add(smT.timers.StateManagerGetBlockRetry)
	if now.After(nextGetBlocksTime) {
		smT.log.Debugf("Input timer tick %v: resending get block messages...", now)
		for _, commitment := range smT.blocksToFetch.getCommitments() {
			result.AddAll(smT.makeGetBlockRequestMessages(commitment))
		}
		smT.lastGetBlocksTime = now
	} else {
		smT.log.Debugf("Input timer tick %v: no need to resend get block messages; next resend at %v",
			now, nextGetBlocksTime)
	}
	nextCleanBlockCacheTime := smT.lastCleanBlockCacheTime.Add(smT.timers.BlockCacheBlockCleaningPeriod)
	if now.After(nextCleanBlockCacheTime) {
		smT.log.Debugf("Input timer tick %v: cleaning block cache...", now)
		smT.blockCache.CleanOlderThan(now.Add(-smT.timers.BlockCacheBlocksInCacheDuration))
		smT.lastCleanBlockCacheTime = now
	} else {
		smT.log.Debugf("Input timer tick %v: no need to clean block cache; next clean at %v", now, nextCleanBlockCacheTime)
	}
	nextCleanRequestsTime := smT.lastCleanRequestsTime.Add(smT.timers.StateManagerRequestCleaningPeriod)
	if now.After(nextCleanRequestsTime) {
		smT.log.Debugf("Input timer tick %v: cleaning requests...", now)
		smT.blocksToFetch.cleanCallbacks()
		smT.blocksFetched.cleanCallbacks()
		smT.lastCleanRequestsTime = now
	} else {
		smT.log.Debugf("Input timer tick %v: no need to clean requests; next clean at %v", now, nextCleanRequestsTime)
	}
	smT.log.Debugf("Input timer tick %v handled", now)
	return result
}
