package main

import (
	"fmt"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_trietest "github.com/nnikolash/wasp-types-exported/packages/trie/test"
	"github.com/samber/lo"
)

func NewPrefixKVStore(s old_kv.KVStore) *PrefixKVStore {
	prefixesTrieStore := old_trietest.NewInMemoryKVStore()
	prefixesTrie := lo.Must(old_trie.NewTrieUpdatable(prefixesTrieStore, old_trie.MustInitRoot(prefixesTrieStore)))

	return &PrefixKVStore{
		s:                 s,
		prefixesTrieStore: prefixesTrieStore,
		prefixesTrie:      prefixesTrie,
		recordsByPrefix:   map[string]map[string][]byte{},
	}
}

type PrefixKVStore struct {
	s                 old_kv.KVStore
	prefixesTrieStore old_trie.KVStore
	prefixesTrie      *old_trie.TrieUpdatable
	recordsByPrefix   map[string]map[string][]byte
}

var _ old_kv.KVStore = &PrefixKVStore{}
var _ old_kv.KVStoreReader = &PrefixKVStore{}

func (s *PrefixKVStore) RegisterPrefix(prefix string, subrealms ...any) {
	// In theory, prefixes could be detected automatically after calls Iterate*().
	// But it is unnecesary complication and is less reliable because of the order of execution.
	// Lets not overcomplicate things.
	// Also, if entries added between multiple calls to RegisterPrefix(), new prefixes does not include old entries.
	// But we dont care - just call RegisterPrefix at the beginning.

	if len(subrealms) > 0 {
		prefixWithSubrealms := ""

		for _, subrealm := range subrealms {
			switch subrealm := subrealm.(type) {
			case string:
				prefixWithSubrealms += subrealm
			case old_isc.Hname:
				prefixWithSubrealms += string(subrealm.Bytes())
			default:
				panic(fmt.Sprintf("Unknown subrealm type: %T", subrealm))
			}
		}

		prefix = prefixWithSubrealms + prefix
	}

	if s.recordsByPrefix[prefix] == nil {
		s.prefixesTrie.UpdateStr(prefix, prefix)
		s.prefixesTrie.Commit(s.prefixesTrieStore)

		s.recordsByPrefix[prefix] = map[string][]byte{}
	}
}

func (s *PrefixKVStore) Set(key old_kv.Key, value []byte) {
	s.s.Set(key, value)

	matchedPrefixes := s.prefixesTrie.GetAllMatches([]byte(key))

	for _, prefix := range matchedPrefixes {
		s.recordsByPrefix[string(prefix)][string(key)] = value
	}
}

func (s *PrefixKVStore) Del(key old_kv.Key) {
	s.s.Del(key)

	matchedPrefixes := s.prefixesTrie.GetAllMatches([]byte(key))
	for _, prefix := range matchedPrefixes {
		delete(s.recordsByPrefix[string(prefix)], string(key))
	}
}

func (s *PrefixKVStore) Get(key old_kv.Key) []byte {
	return s.s.Get(key)
}

func (s *PrefixKVStore) Has(key old_kv.Key) bool {
	return s.s.Has(key)
}

func (s *PrefixKVStore) Iterate(prefix old_kv.Key, f func(key old_kv.Key, value []byte) bool) {
	prefixRecords, prefixKnown := s.recordsByPrefix[string(prefix)]
	if !prefixKnown {
		panic(fmt.Sprintf("Prefix not registered: %v", prefix))
	}

	fmt.Println("XXX", prefixRecords)

	for key, value := range prefixRecords {
		if !f(old_kv.Key(key), value) {
			break
		}
	}
}

func (s *PrefixKVStore) IterateKeys(prefix old_kv.Key, f func(key old_kv.Key) bool) {
	prefixRecords, prefixKnown := s.recordsByPrefix[string(prefix)]
	if !prefixKnown {
		panic(fmt.Sprintf("Prefix not registered: %v", prefix))
	}

	for key := range prefixRecords {
		if !f(old_kv.Key(key)) {
			break
		}
	}
}

func (s *PrefixKVStore) IterateSorted(prefix old_kv.Key, f func(key old_kv.Key, value []byte) bool) {
	s.s.IterateSorted(prefix, f)
}

func (s *PrefixKVStore) IterateKeysSorted(prefix old_kv.Key, f func(key old_kv.Key) bool) {
	s.s.IterateKeysSorted(prefix, f)
}
