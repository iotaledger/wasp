package jsonrpc

/**
	The EngineAPI is irrelevant for our actual public node. The EngineAPI is mainly used in Ethereum for Consensus tasks.
	However, the hive testing tool relies on it being available and functioning.
 	Therefore, the EngineAPI is mostly 1:1 copied from Ethereum where possible. The Consensus related parts are being left out, so the API mainly does validation and enqueuing of TXs.
	All EngineAPI related functions (even pure functions) are bound to the Service implementation to not pollute the code base.
*/

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/iotaledger/wasp/packages/metrics"
)

type EngineService struct {
	evmChain *EVMChain
	metrics  *metrics.ChainWebAPIMetrics
	accounts *AccountManager
	params   *Parameters
}

func NewEngineService(evmChain *EVMChain, metrics *metrics.ChainWebAPIMetrics, accounts *AccountManager, params *Parameters) *EngineService {
	return &EngineService{
		evmChain: evmChain,
		metrics:  metrics,
		accounts: accounts,
		params:   params,
	}
}

// convertRequests converts a hex requests slice to plain [][]byte.
func (e *EngineService) convertRequests(hex []hexutil.Bytes) [][]byte {
	if hex == nil {
		return nil
	}
	req := make([][]byte, len(hex))
	for i := range hex {
		req[i] = hex[i]
	}
	return req
}

// validateRequests checks that requests are ordered by their type and are not empty.
func (e *EngineService) validateRequests(requests [][]byte) error {
	for i, req := range requests {
		// No empty requests.
		if len(req) < 2 {
			return fmt.Errorf("empty request: %v", req)
		}
		// Check that requests are ordered by their type.
		// Each type must appear only once.
		if i > 0 && req[0] <= requests[i-1][0] {
			return fmt.Errorf("invalid request order: %v", req)
		}
	}
	return nil
}

func (e *EngineService) responseInvalid(err error, latestValid *types.Header, blockHash common.Hash) engine.PayloadStatusV1 {
	var currentHash *common.Hash
	if latestValid != nil {
		if latestValid.Difficulty.BitLen() != 0 {
			// Set latest valid hash to 0x0 if parent is PoW block
			currentHash = &common.Hash{}
		} else {
			// Otherwise set latest valid hash to parent hash
			currentHash = &blockHash
		}
	}
	errorMsg := err.Error()
	return engine.PayloadStatusV1{Status: engine.INVALID, LatestValidHash: currentHash, ValidationError: &errorMsg}
}

func (e *EngineService) waitForTransactionConfirmation(transactions []*types.Transaction, timeout time.Duration) (common.Hash, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var blockHashes []common.Hash
	var blockNumbers []uint64

	wg.Add(len(transactions))

	for _, tx := range transactions {
		go func(tx *types.Transaction) {
			defer wg.Done()

			// Poll until transaction is confirmed or timeout
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			timeoutChan := time.After(timeout)

			for {
				select {
				case <-timeoutChan:
					mu.Lock()
					errors = append(errors, fmt.Errorf("timeout waiting for transaction %s", tx.Hash().Hex()))
					mu.Unlock()
					return
				case <-ticker.C:
					_, blockHash, blockNumber, _, err := e.evmChain.TransactionByHash(tx.Hash())
					if err != nil {
						continue // Keep polling
					}

					if blockNumber == 0 {
						continue // Transaction not yet mined
					}

					// Transaction confirmed
					mu.Lock()
					blockHashes = append(blockHashes, blockHash)
					blockNumbers = append(blockNumbers, blockNumber)
					mu.Unlock()
					return
				}
			}
		}(tx)
	}

	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		return common.Hash{}, errors[0]
	}

	// Verify all transactions have the same block number
	if len(blockNumbers) == 0 {
		return common.Hash{}, fmt.Errorf("no transactions confirmed")
	}

	firstBlockNumber := blockNumbers[0]
	for _, bn := range blockNumbers {
		if bn != firstBlockNumber {
			return common.Hash{}, fmt.Errorf("transactions have different block numbers: %s vs %s", firstBlockNumber, bn)
		}
	}

	// All transactions should have the same block hash
	return blockHashes[0], nil
}

