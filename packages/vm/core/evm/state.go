// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/subrealm"
)

// The evm core contract state stores two subrealms.
const (
	// keyEmulatorState is the subrealm prefix for the data stored by the emulator (StateDB + BlockchainDB)
	keyEmulatorState = "s"

	// keyISCMagic is the subrealm prefix for the ISC magic contract
	keyISCMagic = "m"
)

func ContractPartition(chainState kv.KVStore) kv.KVStore {
	return subrealm.New(chainState, kv.Key(Contract.Hname().Bytes()))
}

func ContractPartitionR(chainState kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(Contract.Hname().Bytes()))
}

func EmulatorStateSubrealm(evmPartition kv.KVStore) kv.KVStore {
	return subrealm.New(evmPartition, keyEmulatorState)
}

func EmulatorStateSubrealmR(evmPartition kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(evmPartition, keyEmulatorState)
}

func ISCMagicSubrealm(evmPartition kv.KVStore) kv.KVStore {
	return subrealm.New(evmPartition, keyISCMagic)
}

func ISCMagicSubrealmR(evmPartition kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(evmPartition, keyISCMagic)
}
