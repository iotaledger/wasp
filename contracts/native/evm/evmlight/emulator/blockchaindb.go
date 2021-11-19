// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"bytes"
	"encoding/gob"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

const (
	// config values:

	// EVM chain ID
	keyChainID = "c"
	// Block gas limit
	keyGasLimit = "g"
	// Amount of blocks to keep in DB. Older blocks will be pruned every time a transaction is added
	keyKeepAmount = "k"

	// blocks:

	keyNumber                    = "n"
	keyPendingTimestamp          = "pt"
	keyTransactionsByBlockNumber = "n:t"
	keyReceiptsByBlockNumber     = "n:r"
	keyBlockHeaderByBlockNumber  = "n:bh"

	// indexes:

	keyBlockNumberByBlockHash = "bh:n"
	keyBlockNumberByTxHash    = "th:n"
	keyBlockIndexByTxHash     = "th:i"
)

// BlockchainDB contains logic for storing a fake blockchain (more like a list of blocks),
// intended for satisfying EVM tools that depend on the concept of a block.
type BlockchainDB struct {
	kv kv.KVStore
}

func NewBlockchainDB(store kv.KVStore) *BlockchainDB {
	return &BlockchainDB{kv: store}
}

func (bc *BlockchainDB) Initialized() bool {
	return bc.kv.MustGet(keyChainID) != nil
}

