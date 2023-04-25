// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/labstack/gommon/log"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
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

type EVMChain struct {
	backend  ChainBackend
	chainID  uint16 // cache
	newBlock *event.Event1[*NewBlockEvent]
	log      *logger.Logger
}

type NewBlockEvent struct {
	block *types.Block
	logs  []*types.Log
}

func NewEVMChain(backend ChainBackend, pub *publisher.Publisher, log *logger.Logger) *EVMChain {
	e := &EVMChain{
		backend:  backend,
		newBlock: event.New1[*NewBlockEvent](),
		log:      log,
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
			e.publishNewBlock(ev.BlockInfo.BlockIndex(), ev.TrieRoot)
		}
	}()

	return e
}

func (e *EVMChain) publishNewBlock(blockIndex uint32, trieRoot trie.Hash) {
	state, err := e.backend.ISCStateByTrieRoot(trieRoot)
	if err != nil {
		log.Errorf("EVMChain.publishNewBlock(blockIndex=%v): ISCStateByTrieRoot returned error: %v", blockIndex, err)
		return
	}
	blockNumber := evmBlockNumberByISCBlockIndex(blockIndex)
	db := blockchainDB(state)
	block := db.GetBlockByNumber(blockNumber)
	if block == nil {
		log.Errorf("EVMChain.publishNewBlock(blockIndex=%v) GetBlockByNumber: block not found", blockIndex)
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
		db := blockchainDB(e.backend.ISCLatestState())
		e.chainID = db.GetChainID()
	}
	return e.chainID
}

func (e *EVMChain) ViewCaller(chainState state.State) vmerrors.ViewCaller {
	e.log.Debugf("ViewCaller(chainState=%v)", chainState)
	return func(contractName string, funcName string, params dict.Dict) (dict.Dict, error) {
		return e.backend.ISCCallView(chainState, contractName, funcName, params)
	}
}

func (e *EVMChain) BlockNumber() *big.Int {
	e.log.Debugf("BlockNumber()")
	db := blockchainDB(e.backend.ISCLatestState())
	return big.NewInt(0).SetUint64(db.GetNumber())
}

func (e *EVMChain) GasRatio() util.Ratio32 {
	e.log.Debugf("GasRatio()")
	govPartition := subrealm.NewReadOnly(e.backend.ISCLatestState(), kv.Key(governance.Contract.Hname().Bytes()))
	gasFeePolicy := governance.MustGetGasFeePolicy(govPartition)
	return gasFeePolicy.EVMGasRatio
}

func (e *EVMChain) GasFeePolicy() *gas.FeePolicy {
	govPartition := subrealm.NewReadOnly(e.backend.ISCLatestState(), kv.Key(governance.Contract.Hname().Bytes()))
	gasFeePolicy := governance.MustGetGasFeePolicy(govPartition)
	return gasFeePolicy
}

func (e *EVMChain) gasLimits() *gas.Limits {
	govPartition := subrealm.NewReadOnly(e.backend.ISCLatestState(), kv.Key(governance.Contract.Hname().Bytes()))
	gasLimits := governance.MustGetGasLimits(govPartition)
	return gasLimits
}

func (e *EVMChain) SendTransaction(tx *types.Transaction) error {
	e.log.Debugf("SendTransaction(tx=%v)", tx)
	chainID := e.ChainID()
	if tx.ChainId().Uint64() != uint64(chainID) {
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
	if tx.Nonce() != expectedNonce {
		return fmt.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), expectedNonce)
	}

	if err := e.checkEnoughL2FundsForGasBudget(sender, tx.Gas()); err != nil {
		return err
	}
	return e.backend.EVMSendTransaction(tx)
}

