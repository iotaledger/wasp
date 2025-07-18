// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"bytes"
	"fmt"
	"math/big"

	"fortio.org/safecast"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/v2/packages/evm/evmutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
)

const (
	// config values:

	// EVM chain ID
	keyChainID = "c" // covered in: TestStorageContract

	// blocks:

	keyNumber                    = "n"    // covered in: TestStorageContract
	keyTransactionsByBlockNumber = "n:t"  // covered in: TestStorageContract
	keyReceiptsByBlockNumber     = "n:r"  // covered in: TestStorageContract
	keyBlockHeaderByBlockNumber  = "n:bh" // covered in: TestStorageContract

	// indexes:

	keyBlockNumberByBlockHash = "bh:n" // covered in: TestStorageContract
	keyBlockNumberByTxHash    = "th:n" // covered in: TestStorageContract
	keyBlockIndexByTxHash     = "th:i" // covered in: TestStorageContract

	BlockKeepAll = -1
)

// BlockchainDB contains logic for storing a fake blockchain (more like a list of blocks),
// intended for satisfying EVM tools that depend on the concept of a block.
type BlockchainDB struct {
	kv              kv.KVStore
	blockGasLimit   uint64
	blockKeepAmount int32
}

func NewBlockchainDB(store kv.KVStore, blockGasLimit uint64, blockKeepAmount int32) *BlockchainDB {
	return &BlockchainDB{
		kv:              BlockchainDBSubrealm(store),
		blockGasLimit:   blockGasLimit,
		blockKeepAmount: blockKeepAmount,
	}
}

func (bc *BlockchainDB) Initialized() bool {
	return bc.kv.Get(keyChainID) != nil
}

func (bc *BlockchainDB) Init(chainID uint16, timestamp uint64) {
	bc.SetChainID(chainID)
	bc.addBlock(bc.makeHeader(nil, nil, 0, timestamp))
}

func (bc *BlockchainDB) SetChainID(chainID uint16) {
	bc.kv.Set(keyChainID, codec.Encode(chainID))
}

func GetChainIDFromBlockChainDBState(kv kv.KVStoreReader) uint16 {
	return codec.MustDecode[uint16](kv.Get(keyChainID))
}

func (bc *BlockchainDB) GetChainID() uint16 {
	return GetChainIDFromBlockChainDBState(bc.kv)
}

func (bc *BlockchainDB) setNumber(n uint64) {
	bc.kv.Set(keyNumber, codec.Encode(n))
}

func (bc *BlockchainDB) GetNumber() uint64 {
	return codec.MustDecode[uint64](bc.kv.Get(keyNumber))
}

func makeTransactionsByBlockNumberKey(blockNumber uint64) kv.Key {
	return keyTransactionsByBlockNumber + kv.Key(codec.Encode(blockNumber))
}

func makeReceiptsByBlockNumberKey(blockNumber uint64) kv.Key {
	return keyReceiptsByBlockNumber + kv.Key(codec.Encode(blockNumber))
}

func makeBlockHeaderByBlockNumberKey(blockNumber uint64) kv.Key {
	return keyBlockHeaderByBlockNumber + kv.Key(codec.Encode(blockNumber))
}

func makeBlockNumberByBlockHashKey(hash common.Hash) kv.Key {
	return keyBlockNumberByBlockHash + kv.Key(hash.Bytes())
}

func makeBlockNumberByTxHashKey(hash common.Hash) kv.Key {
	return keyBlockNumberByTxHash + kv.Key(hash.Bytes())
}

func makeTxIndexInBlockByTxHashKey(hash common.Hash) kv.Key {
	return keyBlockIndexByTxHash + kv.Key(hash.Bytes())
}

func (bc *BlockchainDB) getTxArray(blockNumber uint64) *collections.Array {
	return collections.NewArray(bc.kv, string(makeTransactionsByBlockNumberKey(blockNumber)))
}

func (bc *BlockchainDB) getReceiptArray(blockNumber uint64) *collections.Array {
	return collections.NewArray(bc.kv, string(makeReceiptsByBlockNumberKey(blockNumber)))
}

