package migrations

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	old_evm_types "github.com/nnikolash/wasp-types-exported/packages/evm/evmtypes"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

func MigrateEVMContract(oldChainState old_kv.KVStoreReader, newChainState kv.KVStore) {
	cli.DebugLog("Migrating evm contract...")

	oldContractState := old_evm.ContractPartitionR(oldChainState)
	newContractState := evm.ContractPartition(newChainState)

	migrateEVMEmulator(oldContractState, newContractState)

	oldMagicState := old_evm.ISCMagicSubrealmR(oldContractState)
	newMagicState := evm.ISCMagicSubrealm(newContractState)

	migrateISCMagicPrivileged(oldMagicState, newMagicState)
	migrateISCMagicAllowance(oldMagicState, newMagicState)
	migrateISCMagicERC20ExternalNativeTokens(oldMagicState, newMagicState)

	cli.DebugLog("Migrated evm contract")
}

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
	// Data can be just copied over
	oldBlockChain.IterateSorted(old_emulator.KeyBlockNumberByBlockHash, func(key old_kv.Key, value []byte) bool {
		newBlockChain.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyBlockIndexByTxHash
	// (common.Hash:Uint32 map) can be just copied over
	oldBlockChain.IterateSorted(old_emulator.KeyBlockIndexByTxHash, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyBlockIndexByTxHash):])
		if len(keyWithoutPrefix) != common.HashLength {
			panic("unsupported key length")
		}

		newBlockChain.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyBlockNumberByTxHash
	// (common.Hash:Uint64 map) can be just copied over
	oldBlockChain.IterateSorted(old_emulator.KeyBlockNumberByTxHash, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyBlockNumberByTxHash):])
		if len(keyWithoutPrefix) != common.HashLength {
			panic("unsupported key length")
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
		// If the len of the key (Without prefix) is just 8 bytes, it's the key containing the length of the array.
		// Port it over as is.
		if len(keyWithoutPrefix) == 8 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}
		// Otherwise the length of the key (without prefix) is 10. Consisting out of BlockNumber.ArrayIndex
		// If it is not 10, panic for now.
		if len(keyWithoutPrefix) != 10 {
			panic("unsupported receipt key length")
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
		oldHeader := old_emulator.MustHeaderFromBytes(value)

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
		// If the len of the key (Without prefix) is just 8 bytes, it's the key containing the length of the array.
		// Port it over as is.
		if len(keyWithoutPrefix) == 8 {
			newBlockChain.Set(kv.Key(key), value)
			return true
		}

		// Otherwise the length of the key (without prefix) is 10. Consisting out of BlockNumber.ArrayIndex
		// If it is not 10, panic for now.
		if len(keyWithoutPrefix) != 10 {
			panic("unsupported receipt key length")
		}

		/* Seems like the uint64 is encoded equally between old and new codec, so we can reuse the old key as is.
		newKey := []byte(emulator.MakeReceiptsByBlockNumberKey(blockNumber))
		newKey = append(newKey, keyWithoutPrefix[8:]...) // Select the last two bytes for the index
		fmt.Printf("%v\n%v", []byte(key)[:], newKey[:])
		*/

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

func migrateEVMEmulator(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	cli.DebugLog("Migrating evm/emulator...")

	oldEmulatorState := old_evm.EmulatorStateSubrealmR(oldContractState)
	newEmulatorState := evm.EmulatorStateSubrealm(newContractState)

	migrateBlockchainDB(oldEmulatorState, newEmulatorState)

	progress := NewProgressPrinter(500)

	cli.DebugLogf("Migrated %v keys for evm/emulator", progress.Count)
}

func migrateISCMagicPrivileged(oldMagicState old_kv.KVStoreReader, newMagicState kv.KVStore) {
	cli.DebugLog("Migrating iscmagic/privileged...")

	count := 0

	// Simply copy all bytes
	oldMagicState.Iterate(old_evmimpl.PrefixPrivileged, func(k old_kv.Key, v []byte) bool {
		newMagicState.Set(kv.Key(k), v)
		count++
		return true
	})

	cli.DebugLogf("Migrated %v keys for iscmagic/privileged", count)
}

func migrateISCMagicAllowance(oldMagicState old_kv.KVStoreReader, newMagicState kv.KVStore) {
	cli.DebugLog("Migrating iscmagic/allowance...")

	progress := NewProgressPrinter()

	oldMagicState.Iterate(old_evmimpl.PrefixAllowance, func(k old_kv.Key, v []byte) bool {
		k = utils.MustRemovePrefix(k, old_evmimpl.PrefixAllowance)
		if len(k) != 2*common.AddressLength {
			log.Panicf("unexpected key length: %v", len(k))
		}

		oldFromBytes := []byte(k[:common.AddressLength])
		oldToBytes := []byte(k[common.AddressLength:])

		from := common.BytesToAddress(oldFromBytes)
		to := common.BytesToAddress(oldToBytes)

		if v == nil {
			newMagicState.Del(kv.Key(k))
		} else {
			oldAllowance := old_isc.MustAssetsFromBytes(v)
			newAllowance := OldAssetsToNewAssets(oldAllowance)

			newMagicState.Set(evmimpl.KeyAllowance(from, to), newAllowance.Bytes())
		}

		progress.Print()
		return true
	})

	cli.DebugLogf("Migrated %v keys for iscmagic/allowance", progress.Count)
}

func migrateISCMagicERC20ExternalNativeTokens(oldMagicState old_kv.KVStoreReader, newMagicState kv.KVStore) {
	cli.DebugLog("Migrating iscmagic/erc20_external_native_tokens...")

	count := 0

	// Simply copying all bytes, because for now not sure what to do with it, plus according to the information about keys usage
	// this feature seems not even being used.
	// TODO: revisit this before doing actual migration.
	oldMagicState.Iterate(old_evmimpl.PrefixERC20ExternalNativeTokens, func(k old_kv.Key, v []byte) bool {
		newMagicState.Set(kv.Key(k), v)
		count++
		return true
	})

	cli.DebugLogf("Migrated %v keys for iscmagic/erc20_external_native_tokens", count)
}