func (e *EVMChain) checkEnoughL2FundsForGasBudget(sender common.Address, evmGas uint64) error {
	gasRatio := e.GasRatio()
	balance, err := e.Balance(sender, nil)
	if err != nil {
		return fmt.Errorf("could not fetch sender balance: %w", err)
	}
	gasFeePolicy := e.GasFeePolicy()
	iscGasBudgetAffordable := gasFeePolicy.GasBudgetFromTokens(balance.Uint64())

	iscGasBudgetTx := gas.EVMGasToISC(evmGas, &gasRatio)

	gasLimits := e.gasLimits()

	if iscGasBudgetTx > gasLimits.MaxGasPerRequest {
		iscGasBudgetTx = gasLimits.MaxGasPerRequest
	}

	if iscGasBudgetAffordable < iscGasBudgetTx {
		return fmt.Errorf(
			"sender doesn't have enough L2 funds to cover tx gas budget. Balance: %v, expected: %d",
			balance.String(),
			gasFeePolicy.FeeFromGas(iscGasBudgetTx),
		)
	}
	return nil
}

func (e *EVMChain) iscStateFromEVMBlockNumber(blockNumber *big.Int) (state.State, error) {
	if blockNumber == nil {
		return e.backend.ISCLatestState(), nil
	}
	iscBlockIndex, err := iscBlockIndexByEVMBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	return e.backend.ISCStateByBlockIndex(iscBlockIndex)
}

func (e *EVMChain) iscStateFromEVMBlockNumberOrHash(blockNumberOrHash *rpc.BlockNumberOrHash) (state.State, error) {
	if blockNumberOrHash == nil {
		return e.backend.ISCLatestState(), nil
	}
	if blockNumber, ok := blockNumberOrHash.Number(); ok {
		return e.iscStateFromEVMBlockNumber(parseBlockNumber(blockNumber))
	}
	blockHash, _ := blockNumberOrHash.Hash()
	block := e.BlockByHash(blockHash)
	return e.iscStateFromEVMBlockNumber(block.Number())
}

func (e *EVMChain) iscAliasOutputFromEVMBlockNumber(blockNumber *big.Int) (*isc.AliasOutputWithID, error) {
	if blockNumber == nil || blockNumber.Cmp(big.NewInt(int64(e.backend.ISCLatestState().BlockIndex()))) == 0 {
		return e.backend.ISCLatestAliasOutput()
	}
	iscBlockIndex, err := iscBlockIndexByEVMBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	latestBlockIndex := e.backend.ISCLatestState().BlockIndex()
	if iscBlockIndex > latestBlockIndex {
		return nil, fmt.Errorf("no EVM block with number %s", blockNumber)
	}
	if iscBlockIndex == latestBlockIndex {
		return e.backend.ISCLatestAliasOutput()
	}
	nextISCBlockIndex := iscBlockIndex + 1
	nextISCState, err := e.backend.ISCStateByBlockIndex(nextISCBlockIndex)
	if err != nil {
		return nil, err
	}
	blocklogStatePartition := subrealm.NewReadOnly(nextISCState, kv.Key(blocklog.Contract.Hname().Bytes()))
	nextBlock, err := blocklog.GetBlockInfo(blocklogStatePartition, nextISCBlockIndex)
	if err != nil {
		return nil, err
	}
	return nextBlock.PreviousAliasOutput, nil
}

func (e *EVMChain) iscAliasOutputFromEVMBlockNumberOrHash(blockNumberOrHash *rpc.BlockNumberOrHash) (*isc.AliasOutputWithID, error) {
	if blockNumberOrHash == nil {
		return e.backend.ISCLatestAliasOutput()
	}
	if blockNumber, ok := blockNumberOrHash.Number(); ok {
		return e.iscAliasOutputFromEVMBlockNumber(parseBlockNumber(blockNumber))
	}
	blockHash, _ := blockNumberOrHash.Hash()
	block := e.BlockByHash(blockHash)
	if block == nil {
		return nil, fmt.Errorf("block with hash %s not found", blockHash)
	}
	return e.iscAliasOutputFromEVMBlockNumber(block.Number())
}

