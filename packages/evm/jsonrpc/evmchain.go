// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/event"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	vmerrors "github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// this constant is copied from go-ethereum
const maxLogsTopics = 4

// EVMChain provides common functionality to interact with the EVM state.
type EVMChain struct {
	backend  ChainBackend
	chainID  uint16 // cache
	newBlock *event.Event1[*NewBlockEvent]
	log      log.Logger
	index    *Index // only indexes blocks that will be pruned from the active state
}

type NewBlockEvent struct {
	block *types.Block
	logs  []*types.Log
}

type LogsLimits struct {
	MaxBlocksInLogsFilterRange int
	MaxLogsInResult            int
}

func NewEVMChain(
	backend ChainBackend,
	pub *publisher.Publisher,
	isArchiveNode bool,
	indexDBEngine hivedb.Engine,
	indexDBPath string,
	log log.Logger,
) *EVMChain {
	e := &EVMChain{
		backend:  backend,
		newBlock: event.New1[*NewBlockEvent](),
		log:      log,
		index: NewIndex(
			backend.ISCStateByTrieRoot,
			indexDBEngine,
			path.Join(indexDBPath, backend.ISCChainID().String()),
		),
	}

	blocksFromPublisher := pipe.NewInfinitePipe[*publisher.BlockWithTrieRoot]()

	pub.Events.NewBlock.Hook(func(ev *publisher.ISCEvent[*publisher.BlockWithTrieRoot]) {
		if !ev.ChainID.Equals(*e.backend.ISCChainID()) {
			return
		}
		blocksFromPublisher.In() <- ev.Payload
	})

	// publish blocks on a separate goroutine so that we don't block the publisher
	go func() {
		for ev := range blocksFromPublisher.Out() {
			blockIndex := ev.BlockInfo.BlockIndex
			e.publishNewBlock(blockIndex, ev.TrieRoot)
			if isArchiveNode {
				err := e.index.IndexBlock(ev.TrieRoot)
				if err != nil {
					log.LogErrorf("EVMChain.index.IndexBlock() (index %d): %v", blockIndex, err)
				}
			}
		}
	}()

	return e
}

func (e *EVMChain) publishNewBlock(blockIndex uint32, trieRoot trie.Hash) {
	state, err := e.backend.ISCStateByTrieRoot(trieRoot)
	if err != nil {
		e.log.LogErrorf("EVMChain.publishNewBlock(blockIndex=%v): ISCStateByTrieRoot returned error: %v", blockIndex, err)
		return
	}
	blockNumber := evmBlockNumberByISCBlockIndex(blockIndex)
	db := blockchainDB(state)
	block := db.GetBlockByNumber(blockNumber)
	if block == nil {
		e.log.LogErrorf("EVMChain.publishNewBlock(blockIndex=%v) GetBlockByNumber: block not found", blockIndex)
		return
	}
	var logs []*types.Log
	for _, receipt := range db.GetReceiptsByBlockNumber(blockNumber) {
		logs = append(logs, receipt.Logs...)
	}
	e.newBlock.Trigger(&NewBlockEvent{
		block: block,
		logs:  logs,
	})
}

func (e *EVMChain) Signer() (types.Signer, error) {
	chainID := e.ChainID()
	return evmutil.Signer(big.NewInt(int64(chainID))), nil
}

func (e *EVMChain) ChainID() uint16 {
	if e.chainID == 0 {
		db := blockchainDB(lo.Must(e.backend.ISCLatestState()))
		e.chainID = db.GetChainID()
	}
	return e.chainID
}

func (e *EVMChain) ViewCaller(chainState state.State) vmerrors.ViewCaller {
	e.log.LogDebugf("ViewCaller(chainState=%v)", chainState)
	return func(msg isc.Message) (isc.CallArguments, error) {
		return e.backend.ISCCallView(chainState, msg)
	}
}

func (e *EVMChain) BlockNumber() *big.Int {
	e.log.LogDebugf("BlockNumber()")
	db := blockchainDB(lo.Must(e.backend.ISCLatestState()))
	return big.NewInt(0).SetUint64(db.GetNumber())
}

