package state

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/trie"
)

// state is the implementation of the State interface
type state struct {
	trieReader *trie.TrieReader
	kv.KVStoreReader
}

var _ State = &state{}

func newState(db *storeDB, root trie.VCommitment) (*state, error) {
	trie, err := db.trieReader(root)
	if err != nil {
		return nil, err
	}
	return &state{
		KVStoreReader: &trieKVAdapter{trie},
		trieReader:    trie,
	}, nil
}

func (s *state) TrieRoot() trie.VCommitment {
	return s.trieReader.Root()
}

func (s *state) GetMerkleProof(key []byte) *trie.MerkleProof {
	return s.trieReader.MerkleProof(key)
}

func (s *state) ChainID() *isc.ChainID {
	return loadChainIDFromState(s)
}

func loadChainIDFromState(s kv.KVStoreReader) *isc.ChainID {
	chid, err := isc.ChainIDFromBytes(s.MustGet(KeyChainID))
	if err != nil {
		panic(err)
	}
	return chid
}

func (s *state) BlockIndex() uint32 {
	return loadBlockIndexFromState(s)
}

func loadBlockIndexFromState(s kv.KVStoreReader) uint32 {
	return codec.MustDecodeUint32(s.MustGet(kv.Key(coreutil.StatePrefixBlockIndex)))
}

func (s *state) Timestamp() time.Time {
	ts, err := loadTimestampFromState(s)
	mustNoErr(err)
	return ts
}

func loadTimestampFromState(chainState kv.KVStoreReader) (time.Time, error) {
	tsBin, err := chainState.Get(kv.Key(coreutil.StatePrefixTimestamp))
	if err != nil {
		return time.Time{}, err
	}
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
	data, err := chainState.Get(kv.Key(coreutil.StatePrefixPrevL1Commitment))
	mustNoErr(err)
	l1c, err := L1CommitmentFromBytes(data)
	mustNoErr(err)
	return l1c
}
