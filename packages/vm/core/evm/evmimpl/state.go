// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

const (
	// keyEVMState is the subrealm prefix for the EVM state, used by the emulator
	keyEVMState = "s"

	// keyISCMagic is the subrealm prefix for the ISC magic contract
	keyISCMagic = "m"
)

func evmStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyEVMState)
}

func iscMagicSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyISCMagic)
}

func iscMagicSubrealmR(state kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(state, keyISCMagic)
}
