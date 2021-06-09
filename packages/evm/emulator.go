// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package evm provides tools to emulate Ethereum chains and contracts.
//
// Code adapted from go-ethereum/accounts/abi/bind/backends/simulated.go
package evm

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/xerrors"
)

var (
	errBlockNumberUnsupported  = errors.New("EVMEmulator cannot access blocks other than the latest block")
	ErrBlockDoesNotExist       = errors.New("block does not exist in blockchain")
	ErrTransactionDoesNotExist = errors.New("transaction does not exist")
)

type EVMEmulator struct {
	database   ethdb.Database
	blockchain *core.BlockChain

	pendingBlock *types.Block
	pendingState *state.StateDB
}

var (
	TxGas       = uint64(21000) // gas cost of simple transfer (not contract creation / call)
	MaxGasLimit = uint64(math.MaxUint64 / 2)
	GasPrice    = big.NewInt(0)
)

var Config = params.AllEthashProtocolChanges

func Signer() types.Signer {
	return types.NewEIP155Signer(Config.ChainID)
}

func InitGenesis(db ethdb.Database, alloc core.GenesisAlloc) {
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored != common.Hash{}) {
		panic("genesis block already initialized")
	}
	genesis := core.Genesis{Config: Config, Alloc: alloc, GasLimit: MaxGasLimit}
	genesis.MustCommit(db)
}

func NewEVMEmulator(db ethdb.Database) *EVMEmulator {
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		panic("must initialize genesis block first")
	}

	blockchain, _ := core.NewBlockChain(db, nil, Config, ethash.NewFaker(), vm.Config{}, nil, nil)

	e := &EVMEmulator{
		database:   db,
		blockchain: blockchain,
	}
	e.Rollback()
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
	if _, err := e.blockchain.InsertChain([]*types.Block{e.pendingBlock}); err != nil {
		panic(err) // This cannot happen unless the simulator is wrong, fail in that case
	}
	e.Rollback()
}

func (e *EVMEmulator) newBlock(f func(int, *core.BlockGen)) (*types.Block, []*types.Receipt) {
	blocks, receipts := core.GenerateChain(e.blockchain.Config(), e.blockchain.CurrentBlock(), ethash.NewFaker(), e.database, 1, f)
	return blocks[0], receipts[0]
}

// Rollback aborts all pending transactions, reverting to the last committed state.
func (e *EVMEmulator) Rollback() {
	e.pendingBlock, _ = e.newBlock(func(int, *core.BlockGen) {})
	e.pendingState, _ = state.New(e.pendingBlock.Root(), e.blockchain.StateCache(), nil)
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
	receipt, _, _, _ := rawdb.ReadReceipt(e.database, txHash, e.blockchain.Config())
	return receipt, nil
}

// TransactionByHash checks the pool of pending transactions in addition to the
// blockchain. The isPending return value indicates whether the transaction has been
// mined yet. Note that the transaction may not be part of the canonical chain even if
// it's not pending.
func (e *EVMEmulator) TransactionByHash(txHash common.Hash) (*types.Transaction, bool, error) {
	tx := e.pendingBlock.Transaction(txHash)
	if tx != nil {
		return tx, true, nil
	}
	tx, _, _, _ = rawdb.ReadTransaction(e.database, txHash)
	if tx != nil {
		return tx, false, nil
	}
	return nil, false, ethereum.NotFound
}

// BlockByHash retrieves a block based on the block hash.
func (e *EVMEmulator) BlockByHash(hash common.Hash) (*types.Block, error) {
	if hash == e.pendingBlock.Hash() {
		return e.pendingBlock, nil
	}
	block := e.blockchain.GetBlockByHash(hash)
	if block != nil {
		return block, nil
	}
	return nil, ErrBlockDoesNotExist
}

// BlockByNumber retrieves a block from the database by number, caching it
// (associated with its hash) if found.
func (e *EVMEmulator) BlockByNumber(number *big.Int) (*types.Block, error) {
	return e.blockByNumberNoLock(number)
}

