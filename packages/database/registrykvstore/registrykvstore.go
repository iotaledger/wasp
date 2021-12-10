package registrykvstore

import "github.com/iotaledger/hive.go/kvstore"

// registrykvstore is just a wrapper to any kv store that flushes the changes to disk immediately (Sets or Dels)
// this is to prevent that the registry database is corrupted if the node is not shutdown gracefully

type RegistryKVStore struct {
	store kvstore.KVStore
}

func New(store kvstore.KVStore) kvstore.KVStore {
	return &RegistryKVStore{store}
}

func (s *RegistryKVStore) WithRealm(realm kvstore.Realm) kvstore.KVStore {
	return s.store.WithRealm(realm)
}

func (s *RegistryKVStore) Realm() kvstore.Realm {
	return s.store.Realm()
}

func (s *RegistryKVStore) Shutdown() {
	s.store.Shutdown()
}

func (s *RegistryKVStore) Iterate(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyValueConsumerFunc) error {
	return s.store.Iterate(prefix, consumerFunc)
}

func (s *RegistryKVStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc) error {
	return s.store.IterateKeys(prefix, consumerFunc)
}

func (s *RegistryKVStore) Clear() error {
	return s.store.Clear()
}

func (s *RegistryKVStore) Get(key kvstore.Key) (value kvstore.Value, err error) {
	return s.store.Get(key)
}

func (s *RegistryKVStore) Set(key kvstore.Key, value kvstore.Value) error {
	err := s.store.Set(key, value)
	if err != nil {
		return err
	}
	return s.store.Flush()
}

func (s *RegistryKVStore) Has(key kvstore.Key) (bool, error) {
	return s.store.Has(key)
}

func (s *RegistryKVStore) Delete(key kvstore.Key) error {
	err := s.store.Delete(key)
	if err != nil {
		return err
	}
	return s.store.Flush()
}

func (s *RegistryKVStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	return s.store.DeletePrefix(prefix)
}

func (s *RegistryKVStore) Batched() kvstore.BatchedMutations {
	return s.store.Batched()
}

func (s *RegistryKVStore) Flush() error {
	return s.store.Flush()
}

func (s *RegistryKVStore) Close() error {
	return s.store.Close()
}
