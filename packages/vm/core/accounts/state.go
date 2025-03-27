// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
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

// converts an account key from the accounts contract (shortform without chainID) to an AgentID
func AgentIDFromKey(key kv.Key) (isc.AgentID, error) {
	return codec.Decode[isc.AgentID]([]byte(key))
}
