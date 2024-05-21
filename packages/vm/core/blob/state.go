package blob

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
)

type StateReader struct {
	state kv.KVStoreReader
}

func NewStateReader(contractState kv.KVStoreReader) *StateReader {
	return &StateReader{state: contractState}
}

func NewStateReaderFromSandbox(ctx isc.SandboxBase) *StateReader {
	return NewStateReader(ctx.StateR())
}

type StateWriter struct {
	*StateReader
	state kv.KVStore
}

func NewStateWriter(contractState kv.KVStore) *StateWriter {
	return &StateWriter{
		StateReader: NewStateReader(contractState),
		state:       contractState,
	}
}

func NewStateWriterFromSandbox(ctx isc.Sandbox) *StateWriter {
	return NewStateWriter(ctx.State())
}
