// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package emulator provides tools to emulate Ethereum chains and contracts.
//
// Code adapted from go-ethereum/accounts/abi/bind/backends/simulated.go
package emulator

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"golang.org/x/xerrors"
)

var (
	ErrBlockDoesNotExist       = errors.New("block does not exist in blockchain")
	ErrTransactionDoesNotExist = errors.New("transaction does not exist")
)

type pending struct {
	header   *types.Header
	state    *state.StateDB
	txs      []*types.Transaction
	receipts []*types.Receipt
	gasPool  *core.GasPool
}

type EVMEmulator struct {
	database   ethdb.Database
	blockchain *core.BlockChain
	pending    *pending
	engine     consensus.Engine
}

var (
	TxGas     = uint64(21000) // gas cost of simple transfer (not contract creation / call)
	vmConfig  = vm.Config{}
	timeDelta = uint64(1) // amount of seconds to add to timestamp by default for new blocks
)

func MakeConfig(chainID int) *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:             big.NewInt(int64(chainID)),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP150Hash:          common.Hash{},
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		Ethash:              &params.EthashConfig{},
	}
}

func Signer(chainID *big.Int) types.Signer {
	return types.NewEIP155Signer(chainID)
}

func InitGenesis(chainID int, db ethdb.Database, alloc core.GenesisAlloc, gasLimit, timestamp uint64) {
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored != common.Hash{}) {
		panic("genesis block already initialized")
	}
	genesis := core.Genesis{
		Config:    MakeConfig(chainID),
		Alloc:     alloc,
		GasLimit:  gasLimit,
		Timestamp: timestamp,
	}
	genesis.MustCommit(db)
}

// disable caching & shnapshotting, since it produces a nondeterministic result
var cacheConfig = &core.CacheConfig{}

func NewEVMEmulator(db ethdb.Database, timestamp ...uint64) *EVMEmulator {
	canonicalHash := rawdb.ReadCanonicalHash(db, 0)
	if (canonicalHash == common.Hash{}) {
		panic("must initialize genesis block first")
	}

	config := rawdb.ReadChainConfig(db, canonicalHash)
	engine := ethash.NewFaker()
	blockchain, _ := core.NewBlockChain(db, cacheConfig, config, engine, vmConfig, nil, nil)

	e := &EVMEmulator{
		database:   db,
		blockchain: blockchain,
		engine:     engine,
	}

	parentTime := e.blockchain.CurrentBlock().Header().Time
	// ensure that timestamp is larger, even when generating many blocks per second in tests
	ts := parentTime + timeDelta
	if len(timestamp) > 0 && timestamp[0] > parentTime {
		ts = timestamp[0]
	}

	e.Rollback(ts)
	return e
}

// Close terminates the underlying blockchain's update loop.
func (e *EVMEmulator) Close() error {
	e.blockchain.Stop()
	return nil
}

// Commit imports all the pending transactions as a single block and starts a
// fresh new state.
func (e *EVMEmulator) Commit() {
	if len(e.pending.txs) == 0 {
		return
	}
	if _, err := e.blockchain.InsertChain([]*types.Block{e.finalizeBlock()}); err != nil {
		panic(err)
	}
	e.Rollback(e.pending.header.Time + timeDelta)
}

func (e *EVMEmulator) finalizeBlock() *types.Block {
	block, err := e.engine.FinalizeAndAssemble(
		e.blockchain,
		e.pending.header,
		e.pending.state,
		e.pending.txs,
		nil,
		e.pending.receipts,
	)
	if err != nil {
		panic(err)
	}
	root, err := e.pending.state.Commit(true)
	if err != nil {
		panic(err)
	}
	if err := e.pending.state.Database().TrieDB().Commit(root, false, nil); err != nil {
		panic(err)
	}
	return block
}