func (e *EngineService) EnqueueTransactions(block *types.Block, blockHash common.Hash) (*engine.PayloadStatusV1, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	transactions := block.Transactions()
	wg.Add(len(transactions))

	// Launch goroutines for each transaction
	for _, tx := range transactions {
		go func(tx *types.Transaction) {
			defer wg.Done()

			if err := e.evmChain.SendTransaction(tx); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(tx)
	}

	// Wait for all transactions to complete
	wg.Wait()

	// Check if any errors occurred
	if len(errors) > 0 {
		fmt.Println("EnqueueTransactions err: ", errors)
		res := e.responseInvalid(errors[0], nil, block.Hash())
		return &res, nil
	}

	// Wait for transaction confirmation with 30s timeout
	blockHash, err := e.waitForTransactionConfirmation(transactions, 30*time.Second)
	if err != nil {
		fmt.Println("waitForTransactionConfirmation err: ", err)
		res := e.responseInvalid(err, nil, blockHash)
		return &res, nil
	}

	// Success - you can use blockHash as needed
	_ = blockHash // Use blockHash in your response logic

	return &engine.PayloadStatusV1{
		Status:          "VALID",
		LatestValidHash: &common.Hash{},
		ValidationError: nil,
	}, nil
}

func (e *EngineService) newPayload(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, requests [][]byte) (engine.PayloadStatusV1, error) {
	log.Trace("Engine API request received", "method", "NewPayload", "number", params.Number, "hash", params.BlockHash)
	block, err := engine.ExecutableDataToBlock(params, versionedHashes, beaconRoot, requests)
	fmt.Printf("ExecutableDataBlock  err: %v\n", err)

	if err != nil {
		bgu := "nil"
		if params.BlobGasUsed != nil {
			bgu = strconv.Itoa(int(*params.BlobGasUsed))
		}
		ebg := "nil"
		if params.ExcessBlobGas != nil {
			ebg = strconv.Itoa(int(*params.ExcessBlobGas))
		}
		log.Warn("Invalid NewPayload params",
			"params.Number", params.Number,
			"params.ParentHash", params.ParentHash,
			"params.BlockHash", params.BlockHash,
			"params.StateRoot", params.StateRoot,
			"params.FeeRecipient", params.FeeRecipient,
			"params.LogsBloom", common.PrettyBytes(params.LogsBloom),
			"params.Random", params.Random,
			"params.GasLimit", params.GasLimit,
			"params.GasUsed", params.GasUsed,
			"params.Timestamp", params.Timestamp,
			"params.ExtraData", common.PrettyBytes(params.ExtraData),
			"params.BaseFeePerGas", params.BaseFeePerGas,
			"params.BlobGasUsed", bgu,
			"params.ExcessBlobGas", ebg,
			"len(params.Transactions)", len(params.Transactions),
			"len(params.Withdrawals)", len(params.Withdrawals),
			"beaconRoot", beaconRoot,
			"len(requests)", len(requests),
			"error", err)

		return e.responseInvalid(err, nil, params.BlockHash), nil
	}
	fmt.Printf("BlockByHash requests\n")

	if block := e.evmChain.BlockByHash(block.Hash()); block != nil {
		log.Warn("Ignoring already known beacon payload", "number", params.Number, "hash", params.BlockHash, "age", common.PrettyAge(time.Unix(int64(block.Time()), 0)))
		return engine.PayloadStatusV1{Status: engine.VALID, LatestValidHash: &params.BlockHash}, nil
	}

	// For reference: NewPayload would validate the hash here by looking into the Tipsets which we don't have, as we don't use Ethereums consensus.
	// Additionally, it tries to get the ParentBlock which is used to decide if blocks get reposted.
	// This might need our own implementation, for nor leaving it out.
	/*
		// If this block was rejected previously, keep rejecting it
		if res := api.checkInvalidAncestor(block.Hash(), block.Hash()); res != nil {
			return *res, nil
		}

		// If the parent is missing, we - in theory - could trigger a sync, but that
		// would also entail a reorg. That is problematic if multiple sibling blocks
		// are being fed to us, and even more so, if some semi-distant uncle shortens
		// our live chain. As such, payload execution will not permit reorgs and thus
		// will not trigger a sync cycle. That is fine though, if we get a fork choice
		// update after legit payload executions.
		parent := api.eth.BlockChain().GetBlock(block.ParentHash(), block.NumberU64()-1)
		if parent == nil {
			return api.delayPayloadImport(block), nil
		}
	*/
	fmt.Printf("Before Parent\n")

	parent := e.evmChain.BlockByHash(block.ParentHash())
	if parent == nil {
		// Get the latest block to help with debugging
		latestBlock, err := e.evmChain.BlockByNumber(nil) // nil gets latest block
		latestHash := "none"
		latestNumber := "none"
		if err == nil && latestBlock != nil {
			latestHash = latestBlock.Hash().Hex()
			latestNumber = fmt.Sprintf("%d", latestBlock.NumberU64())
		}

		errorMsg := fmt.Sprintf("parent block not found: parentHash=%s, latestBlock=%s (number %s)",
			block.ParentHash().Hex(), latestHash, latestNumber)
		fmt.Printf("Parent validation failed: %s\n", errorMsg)

		fmt.Println("Set parent to the latest block")
		parent = latestBlock
	}

	// if block.Time() <= parent.Time() {
	// 	fmt.Println("Invalid timestamp", "parent", block.Time(), "block", block.Time())
	// 	log.Warn("Invalid timestamp", "parent", block.Time(), "block", block.Time())
	// 	return e.responseInvalid(errors.New("invalid timestamp"), block.Header(), params.BlockHash), nil
	// }

	// Additional Consensus related parts left out
	// We might need to support ACCEPTED though in some way.
	/**
			// Another corner case: if the node is in snap sync mode, but the CL client
	// tries to make it import a block. That should be denied as pushing something
	// into the database directly will conflict with the assumptions of snap sync
	// that it has an empty db that it can fill itself.
	if api.eth.SyncMode() != ethconfig.FullSync {
		return api.delayPayloadImport(block), nil
	}
	if !api.eth.BlockChain().HasBlockAndState(block.ParentHash(), block.NumberU64()-1) {
		api.remoteBlocks.put(block.Hash(), block.Header())
		log.Warn("State not available, ignoring new payload")
		return engine.PayloadStatusV1{Status: engine.ACCEPTED}, nil
	}
	*/
	log.Trace("Inserting block without sethead", "hash", block.Hash(), "number", block.Number())

	if res, _ := e.EnqueueTransactions(block, params.BlockHash); res != nil {
		return *res, nil
	}

	return engine.PayloadStatusV1{Status: engine.VALID, Witness: nil, LatestValidHash: &params.BlockHash}, nil
}

func (e *EngineService) NewPayloadV1(params engine.ExecutableData) (engine.PayloadStatusV1, error) {
	if params.Withdrawals != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("withdrawals not supported in V1"))
	}
	return e.newPayload(params, nil, nil, nil)
}