// blockByNumberNoLock retrieves a block from the database by number, caching it
// (associated with its hash) if found without Lock.
func (e *EVMEmulator) blockByNumberNoLock(number *big.Int) (*types.Block, error) {
	if number == nil || number.Cmp(e.pendingBlock.Number()) == 0 {
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
	if hash == e.pendingBlock.Hash() {
		return e.pendingBlock.Header(), nil
	}
	header := e.blockchain.GetHeaderByHash(hash)
	if header == nil {
		return nil, ErrBlockDoesNotExist
	}
	return header, nil
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (e *EVMEmulator) HeaderByNumber(block *big.Int) (*types.Header, error) {
	if block == nil || block.Cmp(e.pendingBlock.Number()) == 0 {
		return e.blockchain.CurrentHeader(), nil
	}
	return e.blockchain.GetHeaderByNumber(uint64(block.Int64())), nil
}

// TransactionCount returns the number of transactions in a given block.
func (e *EVMEmulator) TransactionCount(blockHash common.Hash) (uint, error) {
	if blockHash == e.pendingBlock.Hash() {
		return uint(e.pendingBlock.Transactions().Len()), nil
	}
	block := e.blockchain.GetBlockByHash(blockHash)
	if block == nil {
		return uint(0), ErrBlockDoesNotExist
	}
	return uint(block.Transactions().Len()), nil
}

// TransactionInBlock returns the transaction for a specific block at a specific index.
func (e *EVMEmulator) TransactionInBlock(blockHash common.Hash, index uint) (*types.Transaction, error) {
	if blockHash == e.pendingBlock.Hash() {
		transactions := e.pendingBlock.Transactions()
		if uint(len(transactions)) < index+1 {
			return nil, ErrTransactionDoesNotExist
		}

		return transactions[index], nil
	}

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

// PendingCodeAt returns the code associated with an account in the pending state.
func (e *EVMEmulator) PendingCodeAt(contract common.Address) ([]byte, error) {
	return e.pendingState.GetCode(contract), nil
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
	if blockNumber != nil && blockNumber.Cmp(e.blockchain.CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	stateDB, err := e.blockchain.State()
	if err != nil {
		return nil, err
	}
	res, err := e.callContract(call, e.blockchain.CurrentBlock(), stateDB)
	if err != nil {
		return nil, err
	}
	// If the result contains a revert reason, try to unpack and return it.
	if len(res.Revert()) > 0 {
		return nil, newRevertError(res)
	}
	return res.Return(), res.Err
}

// PendingCallContract executes a contract call on the pending state.
func (e *EVMEmulator) PendingCallContract(call ethereum.CallMsg) ([]byte, error) {
	defer e.pendingState.RevertToSnapshot(e.pendingState.Snapshot())

	res, err := e.callContract(call, e.pendingBlock, e.pendingState)
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
	return e.pendingState.GetOrNewStateObject(account).Nonce(), nil
}

// SuggestGasPrice implements ContractTransactor.SuggestGasPrice. Since the simulated
// chain doesn't have miners, we just return a gas price of 1 for any call.
func (e *EVMEmulator) SuggestGasPrice() (*big.Int, error) {
	return GasPrice, nil
}

// EstimateGas executes the requested code against the currently pending block/state and
// returns the used amount of gas.
func (e *EVMEmulator) EstimateGas(call ethereum.CallMsg) (uint64, error) {
	// Determine the lowest and highest possible gas limits to binary search in between
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)
	if call.Gas >= params.TxGas {
		hi = call.Gas
	} else {
		hi = e.pendingBlock.GasLimit()
	}
	// Recap the highest gas allowance with account's balance.
	if call.GasPrice != nil && call.GasPrice.BitLen() != 0 {
		balance := e.pendingState.GetBalance(call.From) // from can't be nil
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
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *core.ExecutionResult, error) {
		call.Gas = gas

		snapshot := e.pendingState.Snapshot()
		res, err := e.callContract(call, e.pendingBlock, e.pendingState)
		e.pendingState.RevertToSnapshot(snapshot)

		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, fmt.Errorf("Bail out: %w", err)
		}
		return res.Failed(), res, nil
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
	if hi == cap {
		failed, result, err := executable(hi)
		if err != nil {
			return 0, fmt.Errorf("executable(hi): %w", err)
		}
		if failed {
			if result != nil && result.Err != vm.ErrOutOfGas {
				if len(result.Revert()) > 0 {
					return 0, newRevertError(result)
				}
				return 0, fmt.Errorf("revert: %w", result.Err)
			}
			// Otherwise, the specified gas cap is too low
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", cap)
		}
	}
	return hi, nil
}

// callContract implements common code between normal and pending contract calls.
// state is modified during execution, make sure to copy it if necessary.
func (e *EVMEmulator) callContract(call ethereum.CallMsg, block *types.Block, stateDB *state.StateDB) (*core.ExecutionResult, error) {
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
	evmContext := core.NewEVMBlockContext(block.Header(), e.blockchain, nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmEnv := vm.NewEVM(evmContext, txContext, stateDB, e.blockchain.Config(), vm.Config{})
	gasPool := new(core.GasPool).AddGas(math.MaxUint64)

	return core.NewStateTransition(vmEnv, msg, gasPool).TransitionDb()
}

// SendTransaction updates the pending block to include the given transaction.
// It returns an error if the transaction is invalid.
func (e *EVMEmulator) SendTransaction(tx *types.Transaction) (uint64, error) {
	sender, err := types.Sender(types.NewEIP155Signer(e.blockchain.Config().ChainID), tx)
	if err != nil {
		return 0, xerrors.Errorf("invalid transaction: %w", err)
	}
	nonce := e.pendingState.GetNonce(sender)
	if tx.Nonce() != nonce {
		return 0, xerrors.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), nonce)
	}

	block, receipts := e.newBlock(func(number int, block *core.BlockGen) {
		for _, pendingTx := range e.pendingBlock.Transactions() {
			block.AddTxWithChain(e.blockchain, pendingTx)
		}
		block.AddTxWithChain(e.blockchain, tx)
	})
	stateDB, _ := e.blockchain.State()

	e.pendingBlock = block
	e.pendingState, _ = state.New(e.pendingBlock.Root(), stateDB.Database(), nil)

	return receipts[len(receipts)-1].GasUsed, nil
}

// AdjustTime adds a time shift to the simulated clock.
// It can only be called on empty blocks.
func (e *EVMEmulator) AdjustTime(adjustment time.Duration) error {
	if len(e.pendingBlock.Transactions()) != 0 {
		return errors.New("Could not adjust time on non-empty block")
	}

	block, _ := e.newBlock(func(number int, block *core.BlockGen) {
		block.OffsetTime(int64(adjustment.Seconds()))
	})
	stateDB, _ := e.blockchain.State()

	e.pendingBlock = block
	e.pendingState, _ = state.New(e.pendingBlock.Root(), stateDB.Database(), nil)

	return nil
}

// Blockchain returns the underlying blockchain.
func (e *EVMEmulator) Blockchain() *core.BlockChain {
	return e.blockchain
}

func (e *EVMEmulator) Signer() types.Signer {
	return Signer()
}

// callMsg implements core.Message to allow passing it as a transaction simulator.
type callMsg struct {
	ethereum.CallMsg
}

func (m callMsg) From() common.Address         { return m.CallMsg.From }
func (m callMsg) Nonce() uint64                { return 0 }
func (m callMsg) CheckNonce() bool             { return false }
func (m callMsg) To() *common.Address          { return m.CallMsg.To }
func (m callMsg) GasPrice() *big.Int           { return m.CallMsg.GasPrice }
func (m callMsg) Gas() uint64                  { return m.CallMsg.Gas }
func (m callMsg) Value() *big.Int              { return m.CallMsg.Value }
func (m callMsg) Data() []byte                 { return m.CallMsg.Data }
func (m callMsg) AccessList() types.AccessList { return m.CallMsg.AccessList }