func (bc *BlockchainDB) GetPendingBlockNumber() uint64 {
	return bc.GetNumber() + 1
}

func (bc *BlockchainDB) GetPendingHeader(timestamp uint64) *types.Header {
	return &types.Header{
		Difficulty: &big.Int{},
		Number:     new(big.Int).SetUint64(bc.GetPendingBlockNumber()),
		GasLimit:   bc.blockGasLimit,
		Time:       timestamp,
	}
}

func (bc *BlockchainDB) GetPendingCumulativeGasUsed() uint64 {
	blockNumber := bc.GetPendingBlockNumber()
	receiptArray := bc.getReceiptArray(blockNumber)
	n := receiptArray.Len()
	if n == 0 {
		return 0
	}
	r, err := evmtypes.DecodeReceipt(receiptArray.GetAt(n - 1))
	if err != nil {
		panic(err)
	}
	return r.CumulativeGasUsed
}

func (bc *BlockchainDB) AddTransaction(tx *types.Transaction, receipt *types.Receipt) {
	blockNumber := bc.GetPendingBlockNumber()

	txArray := bc.getTxArray(blockNumber)
	txArray.Push(evmtypes.EncodeTransaction(tx))
	bc.kv.Set(
		makeBlockNumberByTxHashKey(tx.Hash()),
		codec.Encode(blockNumber),
	)
	bc.kv.Set(
		makeTxIndexInBlockByTxHashKey(tx.Hash()),
		codec.Encode(txArray.Len()-1),
	)

	receiptArray := bc.getReceiptArray(blockNumber)
	receiptArray.Push(evmtypes.EncodeReceipt(receipt))
}

func (bc *BlockchainDB) MintBlock(timestamp uint64) {
	blockNumber := bc.GetPendingBlockNumber()
	header := bc.makeHeader(
		bc.GetTransactionsByBlockNumber(blockNumber),
		bc.GetReceiptsByBlockNumber(blockNumber),
		blockNumber,
		timestamp,
	)
	bc.addBlock(header)
	bc.prune(header.Number.Uint64())
}

func (bc *BlockchainDB) prune(currentNumber uint64) {
	if bc.blockKeepAmount <= 0 {
		// keep all blocks
		return
	}
	if currentNumber < uint64(bc.blockKeepAmount) {
		return
	}
	toDelete := currentNumber - uint64(bc.blockKeepAmount)
	// assume that all blocks prior to `toDelete` have been already deleted, so
	// we only need to delete this one.
	bc.deleteBlock(toDelete)
}

func (bc *BlockchainDB) deleteBlock(blockNumber uint64) {
	header := bc.getHeaderByBlockNumber(blockNumber)
	if header == nil {
		// already deleted?
		return
	}
	txs := bc.getTxArray(blockNumber)
	n := txs.Len()
	for i := uint32(0); i < n; i++ {
		txHash := bc.GetTransactionByBlockNumberAndIndex(blockNumber, i).Hash()
		bc.kv.Del(makeBlockNumberByTxHashKey(txHash))
		bc.kv.Del(makeTxIndexInBlockByTxHashKey(txHash))
	}
	txs.Erase()
	bc.getReceiptArray(blockNumber).Erase()
	bc.kv.Del(makeBlockHeaderByBlockNumberKey(blockNumber))
	bc.kv.Del(makeBlockNumberByBlockHashKey(header.Hash))
}

type header struct {
	Hash        common.Hash
	GasLimit    uint64 `bcs:"compact"`
	GasUsed     uint64 `bcs:"compact"`
	Time        uint64
	TxHash      common.Hash
	ReceiptHash common.Hash
	Bloom       types.Bloom
}

func makeHeader(h *types.Header) *header {
	return &header{
		Hash:        h.Hash(),
		GasLimit:    h.GasLimit,
		GasUsed:     h.GasUsed,
		Time:        h.Time,
		TxHash:      h.TxHash,
		ReceiptHash: h.ReceiptHash,
		Bloom:       h.Bloom,
	}
}

// note we do not check for excess data bytes because the old format was longer
func mustHeaderFromBytes(data []byte) (ret *header) {
	return bcs.MustUnmarshalStream[*header](bytes.NewReader(data))
}