func (bc *BlockchainDB) Init(chainID uint16, keepAmount int32, gasLimit, timestamp uint64) {
	bc.SetChainID(chainID)
	bc.SetGasLimit(gasLimit)
	bc.SetKeepAmount(keepAmount)
	bc.addBlock(bc.makeHeader(nil, nil, 0, timestamp), timestamp+1)
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

func (bc *BlockchainDB) SetKeepAmount(keepAmount int32) {
	bc.kv.Set(keyKeepAmount, codec.EncodeInt32(keepAmount))
}

func (bc *BlockchainDB) keepAmount() int32 {
	gas, err := codec.DecodeInt32(bc.kv.MustGet(keyKeepAmount), -1)
	if err != nil {
		panic(err)
	}
	return gas
}

func (bc *BlockchainDB) setPendingTimestamp(timestamp uint64) {
	bc.kv.Set(keyPendingTimestamp, codec.EncodeUint64(timestamp))
}

func (bc *BlockchainDB) getPendingTimestamp() uint64 {
	timestamp, err := codec.DecodeUint64(bc.kv.MustGet(keyPendingTimestamp))
	if err != nil {
		panic(err)
	}
	return timestamp
}

func (bc *BlockchainDB) setNumber(n uint64) {
	bc.kv.Set(keyNumber, codec.EncodeUint64(n))
}

func (bc *BlockchainDB) GetNumber() uint64 {
	n, err := codec.DecodeUint64(bc.kv.MustGet(keyNumber))
	if err != nil {
		panic(err)
	}
	return n
}

func makeTransactionsByBlockNumberKey(blockNumber uint64) kv.Key {
	return keyTransactionsByBlockNumber + kv.Key(codec.EncodeUint64(blockNumber))
}

func makeReceiptsByBlockNumberKey(blockNumber uint64) kv.Key {
	return keyReceiptsByBlockNumber + kv.Key(codec.EncodeUint64(blockNumber))
}

func makeBlockHeaderByBlockNumberKey(blockNumber uint64) kv.Key {
	return keyBlockHeaderByBlockNumber + kv.Key(codec.EncodeUint64(blockNumber))
}

func makeBlockNumberByBlockHashKey(hash common.Hash) kv.Key {
	return keyBlockNumberByBlockHash + kv.Key(hash.Bytes())
}

func makeBlockNumberByTxHashKey(hash common.Hash) kv.Key {
	return keyBlockNumberByTxHash + kv.Key(hash.Bytes())
}

func makeBlockIndexByTxHashKey(hash common.Hash) kv.Key {
	return keyBlockIndexByTxHash + kv.Key(hash.Bytes())
}

func (bc *BlockchainDB) getTxArray(blockNumber uint64) *collections.Array32 {
	return collections.NewArray32(bc.kv, string(makeTransactionsByBlockNumberKey(blockNumber)))
}

func (bc *BlockchainDB) getReceiptArray(blockNumber uint64) *collections.Array32 {
	return collections.NewArray32(bc.kv, string(makeReceiptsByBlockNumberKey(blockNumber)))
}

func (bc *BlockchainDB) GetPendingBlockNumber() uint64 {
	return bc.GetNumber() + 1
}

func (bc *BlockchainDB) GetPendingHeader() *types.Header {
	return &types.Header{
		Difficulty: &big.Int{},
		Number:     new(big.Int).SetUint64(bc.GetPendingBlockNumber()),
		GasLimit:   bc.GetGasLimit(),
		Time:       bc.getPendingTimestamp(),
	}
}

func (bc *BlockchainDB) GetLatestPendingReceipt() *types.Receipt {
	blockNumber := bc.GetPendingBlockNumber()
	receiptArray := bc.getReceiptArray(blockNumber)
	n, err := receiptArray.Len()
	if err != nil {
		panic(err)
	}
	if n == 0 {
		return nil
	}
	return bc.GetReceiptByBlockNumberAndIndex(blockNumber, n-1)
}

func (bc *BlockchainDB) AddTransaction(tx *types.Transaction, receipt *types.Receipt) {
	blockNumber := bc.GetPendingBlockNumber()

	txArray := bc.getTxArray(blockNumber)
	txArray.MustPush(evmtypes.EncodeTransaction(tx))
	bc.kv.Set(
		makeBlockNumberByTxHashKey(tx.Hash()),
		codec.EncodeUint64(blockNumber),
	)
	bc.kv.Set(
		makeBlockIndexByTxHashKey(tx.Hash()),
		codec.EncodeUint32(txArray.MustLen()-1),
	)

	receiptArray := bc.getReceiptArray(blockNumber)
	receiptArray.MustPush(evmtypes.EncodeReceipt(receipt))
}

func (bc *BlockchainDB) MintBlock(timestamp uint64) {
	blockNumber := bc.GetPendingBlockNumber()
	header := bc.makeHeader(
		bc.GetTransactionsByBlockNumber(blockNumber),
		bc.GetReceiptsByBlockNumber(blockNumber),
		blockNumber,
		bc.getPendingTimestamp(),
	)
	bc.addBlock(header, timestamp)
	bc.prune(header.Number.Uint64())
}

func (bc *BlockchainDB) prune(currentNumber uint64) {
	keepAmount := bc.keepAmount()
	if keepAmount < 0 {
		// keep all blocks
		return
	}
	if currentNumber <= uint64(keepAmount) {
		return
	}
	toDelete := currentNumber - uint64(keepAmount)
	// assume that all blocks prior to `toDelete` have been already deleted, so
	// we only need to delete this one.
	bc.deleteBlock(toDelete)
}

func (bc *BlockchainDB) deleteBlock(blockNumber uint64) {
	header := bc.getHeaderGobByBlockNumber(blockNumber)
	if header == nil {
		// already deleted?
		return
	}
	txs := bc.getTxArray(blockNumber)
	n := txs.MustLen()
	for i := uint32(0); i < n; i++ {
		txHash := bc.GetTransactionByBlockNumberAndIndex(blockNumber, i).Hash()
		bc.kv.Del(makeBlockNumberByTxHashKey(txHash))
		bc.kv.Del(makeBlockIndexByTxHashKey(txHash))
	}
	txs.MustErase()
	bc.getReceiptArray(blockNumber).MustErase()
	bc.kv.Del(makeBlockHeaderByBlockNumberKey(blockNumber))
	bc.kv.Del(makeBlockNumberByBlockHashKey(header.Hash))
}

type headerGob struct {
	Hash        common.Hash
	GasUsed     uint64
	Time        uint64
	TxHash      common.Hash
	ReceiptHash common.Hash
	Bloom       types.Bloom
}

func makeHeaderGob(header *types.Header) *headerGob {
	return &headerGob{
		Hash:        header.Hash(),
		GasUsed:     header.GasUsed,
		Time:        header.Time,
		TxHash:      header.TxHash,
		ReceiptHash: header.ReceiptHash,
		Bloom:       header.Bloom,
	}
}

func encodeHeaderGob(g *headerGob) []byte {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(g)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (bc *BlockchainDB) decodeHeaderGob(b []byte) *headerGob {
	var g headerGob
	err := gob.NewDecoder(bytes.NewReader(b)).Decode(&g)
	if err != nil {
		panic(err)
	}
	return &g
}

func (bc *BlockchainDB) headerFromGob(g *headerGob, blockNumber uint64) *types.Header {
	var parentHash common.Hash
	if blockNumber > 0 {
		parentHash = bc.GetBlockHashByBlockNumber(blockNumber - 1)
	}
	return &types.Header{
		Difficulty:  &big.Int{},
		Number:      new(big.Int).SetUint64(blockNumber),
		GasLimit:    bc.GetGasLimit(),
		Time:        g.Time,
		ParentHash:  parentHash,
		GasUsed:     g.GasUsed,
		TxHash:      g.TxHash,
		ReceiptHash: g.ReceiptHash,
		Bloom:       g.Bloom,
		UncleHash:   types.EmptyUncleHash,
	}
}

func (bc *BlockchainDB) addBlock(header *types.Header, pendingTimestamp uint64) {
	blockNumber := header.Number.Uint64()
	bc.kv.Set(
		makeBlockHeaderByBlockNumberKey(blockNumber),
		encodeHeaderGob(makeHeaderGob(header)),
	)
	bc.kv.Set(
		makeBlockNumberByBlockHashKey(header.Hash()),
		codec.EncodeUint64(blockNumber),
	)
	bc.setNumber(blockNumber)
	bc.setPendingTimestamp(pendingTimestamp)
}

func (bc *BlockchainDB) GetReceiptByBlockNumberAndIndex(blockNumber uint64, i uint32) *types.Receipt {
	receipts := bc.getReceiptArray(blockNumber)
	if i >= receipts.MustLen() {
		return nil
	}
	r, err := evmtypes.DecodeReceipt(receipts.MustGetAt(i))
	if err != nil {
		panic(err)
	}
	tx := bc.GetTransactionByBlockNumberAndIndex(blockNumber, i)
	r.TxHash = tx.Hash()
	if tx.To() == nil {
		from, _ := types.Sender(evmtypes.Signer(big.NewInt(int64(bc.GetChainID()))), tx)
		r.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	}
	r.GasUsed = r.CumulativeGasUsed
	if i > 0 {
		prev, err := evmtypes.DecodeReceipt(receipts.MustGetAt(i - 1))
		if err != nil {
			panic(err)
		}
		r.GasUsed -= prev.CumulativeGasUsed
	}
	r.BlockHash = bc.GetBlockHashByBlockNumber(blockNumber)
	r.BlockNumber = new(big.Int).SetUint64(blockNumber)
	return r
}

func (bc *BlockchainDB) getBlockNumberBy(key kv.Key) (uint64, bool) {
	b := bc.kv.MustGet(key)
	if b == nil {
		return 0, false
	}
	n, err := codec.DecodeUint64(b)
	if err != nil {
		panic(err)
	}
	return n, true
}

func (bc *BlockchainDB) GetBlockNumberByTxHash(txHash common.Hash) (uint64, bool) {
	return bc.getBlockNumberBy(makeBlockNumberByTxHashKey(txHash))
}

func (bc *BlockchainDB) GetBlockIndexByTxHash(txHash common.Hash) uint32 {
	n, err := codec.DecodeUint32(bc.kv.MustGet(makeBlockIndexByTxHashKey(txHash)), 0)
	if err != nil {
		panic(err)
	}
	return n
}

func (bc *BlockchainDB) GetReceiptByTxHash(txHash common.Hash) *types.Receipt {
	blockNumber, ok := bc.GetBlockNumberByTxHash(txHash)
	if !ok {
		return nil
	}
	i := bc.GetBlockIndexByTxHash(txHash)
	return bc.GetReceiptByBlockNumberAndIndex(blockNumber, i)
}

func (bc *BlockchainDB) GetTransactionByBlockNumberAndIndex(blockNumber uint64, i uint32) *types.Transaction {
	txs := bc.getTxArray(blockNumber)
	if i >= txs.MustLen() {
		return nil
	}
	tx, err := evmtypes.DecodeTransaction(txs.MustGetAt(i))
	if err != nil {
		panic(err)
	}
	return tx
}

func (bc *BlockchainDB) GetTransactionByHash(txHash common.Hash) *types.Transaction {
	blockNumber, ok := bc.GetBlockNumberByTxHash(txHash)
	if !ok {
		return nil
	}
	i := bc.GetBlockIndexByTxHash(txHash)
	return bc.GetTransactionByBlockNumberAndIndex(blockNumber, i)
}

func (bc *BlockchainDB) GetBlockHashByBlockNumber(blockNumber uint64) common.Hash {
	g := bc.getHeaderGobByBlockNumber(blockNumber)
	if g == nil {
		return common.Hash{}
	}
	return g.Hash
}

func (bc *BlockchainDB) GetBlockNumberByBlockHash(hash common.Hash) (uint64, bool) {
	return bc.getBlockNumberBy(makeBlockNumberByBlockHashKey(hash))
}

func (bc *BlockchainDB) GetTimestampByBlockNumber(blockNumber uint64) uint64 {
	g := bc.getHeaderGobByBlockNumber(blockNumber)
	if g == nil {
		return 0
	}
	return g.Time
}

func (bc *BlockchainDB) makeHeader(txs []*types.Transaction, receipts []*types.Receipt, blockNumber, timestamp uint64) *types.Header {
	header := &types.Header{
		Difficulty:  &big.Int{},
		Number:      new(big.Int).SetUint64(blockNumber),
		GasLimit:    bc.GetGasLimit(),
		Time:        timestamp,
		TxHash:      types.EmptyRootHash,
		ReceiptHash: types.EmptyRootHash,
		UncleHash:   types.EmptyUncleHash,
	}
	if blockNumber == 0 {
		// genesis block hash
		return header
	}
	prevBlockNumber := blockNumber - 1
	gasUsed := uint64(0)
	if len(receipts) > 0 {
		gasUsed = receipts[len(receipts)-1].CumulativeGasUsed
	}
	header.ParentHash = bc.GetBlockHashByBlockNumber(prevBlockNumber)
	header.GasUsed = gasUsed
	if len(txs) > 0 {
		header.TxHash = types.DeriveSha(types.Transactions(txs), &fakeHasher{})
		header.ReceiptHash = types.DeriveSha(types.Receipts(receipts), &fakeHasher{})
	}
	header.Bloom = types.CreateBloom(receipts)
	return header
}

func (bc *BlockchainDB) GetHeaderByBlockNumber(blockNumber uint64) *types.Header {
	if blockNumber > bc.GetNumber() {
		return nil
	}
	return bc.headerFromGob(bc.getHeaderGobByBlockNumber(blockNumber), blockNumber)
}

func (bc *BlockchainDB) getHeaderGobByBlockNumber(blockNumber uint64) *headerGob {
	b := bc.kv.MustGet(makeBlockHeaderByBlockNumberKey(blockNumber))
	if b == nil {
		return nil
	}
	return bc.decodeHeaderGob(b)
}

func (bc *BlockchainDB) GetHeaderByHash(hash common.Hash) *types.Header {
	n, ok := bc.GetBlockNumberByBlockHash(hash)
	if !ok {
		return nil
	}
	return bc.GetHeaderByBlockNumber(n)
}

func (bc *BlockchainDB) GetBlockByHash(hash common.Hash) *types.Block {
	return bc.makeBlock(bc.GetHeaderByHash(hash))
}

func (bc *BlockchainDB) GetBlockByNumber(blockNumber uint64) *types.Block {
	return bc.makeBlock(bc.GetHeaderByBlockNumber(blockNumber))
}

func (bc *BlockchainDB) GetCurrentBlock() *types.Block {
	return bc.GetBlockByNumber(bc.GetNumber())
}

func (bc *BlockchainDB) GetTransactionsByBlockNumber(blockNumber uint64) []*types.Transaction {
	txArray := bc.getTxArray(blockNumber)
	n := txArray.MustLen()
	txs := make([]*types.Transaction, n)
	for i := uint32(0); i < n; i++ {
		txs[i] = bc.GetTransactionByBlockNumberAndIndex(blockNumber, i)
	}
	return txs
}

func (bc *BlockchainDB) GetReceiptsByBlockNumber(blockNumber uint64) []*types.Receipt {
	txArray := bc.getTxArray(blockNumber)
	n := txArray.MustLen()
	receipts := make([]*types.Receipt, n)
	for i := uint32(0); i < n; i++ {
		receipts[i] = bc.GetReceiptByBlockNumberAndIndex(blockNumber, i)
	}
	return receipts
}

func (bc *BlockchainDB) makeBlock(header *types.Header) *types.Block {
	if header == nil {
		return nil
	}
	blockNumber := header.Number.Uint64()
	return types.NewBlock(
		header,
		bc.GetTransactionsByBlockNumber(blockNumber),
		[]*types.Header{},
		bc.GetReceiptsByBlockNumber(blockNumber),
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
