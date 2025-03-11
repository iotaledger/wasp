package migrations

import (
	"github.com/ethereum/go-ethereum/common"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_emulator "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/emulator"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
)

/*
Keys to Migrate
	* keyAccountNonce          = "n"
	* keyAccountCode           = "c"
	* keyAccountState          = "s"
	* keyAccountSelfDestructed = "S"
*/

func migrateStateDB(oldEmulatorStateRealm old_kv.KVStoreReader, newEmulatorStateRealm kv.KVStore) {
	oldStateDB := old_emulator.StateDBSubrealmR(oldEmulatorStateRealm)
	newStateDB := emulator.StateDBSubrealm(newEmulatorStateRealm)

	// Migrate KeyAccountNonce
	// Map(common.Address:uint64) can be just moved over
	oldStateDB.IterateSorted(old_emulator.KeyAccountNonce, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyAccountNonce):])
		if len(keyWithoutPrefix) != common.AddressLength {
			panic("unsupported key length")
		}

		newStateDB.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyAccountNonce
	// Map(common.Address:bytes) can be just moved over
	oldStateDB.IterateSorted(old_emulator.KeyAccountCode, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyAccountCode):])
		if len(keyWithoutPrefix) != common.AddressLength {
			panic("unsupported key length")
		}

		newStateDB.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyAccountState
	// Map(common.Address+common.Hash:common.Hash) can be just moved over
	oldStateDB.IterateSorted(old_emulator.KeyAccountState, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyAccountState):])

		if len(keyWithoutPrefix) != common.AddressLength+common.HashLength {
			panic("unsupported key length")
		}

		newStateDB.Set(kv.Key(key), value)
		return true
	})

	// Migrate KeyAccountSelfDestructed
	// Map(common.Address:bool) can be just moved over
	oldStateDB.IterateSorted(old_emulator.KeyAccountSelfDestructed, func(key old_kv.Key, value []byte) bool {
		keyWithoutPrefix := []byte(key[len(old_emulator.KeyAccountSelfDestructed):])

		if len(keyWithoutPrefix) != common.AddressLength {
			panic("unsupported key length")
		}

		newStateDB.Set(kv.Key(key), value)
		return true
	})

}
