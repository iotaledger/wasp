package state

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// state is the implementation of the State interface
type state struct {
	trieReader *trie.TrieReader
	kv.KVStoreReader
}

var _ State = &state{}

func newState(db *storeDB, root trie.Hash) (*state, error) {
	trie, err := db.trieReader(root)
	if err != nil {
		return nil, err
	}
	return &state{
		KVStoreReader: kv.NewCachedKVStoreReader(&trieKVAdapter{trie}),
		trieReader:    trie,
	}, nil
}

func (s *state) TrieRoot() trie.Hash {
	return s.trieReader.Root()
}

func (s *state) GetMerkleProof(key []byte) *trie.MerkleProof {
	return s.trieReader.MerkleProof(key)
}

func (s *state) BlockIndex() uint32 {
	return loadBlockIndexFromState(s)
}

func loadBlockIndexFromState(s kv.KVStoreReader) uint32 {
	return codec.MustDecodeUint32(s.Get(kv.Key(coreutil.StatePrefixBlockIndex)))
}

func (s *state) Timestamp() time.Time {
	ts, err := loadTimestampFromState(s)
	mustNoErr(err)
	return ts
}

func loadTimestampFromState(chainState kv.KVStoreReader) (time.Time, error) {
	tsBin := chainState.Get(kv.Key(coreutil.StatePrefixTimestamp))
	ts, err := codec.DecodeTime(tsBin)
	if err != nil {
		return time.Time{}, fmt.Errorf("loadTimestampFromState: %w", err)
	}
	return ts, nil
}

func (s *state) PreviousL1Commitment() *L1Commitment {
	return loadPrevL1CommitmentFromState(s)
}

func loadPrevL1CommitmentFromState(chainState kv.KVStoreReader) *L1Commitment {
	data := chainState.Get(kv.Key(coreutil.StatePrefixPrevL1Commitment))
	l1c, err := L1CommitmentFromBytes(data)
	mustNoErr(err)
	return l1c
}

func (s *state) SchemaVersion() isc.SchemaVersion {
	return root.NewStateAccess(s).SchemaVersion()
}

func (s *state) String() string {
	return fmt.Sprintf("State[si#%v]%v", s.BlockIndex(), s.TrieRoot())
}

func (s *state) Equals(other State) bool {
	if !s.TrieRoot().Equals(other.TrieRoot()) ||
		s.BlockIndex() != other.BlockIndex() ||
		s.Timestamp() != other.Timestamp() ||
		!s.PreviousL1Commitment().Equals(other.PreviousL1Commitment()) {
		return false
	}
	commonState := getCommonState(s, other)
	for _, entry := range commonState {
		if !bytes.Equal(entry.value1, entry.value2) {
			return false
		}
	}
	return true
}

type commonEntry struct {
	value1 []byte
	value2 []byte
}

func getCommonState(state1, state2 State) map[kv.Key]*commonEntry {
	result := make(map[kv.Key]*commonEntry)
	iterateFun := func(iterState State, setValueFun func(*commonEntry, []byte)) {
		iterState.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
			entry, ok := result[key]
			if !ok {
				entry = &commonEntry{}
				result[key] = entry
			}
			setValueFun(entry, value)
			return true
		})
	}
	iterateFun(state1, func(entry *commonEntry, value []byte) { entry.value1 = value })
	iterateFun(state2, func(entry *commonEntry, value []byte) { entry.value2 = value })
	return result
}
