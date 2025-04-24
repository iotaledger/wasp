package utils

import (
	"fmt"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_buffered "github.com/nnikolash/wasp-types-exported/packages/kv/buffered"
	old_dict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_trietest "github.com/nnikolash/wasp-types-exported/packages/trie/test"
	"github.com/samber/lo"
)

func NewPrefixKVStore(s old_kv.KVStore, prefixesFromKey func(key old_kv.Key) [][]byte) *PrefixKVStore {
	prefixesTrieStore := old_trietest.NewInMemoryKVStore()
	prefixesTrie := lo.Must(old_trie.NewTrieUpdatable(prefixesTrieStore, old_trie.MustInitRoot(prefixesTrieStore)))

	return &PrefixKVStore{
		s:                 s,
		prefixesFromKey:   prefixesFromKey,
		prefixesTrieStore: prefixesTrieStore,
		prefixesTrie:      prefixesTrie,
		recordsByPrefix:   map[string]map[string][]byte{},
	}
}

type PrefixKVStore struct {
	s                 old_kv.KVStore
	prefixesFromKey   func(key old_kv.Key) [][]byte
	prefixesTrieStore old_trie.KVStore
	prefixesTrie      *old_trie.TrieUpdatable
	recordsByPrefix   map[string]map[string][]byte
}

var _ old_kv.KVStore = &PrefixKVStore{}
var _ old_kv.KVStoreReader = &PrefixKVStore{}

// Needs to be ran after all prefixes are registered when using non-empty store
func (s *PrefixKVStore) IndexRecords() {
	for prefix := range s.recordsByPrefix {
		s.recordsByPrefix[prefix] = map[string][]byte{}
	}

	s.s.Iterate("", func(key old_kv.Key, value []byte) bool {
		s.addToIndex(key, value)
		return true
	})
}

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

	_ = s.registerPrefix(prefix)
}

func (s *PrefixKVStore) registerPrefix(prefix string) map[string][]byte {
	prefixRecords := s.recordsByPrefix[prefix]

	if prefixRecords == nil {
		prefixRecords = map[string][]byte{}
		s.recordsByPrefix[prefix] = prefixRecords
		s.prefixesTrie.UpdateStr(prefix, prefix)
		s.prefixesTrie.Commit(s.prefixesTrieStore)
	}

	return prefixRecords
}

func (s *PrefixKVStore) Set(key old_kv.Key, value []byte) {
	s.s.Set(key, value)
	s.addToIndex(key, value)
}

func (s *PrefixKVStore) addToIndex(key old_kv.Key, value []byte) {
	if s.prefixesFromKey != nil {
		prefixesFromKey := s.prefixesFromKey(key)
		if len(prefixesFromKey) != 0 {
			for _, prefixFromKey := range prefixesFromKey {
				prefixRecords := s.registerPrefix(string(prefixFromKey))
				prefixRecords[string(key)] = value
			}
		}
	}

	matchedPrefixes := s.prefixesTrie.GetAllMatches([]byte(key))

	for _, prefix := range matchedPrefixes {
		s.recordsByPrefix[string(prefix)][string(key)] = value
	}
}

func (s *PrefixKVStore) Del(key old_kv.Key) {
	s.s.Del(key)

	if s.prefixesFromKey != nil {
		prefixesFromKey := s.prefixesFromKey(key)
		if prefixesFromKey != nil {
			for _, prefixFromKey := range prefixesFromKey {
				prefixRecords := s.recordsByPrefix[string(prefixFromKey)]
				if prefixRecords != nil {
					delete(prefixRecords, string(key))

					if len(prefixRecords) == 0 {
						delete(s.recordsByPrefix, string(prefixFromKey))
						s.prefixesTrie.DeleteStr(string(prefixFromKey))
						s.prefixesTrie.Commit(s.prefixesTrieStore)
					}
				}
			}

			// ignoring other keys for performance
			return
		}
	}

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
	if prefix == "" {
		s.s.Iterate("", f)
		return
	}

	prefixRecords, prefixKnown := s.recordsByPrefix[string(prefix)]
	if !prefixKnown {
		if s.prefixesFromKey != nil {
			prefixesFromKey := s.prefixesFromKey(prefix)
			for _, prefixFromKey := range prefixesFromKey {
				if prefixesFromKey != nil && old_kv.Key(prefixFromKey) == prefix {
					// prefix is valid, but no records
					return
				}
			}
		}

		panic(fmt.Sprintf("Prefix not registered: %v", prefix))
	}

	for key, value := range prefixRecords {
		if !f(old_kv.Key(key), value) {
			break
		}
	}
}

func (s *PrefixKVStore) IterateKeys(prefix old_kv.Key, f func(key old_kv.Key) bool) {
	if prefix == "" {
		s.s.IterateKeys("", f)
		return
	}

	prefixRecords, prefixKnown := s.recordsByPrefix[string(prefix)]
	if !prefixKnown {
		if s.prefixesFromKey != nil {
			prefixesFromKey := s.prefixesFromKey(prefix)
			for _, prefixFromKey := range prefixesFromKey {
				if prefixesFromKey != nil && old_kv.Key(prefixFromKey) == prefix {
					// prefix is valid, but no records
					return
				}
			}
		}

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

func (s *PrefixKVStore) ApplyMutations(muts *old_buffered.Mutations) (replacedValues old_dict.Dict) {
	replacedValues = make(map[old_kv.Key][]byte, len(muts.Sets)+len(muts.Dels))

	for k, v := range muts.Sets {
		replaced := s.Get(k)
		if replaced != nil {
			replacedValues[k] = replaced
		}
		s.Set(k, v)
	}
	for k := range muts.Dels {
		replaced := s.Get(k)
		if replaced != nil {
			replacedValues[k] = replaced
		}
		s.Del(k)
	}

	return replacedValues
}
