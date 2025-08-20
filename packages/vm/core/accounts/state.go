// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package accounts implements the accounts core contract which maintains ledger state for the chain
package accounts

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

type StateReader struct {
	v     isc.SchemaVersion
	state kv.KVStoreReader
}

func NewStateReader(v isc.SchemaVersion, contractState kv.KVStoreReader) *StateReader {
	return &StateReader{
		v:     v,
		state: contractState,
	}
}

func NewStateReaderFromSandbox(ctx isc.SandboxBase) *StateReader {
	return NewStateReader(ctx.SchemaVersion(), ctx.StateR())
}

func NewStateReaderFromChainState(v isc.SchemaVersion, chainState kv.KVStoreReader) *StateReader {
	return NewStateReader(v, Contract.StateSubrealmR(chainState))
}

type StateWriter struct {
	*StateReader
	state kv.KVStore
}

func NewStateWriter(v isc.SchemaVersion, contractState kv.KVStore) *StateWriter {
	return &StateWriter{
		StateReader: NewStateReader(v, contractState),
		state:       contractState,
	}
}

func NewStateWriterFromSandbox(ctx isc.Sandbox) *StateWriter {
	return NewStateWriter(ctx.SchemaVersion(), ctx.State())
}

func AgentIDFromKey(key kv.Key) (isc.AgentID, error) {
	return codec.Decode[isc.AgentID]([]byte(key))
}
