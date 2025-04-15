package migrations

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
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

func migrateEVMEmulator(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	cli.DebugLog("Migrating evm/emulator...")

	oldEmulatorState := old_evm.EmulatorStateSubrealmR(oldContractState)
	newEmulatorState := evm.EmulatorStateSubrealm(newContractState)

	migrateStateDB(oldEmulatorState, newEmulatorState)
	migrateBlockchainDB(oldEmulatorState, newEmulatorState)

	cli.DebugLogf("Migrated evm/emulator")
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
	// TODO: this migration does not seem to migrate anything - always zero records found
	cli.DebugLog("Migrating iscmagic/allowance...")

	progress := NewProgressPrinter()

	oldMagicState.Iterate(old_evmimpl.PrefixAllowance, func(oldKeyBytes old_kv.Key, v []byte) bool {
		oldKeyBytes = utils.MustRemovePrefix(oldKeyBytes, old_evmimpl.PrefixAllowance)
		if len(oldKeyBytes) != 2*common.AddressLength {
			log.Panicf("unexpected key length: %v", len(oldKeyBytes))
		}

		oldFromBytes := []byte(oldKeyBytes[:common.AddressLength])
		oldToBytes := []byte(oldKeyBytes[common.AddressLength:])

		from := common.BytesToAddress(oldFromBytes)
		to := common.BytesToAddress(oldToBytes)
		newKeyBytes := evmimpl.KeyAllowance(from, to)

		if v == nil {
			newMagicState.Del(newKeyBytes)
		} else {
			oldAllowance := old_isc.MustAssetsFromBytes(v)
			newAllowance := OldAssetsToNewAssets(oldAllowance)

			newMagicState.Set(newKeyBytes, newAllowance.Bytes())
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
