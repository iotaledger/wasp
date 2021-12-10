// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

const (
	keyChainID                  = "c"
	keyGasLimit                 = "g"
	keyNumber                   = "n"
	keyTransactionByBlockNumber = "n:t"
	keyReceiptByBlockNumber     = "n:r"
	keyTimestampByBlockNumber   = "n:ts"
	keyBlockHashByBlockNumber   = "n:bh"

	// indexes:

	keyBlockNumberByBlockHash = "bh:n"
	keyBlockNumberByTxHash    = "th:n"
)

// Amount of blocks to keep in DB. Older blocks will be pruned every time a transaction is added
const keepAmount = 100

type BlockchainDB struct {
	kv kv.KVStore
}

func NewBlockchainDB(store kv.KVStore) *BlockchainDB {
	return &BlockchainDB{kv: store}
}

func (bc *BlockchainDB) Initialized() bool {
	return bc.kv.MustGet(keyChainID) != nil
}

func (bc *BlockchainDB) Init(chainID uint16, gasLimit, timestamp uint64) {
	bc.SetChainID(chainID)
	bc.SetGasLimit(gasLimit)
	bc.AddBlock(bc.makeHeader(nil, nil, timestamp))
}

func (bc *BlockchainDB) SetChainID(chainID uint16) {
	bc.kv.Set(keyChainID, codec.EncodeUint16(chainID))
}

func (bc *BlockchainDB) GetChainID() uint16 {
	chainID, err := codec.DecodeUint16(bc.kv.MustGet(keyChainID))
	if err != nil {
		panic(err)
	}
	return chainID
}

func (bc *BlockchainDB) SetGasLimit(gas uint64) {
	bc.kv.Set(keyGasLimit, codec.EncodeUint64(gas))
}

func (bc *BlockchainDB) GetGasLimit() uint64 {
	gas, err := codec.DecodeUint64(bc.kv.MustGet(keyGasLimit))
	if err != nil {
		panic(err)
	}
	return gas
}

func (bc *BlockchainDB) SetNumber(n *big.Int) {
	bc.kv.Set(keyNumber, n.Bytes())
}

func (bc *BlockchainDB) GetNumber() *big.Int {
	r := new(big.Int)
	r.SetBytes(bc.kv.MustGet(keyNumber))
	return r
}

func makeTransactionByBlockNumberKey(blockNumber *big.Int) kv.Key {
	return keyTransactionByBlockNumber + kv.Key(blockNumber.Bytes())
}

func makeReceiptByBlockNumberKey(blockNumber *big.Int) kv.Key {
	return keyReceiptByBlockNumber + kv.Key(blockNumber.Bytes())
}

func makeTimestampByBlockNumberKey(blockNumber *big.Int) kv.Key {
	return keyTimestampByBlockNumber + kv.Key(blockNumber.Bytes())
}

func makeBlockHashByBlockNumberKey(blockNumber *big.Int) kv.Key {
	return keyBlockHashByBlockNumber + kv.Key(blockNumber.Bytes())
}

func makeBlockNumberByBlockHashKey(hash common.Hash) kv.Key {
	return keyBlockNumberByBlockHash + kv.Key(hash.Bytes())
}

func makeBlockNumberByTxHashKey(hash common.Hash) kv.Key {
	return keyBlockNumberByTxHash + kv.Key(hash.Bytes())
}

func (bc *BlockchainDB) AddTransaction(tx *types.Transaction, receipt *types.Receipt, timestamp uint64) {
	bc.kv.Set(
		makeTransactionByBlockNumberKey(receipt.BlockNumber),
		evmtypes.EncodeTransaction(tx),
	)
	bc.kv.Set(
		makeReceiptByBlockNumberKey(receipt.BlockNumber),
		evmtypes.EncodeReceipt(receipt),
	)
	bc.kv.Set(
		makeBlockNumberByTxHashKey(tx.Hash()),
		receipt.BlockNumber.Bytes(),
	)
	header := bc.makeHeader(tx, receipt, timestamp)
	bc.AddBlock(header)

	bc.prune(header.Number)
}

