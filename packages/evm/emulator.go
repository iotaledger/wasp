// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"golang.org/x/xerrors"
)

type EVMEmulator struct {
	timestamp   uint64
	chainConfig *params.ChainConfig
	kv          kv.KVStore
	IEVMBackend vm.ISCPBackend
}

var (
	TxGas           = uint64(21000) // gas cost of simple transfer (not contract creation / call)
	GasLimitDefault = uint64(15000000)
	GasPrice        = big.NewInt(0)
)

const DefaultChainID = 1074 // IOTA -- get it?

func makeConfig(chainID int) *params.ChainConfig {
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

const (
	keyStateDB      = "s"
	keyBlockchainDB = "b"
)

func newStateDB(store kv.KVStore) *StateDB {
	return NewStateDB(subrealm.New(store, keyStateDB))
}

func newBlockchainDB(store kv.KVStore) *BlockchainDB {
	return NewBlockchainDB(subrealm.New(store, keyBlockchainDB))
}

// Init initializes the EVM state with the provided genesis allocation parameters
func Init(store kv.KVStore, chainID uint16, gasLimit, timestamp uint64, alloc core.GenesisAlloc) {
	bdb := newBlockchainDB(store)
	if bdb.Initialized() {
		panic("evm state already initialized in kvstore")
	}
	bdb.Init(chainID, gasLimit, timestamp)

	statedb := newStateDB(store)
	for addr, account := range alloc {
		statedb.CreateAccount(addr)
		if account.Balance != nil {
			statedb.AddBalance(addr, account.Balance)
		}
		if account.Code != nil {
			statedb.SetCode(addr, account.Code)
		}
		for k, v := range account.Storage {
			statedb.SetState(addr, k, v)
		}
		statedb.SetNonce(addr, account.Nonce)
	}
}

func NewEVMEmulator(store kv.KVStore, timestamp uint64, backend vm.ISCPBackend) *EVMEmulator {
	bdb := newBlockchainDB(store)
	if !bdb.Initialized() {
		panic("must initialize genesis block first")
	}
	return &EVMEmulator{
		timestamp:   timestamp,
		chainConfig: makeConfig(int(bdb.GetChainID())),
		kv:          store,
		IEVMBackend: backend,
	}
}

func (e *EVMEmulator) StateDB() *StateDB {
	return newStateDB(e.kv)
}

func (e *EVMEmulator) BlockchainDB() *BlockchainDB {
	return newBlockchainDB(e.kv)
}

func (e *EVMEmulator) GasLimit() uint64 {
	return e.BlockchainDB().GetGasLimit()
}

func newRevertError(result *core.ExecutionResult) *revertError {
	reason, errUnpack := abi.UnpackRevert(result.Revert())
	err := xerrors.New("execution reverted")
	if errUnpack == nil {
		err = xerrors.Errorf("execution reverted: %v", reason)
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

// CallContract executes a contract call, without committing changes to the state
func (e *EVMEmulator) CallContract(call ethereum.CallMsg) ([]byte, error) {
	res, err := e.callContract(call)
	if err != nil {
		return nil, err
	}
	// If the result contains a revert reason, try to unpack and return it.
	if len(res.Revert()) > 0 {
		return nil, newRevertError(res)
	}
	return res.Return(), res.Err
}

// EstimateGas executes the requested code against the current state and
// returns the used amount of gas, discarding state changes
func (e *EVMEmulator) EstimateGas(call ethereum.CallMsg) (uint64, error) {
	if call.Gas < params.TxGas {
		call.Gas = e.GasLimit()
	}
	res, err := e.callContract(call)
	if err != nil {
		return 0, err
	}
	return res.UsedGas, nil
}

func (e *EVMEmulator) PendingHeader() *types.Header {
	return &types.Header{
		Difficulty: &big.Int{},
		Number:     new(big.Int).Add(e.BlockchainDB().GetNumber(), common.Big1),
		GasLimit:   e.GasLimit(),
		Time:       e.timestamp,
	}
}

func (e *EVMEmulator) ChainContext() core.ChainContext {
	return &chainContext{
		engine: ethash.NewFaker(),
	}
}

func (e *EVMEmulator) callContract(call ethereum.CallMsg) (*core.ExecutionResult, error) {
	header := e.PendingHeader()
	// Ensure message is initialized properly.
	if call.GasPrice == nil {
		call.GasPrice = big.NewInt(0)
	}
	if call.Gas == 0 {
		call.Gas = e.GasLimit()
	}
	if call.Value == nil {
		call.Value = new(big.Int)
	}

	// Execute the call.
	msg := callMsg{call}

	// run the EVM code on a buffered state (so that writes are not committed)
	statedb := e.StateDB().Buffered().StateDB()

	txContext := core.NewEVMTxContext(msg)
	evmContext := core.NewEVMBlockContext(header, e.ChainContext(), nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmEnv := vm.NewEVM(evmContext, txContext, statedb, e.chainConfig, e.vmConfig())
	gasPool := new(core.GasPool).AddGas(math.MaxUint64)

	return core.NewStateTransition(vmEnv, msg, gasPool).TransitionDb()
}

func (e *EVMEmulator) vmConfig() vm.Config {
	return vm.Config{
		JumpTable: vm.NewISCPInstructionSet(e.GetIEVMBackend),
	}
}

func (e *EVMEmulator) GetIEVMBackend() vm.ISCPBackend {
	return e.IEVMBackend
}

// SendTransaction updates the pending block to include the given transaction.
// It returns an error if the transaction is invalid.
func (e *EVMEmulator) SendTransaction(tx *types.Transaction) (*types.Receipt, error) {
	sender, err := types.Sender(e.Signer(), tx)
	if err != nil {
		return nil, xerrors.Errorf("invalid transaction: %w", err)
	}
	nonce := e.StateDB().GetNonce(sender)
	if tx.Nonce() != nonce {
		return nil, xerrors.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), nonce)
	}

	buf := e.StateDB().Buffered()

	gasPool := core.GasPool(e.GasLimit())

	pendingHeader := e.PendingHeader()

	msg, err := tx.AsMessage(types.MakeSigner(e.chainConfig, pendingHeader.Number), pendingHeader.BaseFee)
	if err != nil {
		return nil, err
	}
	statedb := buf.StateDB()
	blockContext := core.NewEVMBlockContext(pendingHeader, e.ChainContext(), nil)
	evm := vm.NewEVM(blockContext, vm.TxContext{}, statedb, e.chainConfig, e.vmConfig())
	txContext := core.NewEVMTxContext(msg)
	evm.Reset(txContext, statedb)
	result, err := core.ApplyMessage(evm, msg, &gasPool)
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{
		Type:              tx.Type(),
		CumulativeGasUsed: result.UsedGas,
		TxHash:            tx.Hash(),
		GasUsed:           result.UsedGas,
		Logs:              statedb.GetLogs(tx.Hash()),
		BlockNumber:       e.PendingHeader().Number,
	}
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	if result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	}

	buf.Commit()

	e.BlockchainDB().AddTransaction(tx, receipt, e.timestamp)

	return receipt, nil
}

// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
func (e *EVMEmulator) FilterLogs(query *ethereum.FilterQuery) []*types.Log {
	receipts := e.getReceiptsInFilterRange(query)
	return e.filterLogs(query, receipts)
}

func (e *EVMEmulator) getReceiptsInFilterRange(query *ethereum.FilterQuery) []*types.Receipt {
	bc := e.BlockchainDB()

	if query.BlockHash != nil {
		blockNumber := bc.GetBlockNumberByBlockHash(*query.BlockHash)
		if blockNumber == nil {
			return nil
		}
		receipt := bc.GetReceiptByBlockNumber(blockNumber)
		if receipt == nil {
			return nil
		}
		return []*types.Receipt{receipt}
	}

	// Initialize unset filter boundaries to run from genesis to chain head
	first := big.NewInt(1) // skip genesis since it has no logs
	last := bc.GetNumber()
	from := first
	if query.FromBlock != nil && query.FromBlock.Cmp(first) >= 0 && query.FromBlock.Cmp(last) <= 0 {
		from = query.FromBlock
	}
	to := last
	if query.ToBlock != nil && query.ToBlock.Cmp(first) >= 0 && query.ToBlock.Cmp(last) <= 0 {
		to = query.ToBlock
	}

	var receipts []*types.Receipt
	for i := new(big.Int).Set(from); i.Cmp(to) <= 0; i = i.Add(i, common.Big1) {
		receipt := bc.GetReceiptByBlockNumber(i)
		if receipt != nil {
			receipts = append(receipts, receipt)
		}
	}
	return receipts
}

func (e *EVMEmulator) filterLogs(query *ethereum.FilterQuery, receipts []*types.Receipt) []*types.Log {
	var logs []*types.Log
	for _, r := range receipts {
		if !bloomFilter(r.Bloom, query.Addresses, query.Topics) {
			continue
		}
		for _, log := range r.Logs {
			if !logMatches(log, query.Addresses, query.Topics) {
				continue
			}
			logs = append(logs, log)
		}
	}
	return logs
}

func bloomFilter(bloom types.Bloom, addresses []common.Address, topics [][]common.Hash) bool {
	if len(addresses) > 0 {
		var included bool
		for _, addr := range addresses {
			if types.BloomLookup(bloom, addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if types.BloomLookup(bloom, topic) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}

func logMatches(log *types.Log, addresses []common.Address, topics [][]common.Hash) bool {
	if len(addresses) > 0 {
		found := false
		for _, a := range addresses {
			if log.Address == a {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	if len(topics) > 0 {
		if len(topics) > len(log.Topics) {
			return false
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
	}
	return true
}

func (e *EVMEmulator) Signer() types.Signer {
	return evmtypes.Signer(e.chainConfig.ChainID)
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

type chainContext struct {
	engine consensus.Engine
}

var _ core.ChainContext = &chainContext{}

func (c *chainContext) Engine() consensus.Engine {
	return c.engine
}

func (c *chainContext) GetHeader(common.Hash, uint64) *types.Header {
	panic("not implemented")
}