func (e *EVMChain) GasFeePolicy() *gas.FeePolicy {
	govState := governance.NewStateReaderFromChainState(lo.Must(e.backend.ISCLatestState()))
	return govState.GetGasFeePolicy()
}

func (e *EVMChain) gasLimits() *gas.Limits {
	govState := governance.NewStateReaderFromChainState(lo.Must(e.backend.ISCLatestState()))
	return govState.GetGasLimits()
}

func (e *EVMChain) SendTransaction(tx *types.Transaction) error {
	e.log.LogDebugf("SendTransaction(tx=%v)", tx)
	chainID := e.ChainID()
	if tx.Protected() && tx.ChainId().Uint64() != uint64(chainID) {
		return errors.New("chain ID mismatch")
	}
	signer, err := e.Signer()
	if err != nil {
		return err
	}
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	expectedNonce, err := e.TransactionCount(sender, nil)
	if err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}
	if tx.Nonce() < expectedNonce {
		return fmt.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), expectedNonce)
	}

	gasFeePolicy := e.GasFeePolicy()
	if err := evmutil.CheckGasPrice(tx.GasPrice(), gasFeePolicy); err != nil {
		return err
	}
	if err := e.checkEnoughL2FundsForGasBudget(sender, tx, gasFeePolicy); err != nil {
		return err
	}
	return e.backend.EVMSendTransaction(tx)
}

func (e *EVMChain) checkEnoughL2FundsForGasBudget(sender common.Address, tx *types.Transaction, gasFeePolicy *gas.FeePolicy) error {
	balanceFullDecimals, err := e.Balance(sender, nil)
	if err != nil {
		return fmt.Errorf("could not fetch sender balance: %w", err)
	}

	gasLimits := e.gasLimits()
	iscGasBudgetAffordable := min(
		gasFeePolicy.GasBudgetFromTokensFullDecimals(balanceFullDecimals, tx.GasPrice()),
		gasLimits.MaxGasPerRequest,
	)
	iscGasBudgetTx := min(
		gas.EVMGasToISC(tx.Gas(), &gasFeePolicy.EVMGasRatio),
		gasLimits.MaxGasPerRequest,
	)

	if iscGasBudgetAffordable < iscGasBudgetTx {
		return fmt.Errorf(
			"sender doesn't have enough L2 funds to cover tx gas budget. Balance: %d (can afford %d ISC gas units), expected: %d (%d ISC gas units)",
			balanceFullDecimals,
			iscGasBudgetAffordable,
			util.BaseTokensDecimalsToEthereumDecimals(gasFeePolicy.FeeFromGas(iscGasBudgetTx, tx.GasPrice(), parameters.BaseTokenDecimals), parameters.BaseTokenDecimals),
			iscGasBudgetTx,
		)
	}
	return nil
}

func (e *EVMChain) iscStateFromEVMBlockNumber(blockNumber *big.Int) (state.State, error) {
	if blockNumber == nil {
		return e.backend.ISCLatestState()
	}
	iscBlockIndex, err := iscBlockIndexByEVMBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	cachedTrieHash := e.index.BlockTrieRootByIndex(iscBlockIndex)
	if cachedTrieHash != nil {
		return e.backend.ISCStateByTrieRoot(*cachedTrieHash)
	}
	return e.backend.ISCStateByBlockIndex(iscBlockIndex)
}

func (e *EVMChain) iscStateFromEVMBlockNumberOrHash(blockNumberOrHash *rpc.BlockNumberOrHash) (state.State, error) {
	if blockNumberOrHash == nil {
		return e.backend.ISCLatestState()
	}
	if blockNumber, ok := blockNumberOrHash.Number(); ok {
		return e.iscStateFromEVMBlockNumber(parseBlockNumber(blockNumber))
	}
	blockHash, _ := blockNumberOrHash.Hash()
	block := e.BlockByHash(blockHash)
	if block == nil {
		return nil, fmt.Errorf("block with hash %s not found", blockHash)
	}
	return e.iscStateFromEVMBlockNumber(block.Number())
}