func (e *EVMChain) Balance(address common.Address, blockNumberOrHash *rpc.BlockNumberOrHash) (*big.Int, error) {
	e.log.Debugf("Balance(address=%v, blockNumberOrHash=%v)", address, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	db := stateDB(chainState)
	return db.GetBalance(address), nil
}

func (e *EVMChain) Code(address common.Address, blockNumberOrHash *rpc.BlockNumberOrHash) ([]byte, error) {
	e.log.Debugf("Code(address=%v, blockNumberOrHash=%v)", address, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	db := stateDB(chainState)
	return db.GetCode(address), nil
}

func (e *EVMChain) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	e.log.Debugf("BlockByNumber(blockNumber=%v)", blockNumber)
	chainState, err := e.iscStateFromEVMBlockNumber(blockNumber)
	if err != nil {
		return nil, err
	}
	return e.blockByNumber(chainState, blockNumber)
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

func (e *EVMChain) TransactionByHash(hash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	e.log.Debugf("TransactionByHash(hash=%v)", hash)
	db := blockchainDB(e.backend.ISCLatestState())
	blockNumber, ok := db.GetBlockNumberByTxHash(hash)
	if !ok {
		return nil, common.Hash{}, 0, 0, err
	}
	tx = db.GetTransactionByHash(hash)
	block := db.GetBlockByNumber(blockNumber)
	txIndex := uint64(0)
	for i, t := range block.Transactions() {
		if t.Hash() == tx.Hash() {
			txIndex = uint64(i)
			break
		}
	}
	return tx, block.Hash(), blockNumber, txIndex, nil
}

func (e *EVMChain) TransactionByBlockHashAndIndex(hash common.Hash, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumber, indexRet uint64, err error) {
	e.log.Debugf("TransactionByBlockHashAndIndex(hash=%v, index=%v)", hash, index)
	db := blockchainDB(e.backend.ISCLatestState())
	block := db.GetBlockByHash(hash)
	if block == nil {
		return nil, common.Hash{}, 0, 0, err
	}
	txs := block.Transactions()
	return txs[index], hash, block.Number().Uint64(), index, nil
}

func (e *EVMChain) TransactionByBlockNumberAndIndex(blockNumber *big.Int, index uint64) (tx *types.Transaction, blockHash common.Hash, blockNumberRet, indexRet uint64, err error) {
	e.log.Debugf("TransactionByBlockNumberAndIndex(blockNumber=%v, index=%v)", blockNumber, index)
	db := blockchainDB(e.backend.ISCLatestState())
	bn, err := blockNumberU64(db, blockNumber)
	if err != nil {
		return nil, common.Hash{}, 0, 0, err
	}
	block := db.GetBlockByNumber(bn)
	if block == nil {
		return nil, common.Hash{}, 0, 0, err
	}
	txs := block.Transactions()
	return txs[index], block.Hash(), bn, index, nil
}

func (e *EVMChain) BlockByHash(hash common.Hash) *types.Block {
	e.log.Debugf("BlockByHash(hash=%v)", hash)
	db := blockchainDB(e.backend.ISCLatestState())
	block := db.GetBlockByHash(hash)
	return block
}

func (e *EVMChain) TransactionReceipt(txHash common.Hash) *types.Receipt {
	e.log.Debugf("TransactionReceipt(txHash=%v)", txHash)
	db := blockchainDB(e.backend.ISCLatestState())
	return db.GetReceiptByTxHash(txHash)
}

func (e *EVMChain) TransactionCount(address common.Address, blockNumberOrHash *rpc.BlockNumberOrHash) (uint64, error) {
	e.log.Debugf("TransactionCount(address=%v, blockNumberOrHash=%v)", address, blockNumberOrHash)
	chainState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return 0, err
	}
	db := stateDB(chainState)
	return db.GetNonce(address), nil
}

func (e *EVMChain) CallContract(callMsg ethereum.CallMsg, blockNumberOrHash *rpc.BlockNumberOrHash) ([]byte, error) {
	e.log.Debugf("CallContract(callMsg=..., blockNumberOrHash=%v)", blockNumberOrHash)
	aliasOutput, err := e.iscAliasOutputFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return nil, err
	}
	return e.backend.EVMCall(aliasOutput, callMsg)
}