func (bc *BlockchainDB) prune(currentNumber *big.Int) {
	forget := new(big.Int).Sub(currentNumber, big.NewInt(int64(keepAmount)))
	if forget.Cmp(common.Big1) >= 0 {
		blockHash := bc.GetBlockHashByBlockNumber(forget)
		txHash := bc.GetTransactionByBlockNumber(forget).Hash()

		bc.kv.Del(makeTransactionByBlockNumberKey(forget))
		bc.kv.Del(makeReceiptByBlockNumberKey(forget))
		bc.kv.Del(makeTimestampByBlockNumberKey(forget))
		bc.kv.Del(makeBlockHashByBlockNumberKey(forget))
		bc.kv.Del(makeBlockNumberByBlockHashKey(blockHash))
		bc.kv.Del(makeBlockNumberByTxHashKey(txHash))
	}
}

func (bc *BlockchainDB) AddBlock(header *types.Header) {
	bc.kv.Set(
		makeTimestampByBlockNumberKey(header.Number),
		codec.EncodeUint64(header.Time),
	)
	bc.kv.Set(
		makeBlockHashByBlockNumberKey(header.Number),
		header.Hash().Bytes(),
	)
	bc.kv.Set(
		makeBlockNumberByBlockHashKey(header.Hash()),
		header.Number.Bytes(),
	)
	bc.SetNumber(header.Number)
}

func (bc *BlockchainDB) GetReceiptByBlockNumber(blockNumber *big.Int) *types.Receipt {
	b := bc.kv.MustGet(makeReceiptByBlockNumberKey(blockNumber))
	if b == nil {
		return nil
	}
	r, err := evmtypes.DecodeReceipt(b)
	if err != nil {
		panic(err)
	}
	tx := bc.GetTransactionByBlockNumber(blockNumber)
	r.TxHash = tx.Hash()
	if tx.To() == nil {
		from, _ := types.Sender(evmtypes.Signer(big.NewInt(int64(bc.GetChainID()))), tx)
		r.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	}
	r.GasUsed = r.CumulativeGasUsed
	r.BlockHash = bc.GetBlockHashByBlockNumber(blockNumber)
	r.BlockNumber = blockNumber
	return r
}

func (bc *BlockchainDB) GetBlockNumberByTxHash(txHash common.Hash) *big.Int {
	b := bc.kv.MustGet(makeBlockNumberByTxHashKey(txHash))
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}

func (bc *BlockchainDB) GetReceiptByTxHash(txHash common.Hash) *types.Receipt {
	blockNumber := bc.GetBlockNumberByTxHash(txHash)
	if blockNumber == nil {
		return nil
	}
	return bc.GetReceiptByBlockNumber(blockNumber)
}

func (bc *BlockchainDB) GetTransactionByBlockNumber(blockNumber *big.Int) *types.Transaction {
	b := bc.kv.MustGet(makeTransactionByBlockNumberKey(blockNumber))
	if b == nil {
		return nil
	}
	tx, err := evmtypes.DecodeTransaction(b)
	if err != nil {
		panic(err)
	}
	return tx
}

func (bc *BlockchainDB) GetTransactionByHash(txHash common.Hash) *types.Transaction {
	blockNumber := bc.GetBlockNumberByTxHash(txHash)
	if blockNumber == nil {
		return nil
	}
	return bc.GetTransactionByBlockNumber(blockNumber)
}

func (bc *BlockchainDB) GetBlockHashByBlockNumber(blockNumber *big.Int) common.Hash {
	b := bc.kv.MustGet(makeBlockHashByBlockNumberKey(blockNumber))
	if b == nil {
		return common.Hash{}
	}
	return common.BytesToHash(b)
}