// Rollback aborts all pending transactions, reverting to the last committed state.
func (e *EVMEmulator) Rollback(timestamp uint64) {
	statedb, err := e.blockchain.State()
	if err != nil {
		panic(err)
	}

	parent := e.blockchain.CurrentBlock()
	header := &types.Header{
		Root:       statedb.IntermediateRoot(true),
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		Difficulty: e.engine.CalcDifficulty(e.blockchain, timestamp, parent.Header()),
		Number:     new(big.Int).Add(parent.Number(), common.Big1),
		GasLimit:   parent.GasLimit(),
		Time:       timestamp,
	}

	e.pending = &pending{
		header:   header,
		state:    statedb,
		txs:      []*types.Transaction{},
		receipts: []*types.Receipt{},
		gasPool:  new(core.GasPool).AddGas(header.GasLimit),
	}
}

// stateByBlockNumber retrieves a state by a given blocknumber.
func (e *EVMEmulator) stateByBlockNumber(blockNumber *big.Int) (*state.StateDB, error) {
	if blockNumber == nil || blockNumber.Cmp(e.blockchain.CurrentBlock().Number()) == 0 {
		return e.blockchain.State()
	}
	block, err := e.blockByNumberNoLock(blockNumber)
	if err != nil {
		return nil, err
	}
	return e.blockchain.StateAt(block.Root())
}

// CodeAt returns the code associated with a certain account in the blockchain.
func (e *EVMEmulator) CodeAt(contract common.Address, blockNumber *big.Int) ([]byte, error) {
	stateDB, err := e.stateByBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	return stateDB.GetCode(contract), nil
}

// BalanceAt returns the wei balance of a certain account in the blockchain.
func (e *EVMEmulator) BalanceAt(contract common.Address, blockNumber *big.Int) (*big.Int, error) {
	stateDB, err := e.stateByBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	return stateDB.GetBalance(contract), nil
}

// NonceAt returns the nonce of a certain account in the blockchain.
func (e *EVMEmulator) NonceAt(contract common.Address, blockNumber *big.Int) (uint64, error) {
	stateDB, err := e.stateByBlockNumber(blockNumber)
	if err != nil {
		return 0, err
	}
	return stateDB.GetNonce(contract), nil
}

// StorageAt returns the value of key in the storage of an account in the blockchain.
func (e *EVMEmulator) StorageAt(contract common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	stateDB, err := e.stateByBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	val := stateDB.GetState(contract, key)
	return val[:], nil
}

// TransactionReceipt returns the receipt of a transaction.
func (e *EVMEmulator) TransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	receipt, _, _, _ := rawdb.ReadReceipt(e.database, txHash, e.blockchain.Config()) //nolint:dogsled
	return receipt, nil
}

func (e *EVMEmulator) TransactionByHash(txHash common.Hash) *types.Transaction {
	tx, _, _, _ := rawdb.ReadTransaction(e.database, txHash) //nolint:dogsled
	return tx
}

// BlockByHash retrieves a block based on the block hash.
func (e *EVMEmulator) BlockByHash(hash common.Hash) *types.Block {
	block := e.blockchain.GetBlockByHash(hash)
	if block != nil {
		return block
	}
	return nil
}

// BlockByNumber retrieves a block from the database by number, caching it
// (associated with its hash) if found.
func (e *EVMEmulator) BlockByNumber(number *big.Int) (*types.Block, error) {
	return e.blockByNumberNoLock(number)
}

// blockByNumberNoLock retrieves a block from the database by number, caching it
// (associated with its hash) if found without Lock.
func (e *EVMEmulator) blockByNumberNoLock(number *big.Int) (*types.Block, error) {
	if number == nil {
		return e.blockchain.CurrentBlock(), nil
	}
	block := e.blockchain.GetBlockByNumber(uint64(number.Int64()))
	if block == nil {
		return nil, ErrBlockDoesNotExist
	}
	return block, nil
}

