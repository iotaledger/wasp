package migrations

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	old_evm_types "github.com/nnikolash/wasp-types-exported/packages/evm/evmtypes"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_rwutil "github.com/nnikolash/wasp-types-exported/packages/util/rwutil"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
)

/*
Keys to Migrate

# BlockChainDB
  - keyChainID = "c"						(X)
  - keyNumber                    = "n"		(X)
  - keyTransactionsByBlockNumber = "n:t"	(X)
  - keyReceiptsByBlockNumber     = "n:r"    (X)
  - keyBlockHeaderByBlockNumber  = "n:bh"   (X)
  - keyBlockNumberByBlockHash = "bh:n"   	(X)
  - keyBlockNumberByTxHash    = "th:n"		(X)
  - keyBlockIndexByTxHash     = "th:i"		(X)
*/

func migrateBlockchainDB(oldEmulatorStateRealm old_kv.KVStoreReader, newEmulatorStateRealm kv.KVStore) {
	oldBlockChain := old_emulator.BlockchainDBSubrealmR(oldEmulatorStateRealm)
	newBlockChain := emulator.BlockchainDBSubrealm(newEmulatorStateRealm)

	// Migrate KeyNumber
	// old / new codec Uint64 are compatible, so can be moved over as is.
	number := oldBlockChain.Get(old_emulator.KeyNumber)
	if number != nil {
		newBlockChain.Set(emulator.KeyNumber, number)
	}

	// Migrate KeyChainID
	// old / new codec Uint16 are compatible, so can be moved over as is.
	chainID := oldBlockChain.Get(old_emulator.KeyChainID)
	if chainID != nil {
		newBlockChain.Set(emulator.KeyChainID, chainID)
	}

	// Migrate KeyBlockNumberByBlockHash
	// (common.Hash:uint64) can be just copied over
	oldBlockChain.IterateSorted(old_emulator.KeyBlockNumberByBlockHash, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyBlockIndexByTxHash):])
		if len(keyWithoutPrefix) != common.HashLength {
			panic(fmt.Errorf("failed to migrate %s, invalid key length", "KeyBlockNumberByBlockHash"))
		}

		newBlockChain.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyBlockIndexByTxHash
	// (common.Hash:Uint32 map) can be just copied over
	oldBlockChain.IterateSorted(old_emulator.KeyBlockIndexByTxHash, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyBlockIndexByTxHash):])
		if len(keyWithoutPrefix) != common.HashLength {
			panic(fmt.Errorf("failed to migrate %s, invalid key length", "KeyBlockIndexByTxHash"))
		}

		newBlockChain.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyBlockNumberByTxHash
	// (common.Hash:Uint64 map) can be just copied over
	oldBlockChain.IterateSorted(old_emulator.KeyBlockNumberByTxHash, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyBlockNumberByTxHash):])
		if len(keyWithoutPrefix) != common.HashLength {
			panic(fmt.Errorf("failed to migrate %s, invalid key length", "KeyBlockNumberByTxHash"))
		}

		newBlockChain.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyTransactionsByBlockNumber
	oldBlockChain.IterateSorted(old_emulator.KeyTransactionsByBlockNumber, func(oldKey old_kv.Key, value []byte) bool {
		const blockNumberLen = 8
		oldBlockNumberBytes, oldTxIndexBytes := utils.MustSplitArrayKey(oldKey, blockNumberLen, old_emulator.KeyTransactionsByBlockNumber)
		////fmt.Printf("blockNumber: %v, txIndex: %v (INDEX:%v), key: %v (%s)\n", []byte(oldBlockNumberBytes), []byte(oldTxIndexBytes), []byte(key), string(key), len(oldTxIndexBytes) == 0)
		blockNumber := old_codec.MustDecodeUint64([]byte(oldBlockNumberBytes))
		if oldTxIndexBytes == "" {
			// It's a length record
			newBlockChain.Set(emulator.MakeTransactionsByBlockNumberKey(blockNumber), value)
			return true
		}

		txIndex := old_rwutil.NewBytesReader([]byte(oldTxIndexBytes)).Must().ReadSize32()
		newKey := collections.ArrayElemKey(string(emulator.MakeTransactionsByBlockNumberKey(blockNumber)), uint32(txIndex))

		if value == nil {
			newBlockChain.Del(newKey)
			return true
		}

		tx := lo.Must(old_evm_types.DecodeTransaction(value))
		newBlockChain.Set(newKey, evmtypes.EncodeTransaction(tx))

		return true
	})

	// Migrate KeyBlockHeaderByBlockNumber
	oldBlockChain.IterateSorted(old_emulator.KeyBlockHeaderByBlockNumber, func(key old_kv.Key, value []byte) bool {
		key = utils.MustRemovePrefix(key, old_emulator.KeyBlockHeaderByBlockNumber)
		blockNumber := old_codec.MustDecodeUint64([]byte(key))

		if value == nil {
			newBlockChain.Del(emulator.MakeBlockHeaderByBlockNumberKey(blockNumber))
			return true
		}

		oldHeader := lo.Must(old_emulator.HeaderFromBytes(value))

		newHeader := emulator.Header{
			Hash:        oldHeader.Hash,
			GasLimit:    oldHeader.GasLimit,
			GasUsed:     oldHeader.GasUsed,
			Time:        oldHeader.Time,
			TxHash:      oldHeader.TxHash,
			ReceiptHash: oldHeader.ReceiptHash,
			Bloom:       oldHeader.Bloom,
		}

		newBlockChain.Set(emulator.MakeBlockHeaderByBlockNumberKey(blockNumber), newHeader.Bytes())
		return true
	})

	// Migrate KeyReceiptsByBlockNumber
	oldBlockChain.IterateSorted(old_emulator.KeyReceiptsByBlockNumber, func(key old_kv.Key, value []byte) bool {
		const blockNumberLen = 8
		oldBlockNumberBytes, oldRecIndexBytes := utils.MustSplitArrayKey(key, blockNumberLen, old_emulator.KeyReceiptsByBlockNumber)

		blockNumber := old_codec.MustDecodeUint64([]byte(oldBlockNumberBytes))
		if oldRecIndexBytes == "" {
			newBlockChain.Set(emulator.MakeReceiptsByBlockNumberKey(blockNumber), value)
			return true
		}

		recIndex := old_rwutil.NewBytesReader([]byte(oldRecIndexBytes)).Must().ReadSize32()
		newKey := collections.ArrayElemKey(string(emulator.MakeReceiptsByBlockNumberKey(blockNumber)), uint32(recIndex))

		if value == nil {
			newBlockChain.Del(newKey)
			return true
		}

		rec := lo.Must(old_evm_types.DecodeReceipt(value))
		newBlockChain.Set(newKey, evmtypes.EncodeReceipt(rec))

		return true
	})
}
