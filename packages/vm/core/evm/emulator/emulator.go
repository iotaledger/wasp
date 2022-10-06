// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmexceptions"
)

type EVMEmulator struct {
	timestamp   uint64
	chainConfig *params.ChainConfig
	kv          kv.KVStore
	vmConfig    vm.Config
	getBalance  GetBalanceFunc
	subBalance  SubBalanceFunc
	addBalance  AddBalanceFunc
}

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

func newStateDB(
	store kv.KVStore,
	getBalance GetBalanceFunc,
	addBalance SubBalanceFunc,
	subBalance AddBalanceFunc,
) *StateDB {
	return NewStateDB(subrealm.New(store, keyStateDB), getBalance, addBalance, subBalance)
}

func newBlockchainDB(store kv.KVStore) *BlockchainDB {
	return NewBlockchainDB(subrealm.New(store, keyBlockchainDB))
}

// Init initializes the EVM state with the provided genesis allocation parameters
func Init(
	store kv.KVStore,
	chainID uint16,
	blockKeepAmount int32,
	gasLimit,
	timestamp uint64,
	alloc core.GenesisAlloc,
	getBalance GetBalanceFunc,
	debitFromAccount SubBalanceFunc,
	creditToAccount AddBalanceFunc,
) {
	bdb := newBlockchainDB(store)
	if bdb.Initialized() {
		panic("evm state already initialized in kvstore")
	}
	bdb.Init(chainID, blockKeepAmount, gasLimit, timestamp)

	statedb := newStateDB(store, getBalance, debitFromAccount, creditToAccount)
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

func NewEVMEmulator(
	store kv.KVStore,
	timestamp uint64,
	magicContract vm.ISCContract,
	getBalance GetBalanceFunc,
	subBalance SubBalanceFunc,
	addBalance AddBalanceFunc,
) *EVMEmulator {
	bdb := newBlockchainDB(store)
	if !bdb.Initialized() {
		panic("must initialize genesis block first")
	}

	return &EVMEmulator{
		timestamp:   timestamp,
		chainConfig: makeConfig(int(bdb.GetChainID())),
		kv:          store,
		vmConfig:    vm.Config{ISCContract: magicContract},
		getBalance:  getBalance,
		subBalance:  subBalance,
		addBalance:  addBalance,
	}
}

func (e *EVMEmulator) StateDB() *StateDB {
	return newStateDB(e.kv, e.getBalance, e.subBalance, e.addBalance)
}

func (e *EVMEmulator) BlockchainDB() *BlockchainDB {
	return newBlockchainDB(e.kv)
}

func (e *EVMEmulator) GasLimit() uint64 {
	return e.BlockchainDB().GetGasLimit()
}

func (e *EVMEmulator) ChainContext() core.ChainContext {
	return &chainContext{
		engine: ethash.NewFaker(),
	}
}

func (e *EVMEmulator) estimateGas(callMsg ethereum.CallMsg) (uint64, error) {
	lo := params.TxGas
	hi := e.GasLimit()
	lastOk := uint64(0)
	var lastErr error
	for hi >= lo {
		callMsg.Gas = (lo + hi) / 2
		res, err := e.CallContract(callMsg, nil)
		if err != nil {
			return 0, err
		}
		if res.Err != nil {
			lastErr = res.Err
			lo = callMsg.Gas + 1
		} else {
			lastOk = callMsg.Gas
			hi = callMsg.Gas - 1
		}
	}
	if lastOk == 0 {
		if lastErr != nil {
			return 0, xerrors.Errorf("estimateGas failed: %s", lastErr.Error())
		}
		return 0, xerrors.Errorf("estimateGas failed")
	}
	return lastOk, nil
}

// CallContract executes a contract call, without committing changes to the state
func (e *EVMEmulator) CallContract(call ethereum.CallMsg, gasBurnEnable func(bool)) (*core.ExecutionResult, error) {
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

	msg := callMsg{call}
	pendingHeader := e.BlockchainDB().GetPendingHeader()

	// run the EVM code on a buffered state (so that writes are not committed)
	statedb := e.StateDB().Buffered().StateDB()

	return e.applyMessage(msg, statedb, pendingHeader, gasBurnEnable)
}

func (e *EVMEmulator) applyMessage(msg core.Message, statedb vm.StateDB, header *types.Header, gasBurnEnable func(bool)) (res *core.ExecutionResult, err error) {
	blockContext := core.NewEVMBlockContext(header, e.ChainContext(), nil)
	txContext := core.NewEVMTxContext(msg)
	vmEnv := vm.NewEVM(blockContext, txContext, statedb, e.chainConfig, e.vmConfig)
	gasPool := core.GasPool(msg.Gas())
	vmEnv.Reset(txContext, statedb)
	if gasBurnEnable != nil {
		gasBurnEnable(true)
		defer gasBurnEnable(false)
	}

	caughtErr := panicutil.CatchAllExcept(
		func() {
			// catch any exceptions during the execution, so that an EVM receipt is produced
			res, err = core.ApplyMessage(vmEnv, msg, &gasPool)
		},
		vmexceptions.AllProtocolLimits...,
	)
	if caughtErr != nil {
		return nil, caughtErr
	}
	return res, err
}

func (e *EVMEmulator) SendTransaction(tx *types.Transaction, gasBurnEnable func(bool)) (*types.Receipt, *core.ExecutionResult, error) {
	buf := e.StateDB().Buffered()
	statedb := buf.StateDB()
	pendingHeader := e.BlockchainDB().GetPendingHeader()

	sender, err := types.Sender(e.Signer(), tx)
	if err != nil {
		return nil, nil, xerrors.Errorf("invalid transaction: %w", err)
	}
	nonce := e.StateDB().GetNonce(sender)
	if tx.Nonce() != nonce {
		return nil, nil, xerrors.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), nonce)
	}

	msg, err := tx.AsMessage(types.MakeSigner(e.chainConfig, pendingHeader.Number), pendingHeader.BaseFee)
	if err != nil {
		return nil, nil, err
	}

	msgWithZeroGasPrice := callMsg{
		CallMsg: ethereum.CallMsg{
			From:       msg.From(),
			To:         msg.To(),
			Gas:        msg.Gas(),
			GasPrice:   big.NewInt(0),
			GasFeeCap:  big.NewInt(0),
			GasTipCap:  big.NewInt(0),
			Value:      msg.Value(),
			Data:       msg.Data(),
			AccessList: msg.AccessList(),
		},
	}

	result, err := e.applyMessage(msgWithZeroGasPrice, statedb, pendingHeader, gasBurnEnable)

	gasUsed := uint64(0)
	if result != nil {
		gasUsed = result.UsedGas
	}

	cumulativeGasUsed := gasUsed
	index := uint(0)
	latest := e.BlockchainDB().GetLatestPendingReceipt()
	if latest != nil {
		cumulativeGasUsed += latest.CumulativeGasUsed
		index = latest.TransactionIndex + 1
	}

	receipt := &types.Receipt{
		Type:              tx.Type(),
		CumulativeGasUsed: cumulativeGasUsed,
		TxHash:            tx.Hash(),
		GasUsed:           gasUsed,
		Logs:              statedb.GetLogs(tx.Hash()),
		BlockNumber:       pendingHeader.Number,
		TransactionIndex:  index,
	}
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	if result == nil || result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}

	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(msg.From(), tx.Nonce())
	}

	buf.Commit()
	e.BlockchainDB().AddTransaction(tx, receipt)

	return receipt, result, err
}

func (e *EVMEmulator) MintBlock() {
	e.BlockchainDB().MintBlock(e.timestamp)
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
		blockNumber, ok := bc.GetBlockNumberByBlockHash(*query.BlockHash)
		if !ok {
			return nil
		}
		return bc.GetReceiptsByBlockNumber(blockNumber)
	}

	// Initialize unset filter boundaries to run from genesis to chain head
	first := big.NewInt(1) // skip genesis since it has no logs
	last := new(big.Int).SetUint64(bc.GetNumber())
	from := first
	if query.FromBlock != nil && query.FromBlock.Cmp(first) >= 0 && query.FromBlock.Cmp(last) <= 0 {
		from = query.FromBlock
	}
	to := last
	if query.ToBlock != nil && query.ToBlock.Cmp(first) >= 0 && query.ToBlock.Cmp(last) <= 0 {
		to = query.ToBlock
	}

	var receipts []*types.Receipt
	{
		to := to.Uint64()
		for i := from.Uint64(); i <= to; i++ {
			receipts = append(receipts, bc.GetReceiptsByBlockNumber(i)...)
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
	return evmutil.Signer(e.chainConfig.ChainID)
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
