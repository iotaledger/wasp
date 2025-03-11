package migrations

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	old_evm_types "github.com/nnikolash/wasp-types-exported/packages/evm/evmtypes"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
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
	newBlockChain.Set(old_emulator.KeyNumber, number)

	// Migrate KeyChainID
	// old / new codec Uint16 are compatible, so can be moved over as is.
	chainID := oldBlockChain.Get(old_emulator.KeyChainID)
	newBlockChain.Set(emulator.KeyChainID, chainID)

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
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyTransactionsByBlockNumber):])
		// Force decoding of the number to validate
		old_codec.MustDecodeUint64(keyWithoutPrefix[:8])
		// KeyReceiptsByBlockNumber is an array.
		// If the len of the key (Without prefix) is just 8 bytes (uint64 len), it's the key containing the length of the array.
		// Port it over as is.
		if len(keyWithoutPrefix) == 8 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}

		// Otherwise the length of the key (without prefix) is 10 (uint64+uint16). Consisting out of BlockNumber.ArrayIndex
		// If it is not 10, panic for now.
		if len(keyWithoutPrefix) != 10 {
			panic(fmt.Errorf("failed to migrate %s, invalid key length", "KeyTransactionsByBlockNumber"))
		}

		// TODO: This was caught after block 10000, probably pruning in play?
		if len(value) == 0 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}

		tx, err := old_evm_types.DecodeTransaction(value)
		if err != nil {
			panic(fmt.Errorf("failed to decode transaction of old evm type: %v", err))
		}

		// We are working with the same type here, so that should work.
		newBlockChain.Set(kv.Key(key), evmtypes.EncodeTransaction(tx))

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
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyReceiptsByBlockNumber):])
		// Force decoding of the number to validate
		old_codec.MustDecodeUint64(keyWithoutPrefix[:8])

		// KeyReceiptsByBlockNumber is an array.
		// If the len of the key (Without prefix) is just 8 bytes (uint64 len), it's the key containing the length of the array.
		// Port it over as is.
		if len(keyWithoutPrefix) == 8 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}

		// Otherwise the length of the key (without prefix) is 10 (uint64+uint16). Consisting out of BlockNumber.ArrayIndex
		// If it is not 10, panic for now.
		if len(keyWithoutPrefix) != 10 {
			panic("unsupported receipt key length")
		}

		// TODO: This was caught after block 10000, probably pruning in play?
		if len(value) == 0 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}

		oldReceipt, err := old_evm_types.DecodeReceipt(value)
		if err != nil {
			panic(fmt.Errorf("failed to decode receipt from old emulator: %v", err))
		}

		newReceipt := types.Receipt{
			Type:              oldReceipt.Type,
			PostState:         oldReceipt.PostState,
			Status:            oldReceipt.Status,
			CumulativeGasUsed: oldReceipt.CumulativeGasUsed,
			Bloom:             oldReceipt.Bloom,
			Logs:              oldReceipt.Logs,
			TxHash:            oldReceipt.TxHash,
			ContractAddress:   oldReceipt.ContractAddress,
			GasUsed:           oldReceipt.GasUsed,
			EffectiveGasPrice: oldReceipt.EffectiveGasPrice,
			BlobGasUsed:       oldReceipt.BlobGasUsed,
			BlobGasPrice:      oldReceipt.BlobGasPrice,
			BlockHash:         oldReceipt.BlockHash,
			BlockNumber:       oldReceipt.BlockNumber,
			TransactionIndex:  oldReceipt.TransactionIndex,
		}

		newBlockChain.Set(kv.Key(key), evmtypes.EncodeReceipt(&newReceipt))
		return true
	})
}
