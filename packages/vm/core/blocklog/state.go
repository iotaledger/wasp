package blocklog

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/state"
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

func NewStateReaderFromBlockMutations(block state.Block) *StateReader {
	return NewStateReaderFromChainState(block.MutationsReader())
}

func NewStateReaderFromChainState(chainState kv.KVStoreReader) *StateReader {
	return NewStateReader(Contract.StateSubrealmR(chainState))
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
