package migrations

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	oldBlockChain.IterateSorted(old_emulator.KeyTransactionsByBlockNumber, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := utils.MustRemovePrefix(key, old_emulator.KeyTransactionsByBlockNumber)
		const blockNumberLen = 8
		const sepLen = 1
		const txIndexLen = 4
		const expectedTotalLen = blockNumberLen + sepLen + txIndexLen
		if len(keyWithoutPrefix) != expectedTotalLen {
			return true
		}

		if keyWithoutPrefix[blockNumberLen+sepLen] != '#' {
			panic(fmt.Sprintf("unexpected key format: %x / %v", key, string(key)))
		}

		oldBlockNumberBytes := keyWithoutPrefix[:blockNumberLen]
		txIndexBytes := keyWithoutPrefix[blockNumberLen+sepLen:]

		var blockNumber uint64
		if len(oldBlockNumberBytes) == 8 {
			blockNumber = old_codec.MustDecodeUint64([]byte(oldBlockNumberBytes))
		} else if len(oldBlockNumberBytes) < 8 {
			// NOTE: There is a bug in wasp - for old blocks big.Int.Bytes() was used to encode block number,
			// and database was not migrated after implementation changed.
			// Example: key 6e3a740023000000000000 at block 8960

			// TODO:
			// Revisit this. For now, just skipping these records, because I value is also invalid - just one byte 0x09.
			// Maybe they were deleted upon migration in such way?

			return true
			// blockNumber = big.NewInt(0).SetBytes([]byte(oldBlockNumberBytes)).Uint64()
		} else {
			panic(fmt.Sprintf("invalid key length: %v: %x, %x", len(oldBlockNumberBytes), oldBlockNumberBytes, key))
		}

		txIndex := old_rwutil.NewBytesReader([]byte(txIndexBytes)).Must().ReadUint32()
		newKey := collections.ArrayElemKey(string(emulator.MakeTransactionsByBlockNumberKey(blockNumber)), txIndex)

		// TODO: This was caught after block 10000, probably pruning in play?
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
		blockNumber := old_codec.MustDecodeUint64([]byte(key[len(old_emulator.KeyBlockHeaderByBlockNumber):]))

		// TODO: This was caught after block 10000, probably pruning in play?
		if len(value) == 0 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}

		oldHeader, err := old_emulator.HeaderFromBytes(value)
		if err != nil {
			panic(fmt.Errorf("failed to decode header of old evm type: %v", err))
		}

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
		keyWithoutPrefix := utils.MustRemovePrefix(key, old_emulator.KeyReceiptsByBlockNumber)

		const blockNumberLen = 8
		const sepLen = 1
		const recIndexLen = 4
		const expectedTotalLen = blockNumberLen + sepLen + recIndexLen
		if len(keyWithoutPrefix) != expectedTotalLen {
			return true
		}

		if keyWithoutPrefix[blockNumberLen+sepLen] != '#' {
			panic(fmt.Sprintf("unexpected key format: %x / %v", key, string(key)))
		}

		oldBlockNumberBytes := keyWithoutPrefix[:blockNumberLen]
		recIndexBytes := keyWithoutPrefix[blockNumberLen+sepLen:]

		blockNumber := old_codec.MustDecodeUint64([]byte(oldBlockNumberBytes))
		recIndex := old_rwutil.NewBytesReader([]byte(recIndexBytes)).Must().ReadUint32()
		newKey := collections.ArrayElemKey(string(emulator.MakeTransactionsByBlockNumberKey(blockNumber)), recIndex)

		// TODO: This was caught after block 10000, probably pruning in play?
		if value == nil {
			newBlockChain.Del(newKey)
			return true
		}

		oldReceipt := lo.Must(old_evm_types.DecodeReceipt(value))
		newReceipt := types.Receipt(*oldReceipt)

		newBlockChain.Set(newKey, evmtypes.EncodeReceipt(&newReceipt))

		return true
	})
}