func (e *EngineService) NewPayloadV2(params engine.ExecutableData) (engine.PayloadStatusV1, error) {
	if params.ExcessBlobGas != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("non-nil excessBlobGas pre-cancun"))
	}
	if params.BlobGasUsed != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("non-nil blobGasUsed pre-cancun"))
	}
	return e.newPayload(params, nil, nil, nil)
}

func (e *EngineService) NewPayloadV3(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash) (engine.PayloadStatusV1, error) {
	if params.Withdrawals == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil withdrawals post-shanghai"))
	}
	if params.ExcessBlobGas == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil excessBlobGas post-cancun"))
	}
	if params.BlobGasUsed == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil blobGasUsed post-cancun"))
	}

	if versionedHashes == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil versionedHashes post-cancun"))
	}
	if beaconRoot == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil beaconRoot post-cancun"))
	}

	/*
		if api.eth.BlockChain().Config().LatestFork(params.Timestamp) != forks.Cancun {
			return engine.PayloadStatusV1{Status: engine.INVALID}, engine.UnsupportedFork.With(errors.New("newPayloadV3 must only be called for cancun payloads"))
		}*/

	return e.newPayload(params, versionedHashes, beaconRoot, nil)
}

func (e *EngineService) NewPayloadV4(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, executionRequests []hexutil.Bytes) (engine.PayloadStatusV1, error) {
	if params.Withdrawals == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil withdrawals post-shanghai"))
	}
	if params.ExcessBlobGas == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil excessBlobGas post-cancun"))
	}
	if params.BlobGasUsed == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil blobGasUsed post-cancun"))
	}

	if versionedHashes == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil versionedHashes post-cancun"))
	}
	if beaconRoot == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil beaconRoot post-cancun"))
	}
	if executionRequests == nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("nil executionRequests post-prague"))
	}

	requests := e.convertRequests(executionRequests)
	if err := e.validateRequests(requests); err != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(err)
	}

	return e.newPayload(params, versionedHashes, beaconRoot, requests)
}

func (e *EngineService) forkchoiceUpdated(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	if payloadAttributes != nil {
		if payloadAttributes.Withdrawals != nil || payloadAttributes.BeaconRoot != nil {
			return engine.STATUS_INVALID, engine.InvalidParams.With(errors.New("withdrawals and beacon root not supported in V1"))
		}
	}
	return engine.ForkChoiceResponse{
		PayloadStatus: engine.PayloadStatusV1{Status: engine.VALID, LatestValidHash: &update.HeadBlockHash},
		PayloadID:     nil,
	}, nil
}

func (e *EngineService) ForkchoiceUpdatedV1(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	return e.forkchoiceUpdated(update, payloadAttributes)
}

func (e *EngineService) ForkchoiceUpdatedV2(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	return e.forkchoiceUpdated(update, payloadAttributes)
}

func (e *EngineService) ForkchoiceUpdatedV3(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	return e.forkchoiceUpdated(update, payloadAttributes)
}

func (e *EngineService) ForkchoiceUpdatedV4(update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	return e.forkchoiceUpdated(update, payloadAttributes)
}