func (h *header) Bytes() []byte {
	return bcs.MustMarshal(h)
}

func (bc *BlockchainDB) makeEthereumHeader(g *header, blockNumber uint64) *types.Header {
	if g == nil {
		return nil
	}
	var parentHash common.Hash
	if blockNumber > 0 {
		parentHash = bc.GetBlockHashByBlockNumber(blockNumber - 1)
	}
	return &types.Header{
		Difficulty:  &big.Int{},
		Number:      new(big.Int).SetUint64(blockNumber),
		GasLimit:    g.GasLimit,
		Time:        g.Time,
		ParentHash:  parentHash,
		GasUsed:     g.GasUsed,
		TxHash:      g.TxHash,
		ReceiptHash: g.ReceiptHash,
		Bloom:       g.Bloom,
		UncleHash:   types.EmptyUncleHash,
	}
}

func (bc *BlockchainDB) addBlock(header *types.Header) {
	blockNumber := header.Number.Uint64()
	bc.kv.Set(
		makeBlockHeaderByBlockNumberKey(blockNumber),
		makeHeader(header).Bytes(),
	)
	bc.kv.Set(
		makeBlockNumberByBlockHashKey(header.Hash()),
		codec.Encode(blockNumber),
	)
	bc.setNumber(blockNumber)
}

func (bc *BlockchainDB) getRawReceiptByBlockNumberAndIndex(blockNumber uint64, txIndex uint32) *types.Receipt {
	receipts := bc.getReceiptArray(blockNumber)
	if txIndex >= receipts.Len() {
		return nil
	}
	r, err := evmtypes.DecodeReceipt(receipts.GetAt(txIndex))
	if err != nil {
		panic(err)
	}
	return r
}

func (bc *BlockchainDB) getReceiptByBlockNumberAndIndex(
	blockNumber uint64,
	txIndex uint32,
	cumulativeGasUsed uint64,
	logIndex uint,
) *types.Receipt {
	r := bc.getRawReceiptByBlockNumberAndIndex(blockNumber, txIndex)
	if r == nil {
		return nil
	}
	tx := bc.GetTransactionByBlockNumberAndIndex(blockNumber, txIndex)

	r.TxHash = tx.Hash()
	r.BlockHash = bc.GetBlockHashByBlockNumber(blockNumber)
	for i, log := range r.Logs {
		log.TxHash = r.TxHash
		log.TxIndex = uint(txIndex)
		log.BlockHash = r.BlockHash
		log.BlockNumber = blockNumber
		log.Index = logIndex + safecast.MustConvert[uint](i)
	}
	if tx.To() == nil {
		from, _ := types.Sender(evmutil.Signer(big.NewInt(int64(bc.GetChainID()))), tx)
		r.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	}
	r.GasUsed = r.CumulativeGasUsed - cumulativeGasUsed
	r.BlockNumber = new(big.Int).SetUint64(blockNumber)
	r.TransactionIndex = uint(txIndex)
	r.Bloom = types.CreateBloom(r)
	return r
}

func (bc *BlockchainDB) getBlockNumberBy(key kv.Key) (uint64, bool) {
	b := bc.kv.Get(key)
	if b == nil {
		return 0, false
	}
	return codec.MustDecode[uint64](b), true
}

func (bc *BlockchainDB) GetBlockNumberByTxHash(txHash common.Hash) (uint64, bool) {
	return bc.getBlockNumberBy(makeBlockNumberByTxHashKey(txHash))
}

func (bc *BlockchainDB) GetTxIndexInBlockByTxHash(txHash common.Hash) uint32 {
	return codec.MustDecode[uint32](bc.kv.Get(makeTxIndexInBlockByTxHashKey(txHash)), 0)
}

func (bc *BlockchainDB) GetReceiptByTxHash(txHash common.Hash) *types.Receipt {
	blockNumber, ok := bc.GetBlockNumberByTxHash(txHash)
	if !ok {
		return nil
	}
	i := bc.GetTxIndexInBlockByTxHash(txHash)
	receipts := bc.GetReceiptsByBlockNumber(blockNumber)
	if int(i) >= len(receipts) {
		panic(fmt.Sprintf("cannot find evm receipt for tx %s", txHash))
	}
	return receipts[i]
}

