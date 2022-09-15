package textdb

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/mr-tron/base58"

	"github.com/iotaledger/hive.go/core/byteutils"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/hive.go/core/types"
)

const storePerm = 0o664

// a key/value store implementation that uses text files
type textKV struct {
	sync.RWMutex
	marshaller
	filename      string
	log           *logger.Logger
	realm         []byte
	inMemoryStore kvstore.KVStore
}

type marshaller interface {
	marshal(val interface{}) ([]byte, error)
	unmarshal(buf []byte, v interface{}) error
}

func getMarshaller(filename string) marshaller {
	if filepath.Ext(filename) == "yaml" {
		return &yamlMarshaller{}
	}
	return &jsonMarshaller{}
}

// a key/value store for text storage. Works with both yaml and json.
func NewTextKV(log *logger.Logger, filename string) kvstore.KVStore {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, storePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fd, err := f.Stat()
	if err != nil {
		panic(err)
	}
	if fd.Size() == 0 {
		err = os.WriteFile(f.Name(), []byte("{}"), storePerm)
		if err != nil {
			panic(err)
		}
	}
	tKV := &textKV{
		filename:      f.Name(),
		log:           log,
		marshaller:    getMarshaller(filename),
		inMemoryStore: mapdb.NewMapDB(),
	}
	data, err := tKV.load()
	if err != nil {
		panic(err)
	}
	// load data into inMemoryStore
	for key, value := range data {
		keyB, err := base58.Decode(key)
		if err != nil {
			panic(err)
		}
		valB, err := tKV.marshal(value)
		if err != nil {
			panic(err)
		}
		err = tKV.inMemoryStore.Set(keyB, valB)
		if err != nil {
			panic(err)
		}
	}
	return tKV
}

// WithRealm is a factory method for using the same underlying storage with a different realm.
func (s *textKV) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return &textKV{
		filename:      s.filename,
		log:           s.log,
		realm:         realm,
		inMemoryStore: s.inMemoryStore,
	}, nil
}

// Realm returns the configured realm.
func (s *textKV) Realm() kvstore.Realm {
	return byteutils.ConcatBytes(s.realm)
}

// Shutdown marks the store as shutdown.
func (s *textKV) Shutdown() {
}

func (s *textKV) load() (map[string]interface{}, error) {
	data, err := os.ReadFile(s.filename)
	if err != nil {
		return nil, err
	}
	ret := map[string]interface{}{}
	err = s.unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
func (s *textKV) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {
	s.RLock()
	defer s.RUnlock()
	return s.inMemoryStore.Iterate(byteutils.ConcatBytes(s.realm, prefix), kvConsumerFunc, direction...)
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
func (s *textKV) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {
	s.RLock()
	defer s.RUnlock()
	return s.inMemoryStore.IterateKeys(byteutils.ConcatBytes(s.realm, prefix), consumerFunc, direction...)
}

// clear the key/value store
func (s *textKV) Clear() error {
	s.Lock()
	defer s.Unlock()
	err := s.inMemoryStore.Clear()
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, []byte("{}"), storePerm)
}

// Get gets the given key or nil if it doesn't exist or an error if an error occurred.
func (s *textKV) Get(key kvstore.Key) (value kvstore.Value, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.inMemoryStore.Get(byteutils.ConcatBytes(s.realm, key))
}

// Set sets the given key and value.
func (s *textKV) Set(key kvstore.Key, value kvstore.Value) error {
	s.Lock()
	defer s.Unlock()

	// first update in inMemoryStore
	err := s.inMemoryStore.Set(byteutils.ConcatBytes(s.realm, key), value)
	if err != nil {
		return err
	}

	return s.flush()
}

// Has checks whether the given key exists.
func (s *textKV) Has(key kvstore.Key) (bool, error) {
	s.RLock()
	defer s.RUnlock()
	return s.inMemoryStore.Has(byteutils.ConcatBytes(s.realm, key))
}

// Delete deletes the entry for the given key.
func (s *textKV) Delete(key kvstore.Key) error {
	s.Lock()
	defer s.Unlock()

	err := s.inMemoryStore.Delete(byteutils.ConcatBytes(s.realm, key))
	if err != nil {
		return err
	}

	return s.flush()
}

// DeletePrefix deletes all the entries matching the given key prefix.
func (s *textKV) DeletePrefix(prefix kvstore.KeyPrefix) error {
	s.Lock()
	defer s.Unlock()

	err := s.inMemoryStore.DeletePrefix(byteutils.ConcatBytes(s.realm, prefix))
	if err != nil {
		return err
	}
	return s.flush()
}

func (s *textKV) flush() error {
	var err error
	rec := make(map[string]interface{})
	err = s.inMemoryStore.Iterate(s.realm, func(key, value kvstore.Value) bool {
		var val interface{}
		err = s.unmarshal(value, &val)
		if err != nil {
			return false
		}
		rec[base58.Encode(key)] = val
		return true
	})
	if err != nil {
		return err
	}
	data, err := s.marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, storePerm)
}

// Batched returns a BatchedMutations interface to execute batched mutations.
func (s *textKV) Batched() (kvstore.BatchedMutations, error) {
	return &batchedMutations{
		kvStore:          s,
		deleteOperations: make(map[string]types.Empty),
		setOperations:    make(map[string]kvstore.Value),
	}, nil
}

// Flush persists all outstanding write operations to disc.
func (s *textKV) Flush() error {
	return nil
}

// Close closes the database file handles.
func (s *textKV) Close() error {
	return nil
}

type batchedMutations struct {
	sync.Mutex
	kvStore          *textKV
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	b.Lock()
	defer b.Unlock()

	strKey := string(key)
	delete(b.deleteOperations, strKey)
	b.setOperations[strKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	b.Lock()
	defer b.Unlock()

	strKey := string(key)
	delete(b.setOperations, strKey)
	b.deleteOperations[strKey] = types.Void

	return nil
}

func (b *batchedMutations) Cancel() {
	b.Lock()
	defer b.Unlock()

	b.setOperations = make(map[string]kvstore.Value)
	b.deleteOperations = make(map[string]types.Empty)
}

func (b *batchedMutations) Commit() error {
	b.Lock()
	defer b.Unlock()

	for key, value := range b.setOperations {
		err := b.kvStore.Set([]byte(key), value)
		if err != nil {
			return err
		}
	}

	for key := range b.deleteOperations {
		err := b.kvStore.Delete([]byte(key))
		if err != nil {
			return err
		}
	}
	return nil
}