func (e *EVMChain) EstimateGas(callMsg ethereum.CallMsg, blockNumberOrHash *rpc.BlockNumberOrHash) (uint64, error) {
	e.log.Debugf("EstimateGas(callMsg=..., blockNumberOrHash=%v)", blockNumberOrHash)
	aliasOutput, err := e.iscAliasOutputFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return 0, err
	}
	return e.backend.EVMEstimateGas(aliasOutput, callMsg)
}

func (e *EVMChain) GasPrice() *big.Int {
	e.log.Debugf("GasPrice()")

	iscState := e.backend.ISCLatestState()
	governancePartition := subrealm.NewReadOnly(iscState, kv.Key(governance.Contract.Hname().Bytes()))
	feePolicy := governance.MustGetGasFeePolicy(governancePartition)

	// convert to wei (18 decimals)
	decimalsDifference := 18 - parameters.L1().BaseToken.Decimals
	price := big.NewInt(10)
	price.Exp(price, new(big.Int).SetUint64(uint64(decimalsDifference)), nil)

	price.Mul(price, new(big.Int).SetUint64(uint64(feePolicy.GasPerToken.B)))
	price.Div(price, new(big.Int).SetUint64(uint64(feePolicy.GasPerToken.A)))
	price.Mul(price, new(big.Int).SetUint64(uint64(feePolicy.EVMGasRatio.A)))
	price.Div(price, new(big.Int).SetUint64(uint64(feePolicy.EVMGasRatio.B)))

	return price
}

func (e *EVMChain) StorageAt(address common.Address, key common.Hash, blockNumberOrHash *rpc.BlockNumberOrHash) (common.Hash, error) {
	e.log.Debugf("StorageAt(address=%v, key=%v, blockNumberOrHash=%v)", address, key, blockNumberOrHash)
	latestState, err := e.iscStateFromEVMBlockNumberOrHash(blockNumberOrHash)
	if err != nil {
		return common.Hash{}, err
	}
	db := stateDB(latestState)
	return db.GetState(address, key), nil
}

func (e *EVMChain) BlockTransactionCountByHash(blockHash common.Hash) uint64 {
	e.log.Debugf("BlockTransactionCountByHash(blockHash=%v)", blockHash)
	block := e.BlockByHash(blockHash)
	if block == nil {
		return 0
	}
	return uint64(len(block.Transactions()))
}

func (e *EVMChain) BlockTransactionCountByNumber(blockNumber *big.Int) (uint64, error) {
	e.log.Debugf("BlockTransactionCountByNumber(blockNumber=%v)", blockNumber)
	block, err := e.BlockByNumber(blockNumber)
	if err != nil {
		return 0, err
	}
	return uint64(len(block.Transactions())), nil
}

func (e *EVMChain) Logs(q *ethereum.FilterQuery) ([]*types.Log, error) {
	e.log.Debugf("Logs(q=%v)", q)
	db := blockchainDB(e.backend.ISCLatestState())
	return db.FilterLogs(q)
}

func (e *EVMChain) BaseToken() *parameters.BaseToken {
	e.log.Debugf("BaseToken()")
	return e.backend.BaseToken()
}

func (e *EVMChain) SubscribeNewHeads(ch chan<- *types.Header) (unsubscribe func()) {
	e.log.Debugf("SubscribeNewHeads(ch=?)")
	return e.newBlock.Hook(func(ev *NewBlockEvent) {
		ch <- ev.block.Header()
	}).Unhook
}