// Returns the anchor corresponding to the given block (i.e. such that
// anchor.StateIndex == blockNumber; running the VM from this anchor will
// produce block n+1).
//
// [anchor n-1] -> VM -> [state n] -> [anchor n] -> VM -> [state n+1] -> [anchor n+1] ...
func (e *EVMChain) iscAnchorFromEVMBlockNumberOrHash(blockNumberOrHash *rpc.BlockNumberOrHash) (*isc.StateAnchor, error) {
	latest, err := e.backend.ISCLatestAnchor()
	if err != nil {
		return nil, fmt.Errorf("retrieving latest anchor: %w", err)
	}
	if blockNumberOrHash == nil {
		return latest, nil
	}
	var stateIndex uint32
	if blockNumber, ok := blockNumberOrHash.Number(); ok {
		bn := parseBlockNumber(blockNumber)
		if bn == nil {
			return e.backend.ISCLatestAnchor()
		}
		stateIndex = blockNumberToStateIndex(bn)
	} else {
		blockHash, _ := blockNumberOrHash.Hash()
		block := e.BlockByHash(blockHash)
		stateIndex = blockNumberToStateIndex(block.Number())
	}
	if stateIndex == latest.GetStateIndex() {
		return latest, nil
	}
	if stateIndex > latest.GetStateIndex() {
		return nil, fmt.Errorf("block %d not found", stateIndex)
	}
	return e.previousAnchor(stateIndex + 1)
}

// Returns the anchor, which was used to form state of given index.
func (e *EVMChain) previousAnchor(stateIndex uint32) (*isc.StateAnchor, error) {
	state, err := e.backend.ISCLatestState()
	if err != nil {
		return nil, fmt.Errorf("retrieving latest state: %w", err)
	}
	blocklogState := blocklog.NewStateReader(blocklog.Contract.StateSubrealmR(state))
	block, found := blocklogState.GetBlockInfo(stateIndex)
	if !found {
		return nil, fmt.Errorf("block %d not found", stateIndex)
	}
	return block.PreviousAnchor, nil
}

func blockNumberToStateIndex(blockNumber *big.Int) uint32 {
	n := blockNumber.Uint64()
	if n > math.MaxUint32 {
		panic("block number too large")
	}
	return uint32(n)
}

func (e *EVMChain) accountsState(chainState state.State) *accounts.StateReader {
	return accounts.NewStateReaderFromChainState(
		chainState.SchemaVersion(),
		chainState,
	)
}