// HeaderByHash returns a block header from the current canonical chain.
func (e *EVMEmulator) HeaderByHash(hash common.Hash) (*types.Header, error) {
	header := e.blockchain.GetHeaderByHash(hash)
	if header == nil {
		return nil, ErrBlockDoesNotExist
	}
	return header, nil
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (e *EVMEmulator) HeaderByNumber(block *big.Int) (*types.Header, error) {
	if block == nil {
		return e.blockchain.CurrentHeader(), nil
	}
	return e.blockchain.GetHeaderByNumber(uint64(block.Int64())), nil
}

// TransactionCount returns the number of transactions in a given block.
func (e *EVMEmulator) TransactionCount(blockHash common.Hash) (uint, error) {
	block := e.blockchain.GetBlockByHash(blockHash)
	if block == nil {
		return uint(0), ErrBlockDoesNotExist
	}
	return uint(block.Transactions().Len()), nil
}

// TransactionInBlock returns the transaction for a specific block at a specific index.
func (e *EVMEmulator) TransactionInBlock(blockHash common.Hash, index uint) (*types.Transaction, error) {
	block := e.blockchain.GetBlockByHash(blockHash)
	if block == nil {
		return nil, ErrBlockDoesNotExist
	}

	transactions := block.Transactions()
	if uint(len(transactions)) < index+1 {
		return nil, ErrTransactionDoesNotExist
	}

	return transactions[index], nil
}

func newRevertError(result *core.ExecutionResult) *revertError {
	reason, errUnpack := abi.UnpackRevert(result.Revert())
	err := errors.New("execution reverted")
	if errUnpack == nil {
		err = fmt.Errorf("execution reverted: %v", reason)
	}
	return &revertError{
		error:  err,
		reason: hexutil.Encode(result.Revert()),
	}
}

// revertError is an API error that encompasses an EVM revert with JSON error
// code and a binary data blob.
type revertError struct {
	error
	reason string // revert reason hex encoded
}

// ErrorCode returns the JSON error code for a revert.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func (e *revertError) ErrorCode() int {
	return 3
}

// ErrorData returns the hex encoded revert reason.
func (e *revertError) ErrorData() interface{} {
	return e.reason
}

// CallContract executes a contract call.
func (e *EVMEmulator) CallContract(call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	header, err := e.HeaderByNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	stateDB, err := e.blockchain.StateAt(header.Root)
	if err != nil {
		return nil, err
	}
	res, err := e.callContract(call, header, stateDB)
	if err != nil {
		return nil, err
	}
	// If the result contains a revert reason, try to unpack and return it.
	if len(res.Revert()) > 0 {
		return nil, newRevertError(res)
	}
	return res.Return(), res.Err
}

// PendingNonceAt implements PendingStateReader.PendingNonceAt, retrieving
// the nonce currently pending for the account.
func (e *EVMEmulator) PendingNonceAt(account common.Address) (uint64, error) {
	return e.pending.state.GetOrNewStateObject(account).Nonce(), nil
}

// SuggestGasPrice implements ContractTransactor.SuggestGasPrice. Since the simulated
// chain doesn't have miners, we just return a gas price of 1 for any call.
func (e *EVMEmulator) SuggestGasPrice() (*big.Int, error) {
	return evm.GasPrice, nil
}