func (e *EVMChain) SubscribeLogs(q *ethereum.FilterQuery, ch chan<- []*types.Log) (unsubscribe func()) {
	e.log.Debugf("SubscribeLogs(q=%v, ch=?)", q)
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

func (e *EVMChain) iscRequestsInBlock(blockIndex uint32) (*blocklog.BlockInfo, []isc.Request, error) {
	iscState, err := e.backend.ISCStateByBlockIndex(blockIndex)
	if err != nil {
		return nil, nil, err
	}
	reqIDs, err := blocklog.GetRequestIDsForBlock(iscState, blockIndex)
	if err != nil {
		return nil, nil, err
	}
	reqs := make([]isc.Request, len(reqIDs))
	for i, reqID := range reqIDs {
		var receipt *blocklog.RequestReceipt
		receipt, err = blocklog.GetRequestReceipt(iscState, reqID)
		if err != nil {
			return nil, nil, err
		}
		reqs[i] = receipt.Request
	}
	blocklogStatePartition := subrealm.NewReadOnly(iscState, kv.Key(blocklog.Contract.Hname().Bytes()))
	block, err := blocklog.GetBlockInfo(blocklogStatePartition, blockIndex)
	if err != nil {
		return nil, nil, err
	}
	return block, reqs, nil
}

func (e *EVMChain) TraceTransaction(txHash common.Hash, config *tracers.TraceConfig) (any, error) {
	e.log.Debugf("TraceTransaction(txHash=%v, config=?)", txHash)
	tracerType := "callTracer"
	if config.Tracer != nil {
		tracerType = *config.Tracer
	}
	tracer, err := newTracer(tracerType, config.TracerConfig)
	if err != nil {
		return nil, err
	}

	_, _, blockNumber, txIndex, err := e.TransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	if blockNumber == 0 {
		return nil, errors.New("tx not found")
	}

	blockIndex, err := iscBlockIndexByEVMBlockNumber(big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return nil, err
	}
	iscBlock, iscRequestsInBlock, err := e.iscRequestsInBlock(blockIndex)
	if err != nil {
		return nil, err
	}

	err = e.backend.EVMTraceTransaction(
		iscBlock.PreviousAliasOutput,
		iscBlock.Timestamp,
		iscRequestsInBlock,
		txIndex,
		tracer,
	)
	if err != nil {
		return nil, err
	}

	return tracer.GetResult()
}

var maxUint32 = big.NewInt(math.MaxUint32)

// the first EVM block (number 0) is "minted" at ISC block index 1 (init chain)
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
	govPartition := subrealm.NewReadOnly(chainState, kv.Key(governance.Contract.Hname().Bytes()))
	gasLimits := governance.MustGetGasLimits(govPartition)
	gasFeePolicy := governance.MustGetGasFeePolicy(govPartition)
	bdbPartition := subrealm.NewReadOnly(
		chainState,
		kv.Key(evm.Contract.Hname().Bytes())+evm.KeyEVMState+emulator.KeyBlockchainDB,
	)
	return emulator.NewBlockchainDB(
		buffered.NewBufferedKVStore(bdbPartition),
		gas.EVMBlockGasLimit(gasLimits, &gasFeePolicy.EVMGasRatio),
	)
}

func stateDB(chainState state.State) *emulator.StateDB {
	sdbPartition := subrealm.NewReadOnly(
		chainState,
		kv.Key(evm.Contract.Hname().Bytes())+evm.KeyEVMState+emulator.KeyStateDB,
	)
	accountsPartition := subrealm.NewReadOnly(chainState, kv.Key(accounts.Contract.Hname().Bytes()))
	return emulator.NewStateDB(
		buffered.NewBufferedKVStore(sdbPartition),
		newL2Balance(accountsPartition),
	)
}

type l2BalanceR struct {
	accounts kv.KVStoreReader
}

func newL2Balance(accounts kv.KVStoreReader) *l2BalanceR {
	return &l2BalanceR{
		accounts: accounts,
	}
}

func (b *l2BalanceR) Get(addr common.Address) *big.Int {
	bal := accounts.GetBaseTokensBalance(b.accounts, isc.NewEthereumAddressAgentID(addr))
	decimals := parameters.L1().BaseToken.Decimals
	ret := new(big.Int).SetUint64(bal)
	return util.CustomTokensDecimalsToEthereumDecimals(ret, decimals)
}

func (b *l2BalanceR) Add(addr common.Address, amount *big.Int) {
	panic("should not be called")
}

func (b *l2BalanceR) Sub(addr common.Address, amount *big.Int) {
	panic("should not be called")
}