func (e *EVMChain) Balance(address common.Address, blockNumberOrHash *rpc.BlockNumberOrHash) (*big.Int, error) {
	e.log.LogDebugf("Balance(address=%v, blockNumberOrHash=%v)", address, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	return e.accountsState(chainState).GetBaseTokensBalanceFullDecimals(
		isc.NewEthereumAddressAgentID(address),
	), nil
}

func (e *EVMChain) Code(address common.Address, blockNumberOrHash *rpc.BlockNumberOrHash) ([]byte, error) {
	e.log.LogDebugf("Code(address=%v, blockNumberOrHash=%v)", address, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	return emulator.GetCode(stateDBSubrealmR(chainState), address), nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	e.log.LogDebugf("BlockByNumber(blockNumber=%v)", blockNumber)

	cachedBlock := e.index.BlockByNumber(blockNumber)
	if cachedBlock != nil {
		return cachedBlock, nil
	}

	latestState, err := e.backend.ISCLatestState()
	if err != nil {
		return nil, err
	}

	block, err := e.blockByNumber(latestState, blockNumber)
	if err == nil && block == nil {
		return nil, fmt.Errorf("not found")
	}
	return block, err
}

func (e *EVMChain) blockByNumber(chainState state.State, blockNumber *big.Int) (*types.Block, error) {
	db := blockchainDB(chainState)
	bn, err := blockNumberU64(db, blockNumber)
	if err != nil {
		return nil, err
	}
	return db.GetBlockByNumber(bn), nil
}

func blockNumberU64(db *emulator.BlockchainDB, blockNumber *big.Int) (uint64, error) {
	if blockNumber == nil {
		return db.GetNumber(), nil
	}
	if !blockNumber.IsUint64() {
		return 0, fmt.Errorf("block number is too large: %s", blockNumber)
	}
	return blockNumber.Uint64(), nil
}

func (e *EVMChain) TransactionByHash(hash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber, txIndex uint64, err error) {
	e.log.LogDebugf("TransactionByHash(hash=%v)", hash)
	cachedTx, blockHash, blockNumber, txIndex := e.index.TxByHash(hash)
	if cachedTx != nil {
		return cachedTx, blockHash, blockNumber, txIndex, nil
	}
	latestState, err := e.backend.ISCLatestState()
	if err != nil {
		return nil, common.Hash{}, 0, 0, err
	}
	db := blockchainDB(latestState)
	return db.GetTransactionByHash(hash)
}

func (e *EVMChain) TransactionByBlockHashAndIndex(hash common.Hash, index uint64) (tx *types.Transaction, blockNumber uint64, err error) {
	e.log.LogDebugf("TransactionByBlockHashAndIndex(hash=%v, index=%v)", hash, index)
	cachedTx, bn := e.index.TxByBlockHashAndIndex(hash, index)
	if cachedTx != nil {
		return cachedTx, bn, nil
	}
	latestState, err := e.backend.ISCLatestState()
	if err != nil {
		return nil, 0, err
	}
	db := blockchainDB(latestState)
	block := db.GetBlockByHash(hash)
	if block == nil {
		return nil, 0, err
	}
	txs := block.Transactions()
	return txs[index], block.Number().Uint64(), nil
}

func (e *EVMChain) TransactionByBlockNumberAndIndex(blockNumber *big.Int, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumberRet uint64, err error) {
	e.log.LogDebugf("TransactionByBlockNumberAndIndex(blockNumber=%v, index=%v)", blockNumber, index)
	cachedTx, blockHash := e.index.TxByBlockNumberAndIndex(blockNumber, index)
	if cachedTx != nil {
		return cachedTx, blockHash, blockNumber.Uint64(), nil
	}
	latestState, err := e.backend.ISCLatestState()
	if err != nil {
		return nil, common.Hash{}, 0, err
	}
	db := blockchainDB(latestState)
	bn, err := blockNumberU64(db, blockNumber)
	if err != nil {
		return nil, common.Hash{}, 0, err
	}
	block := db.GetBlockByNumber(bn)
	if block == nil {
		return nil, common.Hash{}, 0, err
	}
	txs := block.Transactions()
	if index >= uint64(len(txs)) {
		return nil, common.Hash{}, 0, fmt.Errorf("given index exceeds the amount of txs in block")
	}
	return txs[index], block.Hash(), bn, nil
}

func (e *EVMChain) txsByBlockNumber(blockNumber *big.Int) (txs types.Transactions, err error) {
	e.log.LogDebugf("TxsByBlockNumber(blockNumber=%v, index=%v)", blockNumber)
	cachedTxs := e.index.TxsByBlockNumber(blockNumber)
	if cachedTxs != nil {
		return cachedTxs, nil
	}
	latestState, err := e.backend.ISCLatestState()
	if err != nil {
		return nil, err
	}
	db := blockchainDB(latestState)
	block := db.GetBlockByNumber(blockNumber.Uint64())
	if block == nil {
		return nil, err
	}

	return block.Transactions(), nil
}

func (e *EVMChain) BlockByHash(hash common.Hash) *types.Block {
	e.log.LogDebugf("BlockByHash(hash=%v)", hash)

	cachedBlock := e.index.BlockByHash(hash)
	if cachedBlock != nil {
		return cachedBlock
	}

	db := blockchainDB(lo.Must(e.backend.ISCLatestState()))
	block := db.GetBlockByHash(hash)
	return block
}

func (e *EVMChain) TransactionReceipt(txHash common.Hash) *types.Receipt {
	e.log.LogDebugf("TransactionReceipt(txHash=%v)", txHash)
	rec := e.index.GetReceiptByTxHash(txHash)
	if rec != nil {
		return rec
	}
	db := blockchainDB(lo.Must(e.backend.ISCLatestState()))
	return db.GetReceiptByTxHash(txHash)
}

func (e *EVMChain) TransactionCount(address common.Address, blockNumberOrHash *rpc.BlockNumberOrHash) (uint64, error) {
	e.log.LogDebugf("TransactionCount(address=%v, blockNumberOrHash=%v)", address, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return 0, err
	}
	return emulator.GetNonce(stateDBSubrealmR(chainState), address), nil
}

func (e *EVMChain) CallContract(callMsg ethereum.CallMsg, blockNumberOrHash *rpc.BlockNumberOrHash) ([]byte, error) {
	anchor, err := e.iscAnchorFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	blockinfo, err := e.getBlockInfoByAnchor(anchor)
	if err != nil {
		return nil, err
	}
	return e.backend.EVMCall(anchor, callMsg, blockinfo.L1Params)
}

func (e *EVMChain) EstimateGas(callMsg ethereum.CallMsg, blockNumberOrHash *rpc.BlockNumberOrHash) (uint64, error) {
	anchor, err := e.iscAnchorFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return 0, err
	}
	blockinfo, err := e.getBlockInfoByAnchor(anchor)
	if err != nil {
		return 0, err
	}
	return e.backend.EVMEstimateGas(anchor, callMsg, blockinfo.L1Params)
}

func (e *EVMChain) GasPrice() *big.Int {
	e.log.LogDebugf("GasPrice()")
	return e.GasFeePolicy().DefaultGasPriceFullDecimals(parameters.BaseTokenDecimals)
}

func (e *EVMChain) StorageAt(address common.Address, key common.Hash, blockNumberOrHash *rpc.BlockNumberOrHash) (common.Hash, error) {
	e.log.LogDebugf("StorageAt(address=%v, key=%v, blockNumberOrHash=%v)", address, key, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return common.Hash{}, err
	}
	return emulator.GetState(stateDBSubrealmR(chainState), address, key), nil
}

func (e *EVMChain) BlockTransactionCountByHash(blockHash common.Hash) uint64 {
	e.log.LogDebugf("BlockTransactionCountByHash(blockHash=%v)", blockHash)
	block := e.BlockByHash(blockHash)
	if block == nil {
		return 0
	}
	return uint64(len(block.Transactions()))
}

func (e *EVMChain) BlockTransactionCountByNumber(blockNumber *big.Int) (uint64, error) {
	e.log.LogDebugf("BlockTransactionCountByNumber(blockNumber=%v)", blockNumber)
	block, err := e.BlockByNumber(blockNumber)
	if err != nil {
		return 0, err
	}
	return uint64(len(block.Transactions())), nil
}

// Logs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
func (e *EVMChain) Logs(query *ethereum.FilterQuery, params *LogsLimits) ([]*types.Log, error) {
	e.log.LogDebugf("Logs(q=%v)", query)
	if query == nil {
		query = &ethereum.FilterQuery{}
	}

	// this condition is copied from go-ethereum
	if len(query.Topics) > maxLogsTopics {
		return nil, errors.New("too many topics in filter")
	}

	if params == nil {
		params = &LogsLimits{}
	}

	logs := make([]*types.Log, 0)

	// single block query
	if query.BlockHash != nil {
		state, err := e.iscStateFromEVMBlockNumberOrHash(&rpc.BlockNumberOrHash{
			BlockHash: query.BlockHash,
		})
		if err != nil {
			return nil, err
		}
		db := blockchainDB(state)
		receipts := db.GetReceiptsByBlockNumber(uint64(state.BlockIndex()))
		err = filterAndAppendToLogs(query, receipts, &logs, params.MaxLogsInResult)
		if err != nil {
			return nil, err
		}
		return logs, nil
	}

	// block range query

	// Initialize unset filter boundaries to run from genesis to chain head
	first := big.NewInt(1) // skip genesis since it has no logs
	last := new(big.Int).SetUint64(uint64(lo.Must(e.backend.ISCLatestState()).BlockIndex()))
	from := first
	if query.FromBlock != nil && query.FromBlock.Cmp(first) >= 0 && query.FromBlock.Cmp(last) <= 0 {
		from = query.FromBlock
	}
	to := last
	if query.ToBlock != nil && query.ToBlock.Cmp(first) >= 0 && query.ToBlock.Cmp(last) <= 0 {
		to = query.ToBlock
	}

	if !from.IsUint64() || !to.IsUint64() {
		return nil, errors.New("block number is too large")
	}
	{
		from := from.Uint64()
		to := to.Uint64()
		if to > from && to-from > uint64(params.MaxBlocksInLogsFilterRange) {
			return nil, errors.New("ServerError(-32000) too many blocks in filter range") // ServerError(-32000) part is necessary because subgraph expects that string in the error msg: https://github.com/graphprotocol/graph-node/blob/591ad93b5144ff5e6037b73862c607effad90e7f/chain/ethereum/src/ethereum_adapter.rs#L335
		}
		for i := from; i <= to; i++ {
			state, err := e.iscStateFromEVMBlockNumber(new(big.Int).SetUint64(i))
			if err != nil {
				return nil, err
			}
			err = filterAndAppendToLogs(
				query,
				blockchainDB(state).GetReceiptsByBlockNumber(i),
				&logs,
				params.MaxLogsInResult,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	return logs, nil
}

func filterAndAppendToLogs(query *ethereum.FilterQuery, receipts []*types.Receipt, logs *[]*types.Log, maxLogsInResult int) error {
	for _, r := range receipts {
		if r.Status == types.ReceiptStatusFailed {
			continue
		}
		if !evmtypes.BloomFilter(r.Bloom, query.Addresses, query.Topics) {
			continue
		}
		for _, log := range r.Logs {
			if !evmtypes.LogMatches(log, query.Addresses, query.Topics) {
				continue
			}
			if len(*logs) >= maxLogsInResult {
				return errors.New("too many logs in result")
			}
			*logs = append(*logs, log)
		}
	}
	return nil
}

func (e *EVMChain) SubscribeNewHeads(ch chan<- *types.Header) (unsubscribe func()) {
	e.log.LogDebugf("SubscribeNewHeads(ch=?)")
	return e.newBlock.Hook(func(ev *NewBlockEvent) {
		ch <- ev.block.Header()
	}).Unhook
}

func (e *EVMChain) SubscribeLogs(q *ethereum.FilterQuery, ch chan<- []*types.Log) (unsubscribe func()) {
	e.log.LogDebugf("SubscribeLogs(q=%v, ch=?)", q)
	return e.newBlock.Hook(func(ev *NewBlockEvent) {
		if q.BlockHash != nil && *q.BlockHash != ev.block.Hash() {
			return
		}
		if q.FromBlock != nil && q.FromBlock.IsUint64() && q.FromBlock.Cmp(ev.block.Number()) > 0 {
			return
		}
		if q.ToBlock != nil && q.ToBlock.IsUint64() && q.ToBlock.Cmp(ev.block.Number()) < 0 {
			return
		}

		var matchedLogs []*types.Log
		for _, log := range ev.logs {
			if evmtypes.LogMatches(log, q.Addresses, q.Topics) {
				matchedLogs = append(matchedLogs, log)
			}
		}
		if len(matchedLogs) > 0 {
			ch <- matchedLogs
		}
	}).Unhook
}

func (e *EVMChain) iscRequestsInBlock(evmBlockNumber uint64) (*blocklog.BlockInfo, []isc.Request, error) {
	iscState, err := e.iscStateFromEVMBlockNumber(new(big.Int).SetUint64(evmBlockNumber))
	if err != nil {
		return nil, nil, err
	}
	return blocklog.NewStateReaderFromChainState(iscState).GetRequestsInBlock(iscState.BlockIndex())
}

// traceTransaction allows the tracing of a single EVM transaction.
// "Fake" transactions that are emitted e.g. for L1 deposits return some mocked trace.
func (e *EVMChain) traceTransaction(
	config *tracers.TraceConfig,
	blockInfo *blocklog.BlockInfo,
	requestsInBlock []isc.Request,
	tx *types.Transaction,
	txIndex uint64,
	blockHash common.Hash,
) (json.RawMessage, error) {
	tracerType := "callTracer"
	if config.Tracer != nil {
		tracerType = *config.Tracer
	}

	tracer, err := newTracer(
		tracerType,
		&tracers.Context{
			BlockHash:   blockHash,
			BlockNumber: new(big.Int).SetUint64(uint64(blockInfo.BlockIndex)),
			TxIndex:     int(txIndex),
			TxHash:      tx.Hash(),
		},
		config.TracerConfig,
	)
	if err != nil {
		return nil, err
	}

	err = e.backend.EVMTrace(
		blockInfo.PreviousAnchor,
		blockInfo.Timestamp,
		requestsInBlock,
		tracer,
		blockInfo.L1Params,
	)
	if err != nil {
		return nil, err
	}

	res, err := tracer.GetResult()
	if err != nil {
		return nil, err
	}

	var txResults []TxTraceResult
	err = json.Unmarshal(res, &txResults)
	if err != nil {
		return nil, err
	}

	if len(txResults) <= int(txIndex) {
		return nil, errors.New("tx trace not found in tracer result")
	}
	txTrace := txResults[int(txIndex)]
	if txTrace.Error != "" {
		return nil, errors.New(txTrace.Error)
	}
	return txTrace.Result, nil
}

func (e *EVMChain) debugTraceBlock(config *tracers.TraceConfig, block *types.Block) (any, error) {
	iscBlock, iscRequestsInBlock, err := e.iscRequestsInBlock(block.NumberU64())
	if err != nil {
		return nil, err
	}

	tracerType := "callTracer"
	if config.Tracer != nil {
		tracerType = *config.Tracer
	}

	tracer, err := newTracer(
		tracerType,
		&tracers.Context{
			BlockHash:   block.Hash(),
			BlockNumber: new(big.Int).SetUint64(uint64(iscBlock.BlockIndex)),
		},
		config.TracerConfig,
	)
	if err != nil {
		return nil, err
	}

	err = e.backend.EVMTrace(
		iscBlock.PreviousAnchor,
		iscBlock.Timestamp,
		iscRequestsInBlock,
		tracer,
		iscBlock.L1Params,
	)
	if err != nil {
		return nil, err
	}
	return tracer.GetResult()
}

func (e *EVMChain) TraceTransaction(txHash common.Hash, config *tracers.TraceConfig) (any, error) {
	e.log.LogDebugf("TraceTransaction(txHash=%v, config=?)", txHash)

	tx, blockHash, blockNumber, txIndex, err := e.TransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	if blockNumber == 0 {
		return nil, errors.New("transaction not found")
	}

	iscBlock, iscRequestsInBlock, err := e.iscRequestsInBlock(blockNumber)
	if err != nil {
		return nil, err
	}

	return e.traceTransaction(
		config,
		iscBlock,
		iscRequestsInBlock,
		tx,
		txIndex,
		blockHash,
	)
}

func (e *EVMChain) TraceBlockByHash(blockHash common.Hash, config *tracers.TraceConfig) (any, error) {
	e.log.LogDebugf("TraceBlockByHash(blockHash=%v, config=?)", blockHash)

	block := e.BlockByHash(blockHash)
	if block == nil {
		return nil, fmt.Errorf("block not found: %s", blockHash.String())
	}

	return e.debugTraceBlock(config, block)
}

func (e *EVMChain) TraceBlockByNumber(blockNumber uint64, config *tracers.TraceConfig) (any, error) {
	e.log.LogDebugf("TraceBlockByNumber(blockNumber=%v, config=?)", blockNumber)

	block, err := e.BlockByNumber(big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("block not found: %d", blockNumber)
	}

	return e.debugTraceBlock(config, block)
}

func (e *EVMChain) getBlockByNumberOrHash(blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if h, ok := blockNrOrHash.Hash(); ok {
		block := e.BlockByHash(h)
		if block == nil {
			return nil, fmt.Errorf("block not found: %v", blockNrOrHash.String())
		}
		return block, nil
	} else if n, ok := blockNrOrHash.Number(); ok {
		switch n {
		case rpc.LatestBlockNumber:
			return e.BlockByNumber(nil)
		default:
			if n < 0 {
				return nil, fmt.Errorf("%v is unsupported", blockNrOrHash.String())
			}

			return e.BlockByNumber(big.NewInt(n.Int64()))
		}
	}

	return nil, fmt.Errorf("block not found: %v", blockNrOrHash.String())
}

func (e *EVMChain) GetRawBlock(blockNrOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	block, err := e.getBlockByNumberOrHash(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	return rlp.EncodeToBytes(block)
}

func (e *EVMChain) GetBlockReceipts(blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, []*types.Transaction, error) {
	e.log.LogDebugf("GetBlockReceipts(blockNumber=%v)", blockNrOrHash.String())

	block, err := e.getBlockByNumberOrHash(blockNrOrHash)
	if err != nil {
		return nil, nil, err
	}

	chainState, err := e.iscStateFromEVMBlockNumber(block.Number())
	if err != nil {
		return nil, nil, err
	}

	db := blockchainDB(chainState)

	return db.GetReceiptsByBlockNumber(block.NumberU64()), db.GetTransactionsByBlockNumber(block.NumberU64()), nil
}

func (e *EVMChain) TraceBlock(bn rpc.BlockNumber) (any, error) {
	e.log.LogDebugf("TraceBlock(blockNumber=%v)", bn)

	block, err := e.getBlockByNumberOrHash(rpc.BlockNumberOrHashWithNumber(bn))
	if err != nil {
		return nil, err
	}
	iscBlock, iscRequestsInBlock, err := e.iscRequestsInBlock(block.NumberU64())
	if err != nil {
		return nil, err
	}

	blockTxs, err := e.txsByBlockNumber(new(big.Int).SetUint64(block.NumberU64()))
	if err != nil {
		return nil, err
	}

	results := TraceBlock{
		Jsonrpc: "2.0",
		Result:  make([]*Trace, 0),
		ID:      1,
	}

	for i, tx := range blockTxs {
		debugResultJSON, err := e.traceTransaction(
			&tracers.TraceConfig{},
			iscBlock,
			iscRequestsInBlock,
			tx,
			uint64(i),
			block.Hash(),
		)

		if err == nil {
			var debugResult CallFrame
			err = json.Unmarshal(debugResultJSON, &debugResult)
			if err != nil {
				return nil, err
			}

			blockHash := block.Hash()
			txHash := tx.Hash()
			results.Result = append(results.Result, convertToTrace(debugResult, &blockHash, block.NumberU64(), &txHash, uint64(i))...)
		}
	}

	return results, nil
}

func (e *EVMChain) getBlockInfoByAnchor(anchor *isc.StateAnchor) (*blocklog.BlockInfo, error) {
	stateMetadata, err := transaction.StateMetadataFromBytes(anchor.GetStateMetadata())
	if err != nil {
		return nil, err
	}
	state, err := e.backend.ISCStateByTrieRoot(stateMetadata.L1Commitment.TrieRoot())
	if err != nil {
		return nil, err
	}
	blockInfo, ok := blocklog.NewStateReaderFromChainState(state).GetBlockInfo(anchor.GetStateIndex())
	if !ok {
		return nil, fmt.Errorf("blockinfo not found")
	}
	return blockInfo, nil
}

var maxUint32 = big.NewInt(math.MaxUint32)

// the first EVM block (number 0) is "minted" at ISC block index 0 (init chain)
func iscBlockIndexByEVMBlockNumber(blockNumber *big.Int) (uint32, error) {
	if blockNumber.Cmp(maxUint32) > 0 {
		return 0, fmt.Errorf("block number is too large: %s", blockNumber)
	}
	return uint32(blockNumber.Uint64()), nil
}

// the first EVM block (number 0) is "minted" at ISC block index 1 (init chain)
func evmBlockNumberByISCBlockIndex(n uint32) uint64 {
	return uint64(n)
}

func blockchainDB(chainState state.State) *emulator.BlockchainDB {
	govState := governance.NewStateReaderFromChainState(chainState)
	gasLimits := govState.GetGasLimits()
	gasFeePolicy := govState.GetGasFeePolicy()
	blockKeepAmount := govState.GetBlockKeepAmount()
	return emulator.NewBlockchainDB(
		buffered.NewBufferedKVStore(evm.EmulatorStateSubrealmR(evm.ContractPartitionR(chainState))),
		gas.EVMBlockGasLimit(gasLimits, &gasFeePolicy.EVMGasRatio),
		blockKeepAmount,
	)
}

func stateDBSubrealmR(chainState state.State) kv.KVStoreReader {
	return emulator.StateDBSubrealmR(evm.EmulatorStateSubrealmR(evm.ContractPartitionR(chainState)))
}