func (bc *BlockchainDB) GetTransactionByBlockNumberAndIndex(blockNumber uint64, i uint32) *types.Transaction {
	txs := bc.getTxArray(blockNumber)
	if i >= txs.Len() {
		return nil
	}
	tx, err := evmtypes.DecodeTransaction(txs.GetAt(i))
	if err != nil {
		panic(err)
	}
	return tx
}

func (bc *BlockchainDB) GetTransactionByHash(txHash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64, err error) {
	blockNumber, ok := bc.GetBlockNumberByTxHash(txHash)
	if !ok {
		return nil, common.Hash{}, 0, 0, err
	}
	txIndex := bc.GetTxIndexInBlockByTxHash(txHash)
	block := bc.GetBlockByNumber(blockNumber)
	tx = bc.GetTransactionByBlockNumberAndIndex(blockNumber, txIndex)
	return tx, block.Hash(), blockNumber, uint64(txIndex), nil
}

func (bc *BlockchainDB) GetBlockHashByBlockNumber(blockNumber uint64) common.Hash {
	g := bc.getHeaderByBlockNumber(blockNumber)
	if g == nil {
		return common.Hash{}
	}
	return g.Hash
}

func (bc *BlockchainDB) GetBlockNumberByBlockHash(hash common.Hash) (uint64, bool) {
	return bc.getBlockNumberBy(makeBlockNumberByBlockHashKey(hash))
}

func (bc *BlockchainDB) GetTimestampByBlockNumber(blockNumber uint64) uint64 {
	g := bc.getHeaderByBlockNumber(blockNumber)
	if g == nil {
		return 0
	}
	return g.Time
}

func (bc *BlockchainDB) makeHeader(txs []*types.Transaction, receipts []*types.Receipt, blockNumber, timestamp uint64) *types.Header {
	header := &types.Header{
		Difficulty:  &big.Int{},
		Number:      new(big.Int).SetUint64(blockNumber),
		GasLimit:    bc.blockGasLimit,
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
	header.Bloom = types.MergeBloom(receipts)
	return header
}

func (bc *BlockchainDB) GetHeaderByBlockNumber(blockNumber uint64) *types.Header {
	if blockNumber > bc.GetNumber() {
		return nil
	}
	return bc.makeEthereumHeader(bc.getHeaderByBlockNumber(blockNumber), blockNumber)
}

func (bc *BlockchainDB) getHeaderByBlockNumber(blockNumber uint64) *header {
	b := bc.kv.Get(makeBlockHeaderByBlockNumberKey(blockNumber))
	if b == nil {
		return nil
	}
	return mustHeaderFromBytes(b)
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
	txs := make([]*types.Transaction, txArray.Len())
	for i := range txs {
		txs[i] = bc.GetTransactionByBlockNumberAndIndex(blockNumber, safecast.MustConvert[uint32](i))
	}
	return txs
}

func (bc *BlockchainDB) GetReceiptsByBlockNumber(blockNumber uint64) []*types.Receipt {
	txArray := bc.getTxArray(blockNumber)
	receipts := make([]*types.Receipt, txArray.Len())
	logIndex := uint(0)
	cumulativeGasUsed := uint64(0)
	for i := range receipts {
		receipts[i] = bc.getReceiptByBlockNumberAndIndex(
			blockNumber,
			safecast.MustConvert[uint32](i),
			cumulativeGasUsed,
			logIndex,
		)
		logIndex += uint(len(receipts[i].Logs))
		cumulativeGasUsed = receipts[i].CumulativeGasUsed
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
		&types.Body{
			Transactions: bc.GetTransactionsByBlockNumber(blockNumber),
		},
		bc.GetReceiptsByBlockNumber(blockNumber),
		&fakeHasher{},
	)
}

type fakeHasher struct{}

var _ types.TrieHasher = &fakeHasher{}

func (d *fakeHasher) Reset() {
}

func (d *fakeHasher) Update(i1, i2 []byte) error {
	return nil
}

func (d *fakeHasher) Hash() common.Hash {
	return common.Hash{}
}
