// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

func evmStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, evm.KeyEVMState)
}

func iscMagicSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, evm.KeyISCMagic)
}

func iscMagicSubrealmR(state kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(state, evm.KeyISCMagic)
}