func (bc *BlockchainDB) GetBlockNumberByBlockHash(hash common.Hash) *big.Int {
	b := bc.kv.MustGet(makeBlockNumberByBlockHashKey(hash))
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}

func (bc *BlockchainDB) GetTimestampByBlockNumber(blockNumber *big.Int) uint64 {
	n, err := codec.DecodeUint64(bc.kv.MustGet(makeTimestampByBlockNumberKey(blockNumber)))
	if err != nil {
		panic(err)
	}
	return n
}

func (bc *BlockchainDB) makeHeader(tx *types.Transaction, receipt *types.Receipt, timestamp uint64) *types.Header {
	if tx == nil {
		// genesis block hash
		blockNumber := common.Big0
		return &types.Header{
			Number:      blockNumber,
			GasLimit:    bc.GetGasLimit(),
			Time:        timestamp,
			TxHash:      types.EmptyRootHash,
			ReceiptHash: types.EmptyRootHash,
			UncleHash:   types.EmptyUncleHash,
		}
	}
	blockNumber := receipt.BlockNumber
	prevBlockNumber := new(big.Int).Sub(blockNumber, common.Big1)
	return &types.Header{
		ParentHash:  bc.GetBlockHashByBlockNumber(prevBlockNumber),
		Number:      blockNumber,
		GasLimit:    bc.GetGasLimit(),
		GasUsed:     receipt.GasUsed,
		Time:        timestamp,
		TxHash:      types.DeriveSha(types.Transactions{tx}, &fakeHasher{}),
		ReceiptHash: types.DeriveSha(types.Receipts{receipt}, &fakeHasher{}),
		Bloom:       types.CreateBloom([]*types.Receipt{receipt}),
		UncleHash:   types.EmptyUncleHash,
	}
}

func (bc *BlockchainDB) GetHeaderByBlockNumber(blockNumber *big.Int) *types.Header {
	if blockNumber.Cmp(bc.GetNumber()) > 0 {
		return nil
	}
	return bc.makeHeader(
		bc.GetTransactionByBlockNumber(blockNumber),
		bc.GetReceiptByBlockNumber(blockNumber),
		bc.GetTimestampByBlockNumber(blockNumber),
	)
}

func (bc *BlockchainDB) GetHeaderByHash(hash common.Hash) *types.Header {
	n := bc.GetBlockNumberByBlockHash(hash)
	if n == nil {
		return nil
	}
	return bc.GetHeaderByBlockNumber(n)
}

func (bc *BlockchainDB) GetBlockByHash(hash common.Hash) *types.Block {
	return bc.makeBlock(bc.GetHeaderByHash(hash))
}

func (bc *BlockchainDB) GetBlockByNumber(blockNumber *big.Int) *types.Block {
	return bc.makeBlock(bc.GetHeaderByBlockNumber(blockNumber))
}

func (bc *BlockchainDB) GetCurrentBlock() *types.Block {
	return bc.GetBlockByNumber(bc.GetNumber())
}

func (bc *BlockchainDB) makeBlock(header *types.Header) *types.Block {
	if header == nil {
		return nil
	}
	tx := bc.GetTransactionByBlockNumber(header.Number)
	if tx == nil {
		return types.NewBlock(
			header,
			[]*types.Transaction{},
			[]*types.Header{},
			[]*types.Receipt{},
			&fakeHasher{},
		)
	}
	receipt := bc.GetReceiptByBlockNumber(header.Number)
	return types.NewBlock(
		header,
		[]*types.Transaction{tx},
		[]*types.Header{},
		[]*types.Receipt{receipt},
		&fakeHasher{},
	)
}

type fakeHasher struct{}

func (d *fakeHasher) Reset() {
}

func (d *fakeHasher) Update(i1, i2 []byte) {
}

func (d *fakeHasher) Hash() common.Hash {
	return common.Hash{}
}