// EstimateGas executes the requested code against the currently pending block/state and
// returns the used amount of gas.
func (e *EVMEmulator) EstimateGas(call ethereum.CallMsg) (uint64, error) {
	// Determine the lowest and highest possible gas limits to binary search in between
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		max uint64
	)
	if call.Gas >= params.TxGas {
		hi = call.Gas
	} else {
		hi = e.pending.header.GasLimit
	}
	// Recap the highest gas allowance with account's balance.
	if call.GasPrice != nil && call.GasPrice.BitLen() != 0 {
		balance := e.pending.state.GetBalance(call.From) // from can't be nil
		available := new(big.Int).Set(balance)
		if call.Value != nil {
			if call.Value.Cmp(available) >= 0 {
				return 0, errors.New("insufficient funds for transfer")
			}
			available.Sub(available, call.Value)
		}
		allowance := new(big.Int).Div(available, call.GasPrice)
		if allowance.IsUint64() && hi > allowance.Uint64() {
			transfer := call.Value
			if transfer == nil {
				transfer = new(big.Int)
			}
			log.Warn("Gas estimation capped by limited funds", "original", hi, "balance", balance,
				"sent", transfer, "gasprice", call.GasPrice, "fundable", allowance)
			hi = allowance.Uint64()
		}
	}
	max = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *core.ExecutionResult, error) {
		call.Gas = gas

		snapshot := e.pending.state.Snapshot()
		res, err := e.callContract(call, e.pending.header, e.pending.state)
		e.pending.state.RevertToSnapshot(snapshot)

		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, fmt.Errorf("Bail out: %w", err)
		}
		return res.Failed(), res, nil //nolint:gocritic
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		failed, _, err := executable(mid)
		// If the error is not nil(consensus error), it means the provided message
		// call or transaction will never be accepted no matter how much gas it is
		// assigned. Return the error directly, don't struggle any more
		if err != nil {
			return 0, fmt.Errorf("executable(mid): %w", err)
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == max {
		failed, result, err := executable(hi)
		if err != nil {
			return 0, fmt.Errorf("executable(hi): %w", err)
		}
		if failed {
			if result != nil && !errors.Is(result.Err, vm.ErrOutOfGas) {
				if len(result.Revert()) > 0 {
					return 0, newRevertError(result)
				}
				return 0, fmt.Errorf("revert: %w", result.Err)
			}
			// Otherwise, the specified gas cap is too low
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", max)
		}
	}
	return hi, nil
}

// callContract implements common code between normal and pending contract calls.
// state is modified during execution, make sure to copy it if necessary.
func (e *EVMEmulator) callContract(call ethereum.CallMsg, header *types.Header, stateDB *state.StateDB) (*core.ExecutionResult, error) {
	// Ensure message is initialized properly.
	if call.GasPrice == nil {
		call.GasPrice = big.NewInt(1)
	}
	if call.Gas == 0 {
		call.Gas = 50000000
	}
	if call.Value == nil {
		call.Value = new(big.Int)
	}
	// Set infinite balance to the fake caller account.
	from := stateDB.GetOrNewStateObject(call.From)
	from.SetBalance(math.MaxBig256)
	// Execute the call.
	msg := callMsg{call}

	txContext := core.NewEVMTxContext(msg)
	evmContext := core.NewEVMBlockContext(header, e.blockchain, nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmEnv := vm.NewEVM(evmContext, txContext, stateDB, e.blockchain.Config(), vmConfig)
	gasPool := new(core.GasPool).AddGas(math.MaxUint64)

	return core.NewStateTransition(vmEnv, msg, gasPool).TransitionDb()
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// SendTransaction updates the pending block to include the given transaction.
// It returns an error if the transaction is invalid.
func (e *EVMEmulator) SendTransaction(tx *types.Transaction, gasLimit uint64) (*types.Receipt, uint64, error) {
	sender, err := types.Sender(e.Signer(), tx)
	if err != nil {
		return nil, 0, xerrors.Errorf("invalid transaction: %w", err)
	}
	nonce := e.pending.state.GetNonce(sender)
	if tx.Nonce() != nonce {
		return nil, 0, xerrors.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), nonce)
	}

	snap := e.pending.state.Snapshot()

	e.pending.state.Prepare(tx.Hash(), len(e.pending.txs))

	availableGas := minUint64(gasLimit, e.pending.gasPool.Gas())
	gasPool := core.GasPool(availableGas)

	receipt, err := core.ApplyTransaction(
		e.blockchain.Config(),
		e.blockchain,
		nil,
		&gasPool,
		e.pending.state,
		e.pending.header,
		tx,
		&e.pending.header.GasUsed,
		vmConfig,
	)
	gasUsed := availableGas - gasPool.Gas()
	if err != nil {
		e.pending.state.RevertToSnapshot(snap)
		return nil, gasUsed, err
	}

	e.pending.txs = append(e.pending.txs, tx)
	e.pending.receipts = append(e.pending.receipts, receipt)
	e.pending.gasPool.SubGas(receipt.GasUsed)

	return receipt, gasUsed, nil
}

// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
func (e *EVMEmulator) FilterLogs(query *ethereum.FilterQuery) ([]*types.Log, error) {
	var filter *filters.Filter
	if query.BlockHash != nil {
		// Block filter requested, construct a single-shot filter
		filter = filters.NewBlockFilter(&filterBackend{e.database, e.blockchain}, *query.BlockHash, query.Addresses, query.Topics)
	} else {
		// Initialize unset filter boundaries to run from genesis to chain head
		from := int64(0)
		if query.FromBlock != nil {
			from = query.FromBlock.Int64()
		}
		to := int64(-1)
		if query.ToBlock != nil {
			to = query.ToBlock.Int64()
		}
		// Construct the range filter
		filter = filters.NewRangeFilter(&filterBackend{e.database, e.blockchain}, from, to, query.Addresses, query.Topics)
	}
	return filter.Logs(context.Background())
}

// Blockchain returns the underlying blockchain.
func (e *EVMEmulator) Blockchain() *core.BlockChain {
	return e.blockchain
}

func (e *EVMEmulator) Signer() types.Signer {
	return Signer(e.blockchain.Config().ChainID)
}

// callMsg implements core.Message to allow passing it as a transaction simulator.
type callMsg struct {
	ethereum.CallMsg
}

func (m callMsg) From() common.Address         { return m.CallMsg.From }
func (m callMsg) Nonce() uint64                { return 0 }
func (m callMsg) IsFake() bool                 { return true }
func (m callMsg) To() *common.Address          { return m.CallMsg.To }
func (m callMsg) GasPrice() *big.Int           { return m.CallMsg.GasPrice }
func (m callMsg) GasFeeCap() *big.Int          { return m.CallMsg.GasFeeCap }
func (m callMsg) GasTipCap() *big.Int          { return m.CallMsg.GasTipCap }
func (m callMsg) Gas() uint64                  { return m.CallMsg.Gas }
func (m callMsg) Value() *big.Int              { return m.CallMsg.Value }
func (m callMsg) Data() []byte                 { return m.CallMsg.Data }
func (m callMsg) AccessList() types.AccessList { return m.CallMsg.AccessList }

// filterBackend implements filters.Backend.
// TODO: take bloom-bits acceleration structures into account
type filterBackend struct {
	db ethdb.Database
	bc *core.BlockChain
}

var _ filters.Backend = &filterBackend{}

func (fb *filterBackend) ChainDb() ethdb.Database  { return fb.db }
func (fb *filterBackend) EventMux() *event.TypeMux { panic("not supported") } //nolint:staticcheck

func (fb *filterBackend) HeaderByNumber(ctx context.Context, block rpc.BlockNumber) (*types.Header, error) {
	if block == rpc.LatestBlockNumber {
		return fb.bc.CurrentHeader(), nil
	}
	return fb.bc.GetHeaderByNumber(uint64(block.Int64())), nil
}

func (fb *filterBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return fb.bc.GetHeaderByHash(hash), nil
}

func (fb *filterBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	number := rawdb.ReadHeaderNumber(fb.db, hash)
	if number == nil {
		return nil, nil
	}
	return rawdb.ReadReceipts(fb.db, hash, *number, fb.bc.Config()), nil
}

func (fb *filterBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(fb.db, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(fb.db, hash, *number, fb.bc.Config())
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (fb *filterBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return nullSubscription()
}

func (fb *filterBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return fb.bc.SubscribeChainEvent(ch)
}

func (fb *filterBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return fb.bc.SubscribeRemovedLogsEvent(ch)
}

func (fb *filterBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return fb.bc.SubscribeLogsEvent(ch)
}

func (fb *filterBackend) SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return nullSubscription()
}

func (fb *filterBackend) BloomStatus() (uint64, uint64) { return 4096, 0 }

func (fb *filterBackend) ServiceFilter(ctx context.Context, ms *bloombits.MatcherSession) {
	panic("not supported")
}

func nullSubscription() event.Subscription {
	return event.NewSubscription(func(quit <-chan struct{}) error {
		<-quit
		return nil
	})
}
