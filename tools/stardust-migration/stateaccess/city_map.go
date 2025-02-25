package stateaccess

import (
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/samber/lo"
)

var _ kvstore.KVStore = &CityMap{}

type CityMap struct {
	Data   map[string][]byte
	_Realm []byte
}

func (s *CityMap) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {
	panic("implement me")
}

func (s *CityMap) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {
	panic("implement me")
}

func (s *CityMap) Clear() error {
	for k := range s.Data {
		delete(s.Data, k)
	}

	s.Data = nil
	return nil
}

func (s *CityMap) Get(key kvstore.Key) (value kvstore.Value, err error) {
	mapKey := string(byteutils.ConcatBytes(s._Realm, key))
	if !lo.Must(s.Has(key)) {
		return nil, kvstore.ErrKeyNotFound
	}

	return s.Data[mapKey], nil
}

func (s *CityMap) Set(key kvstore.Key, value kvstore.Value) error {
	mapKey := string(byteutils.ConcatBytes(s._Realm, key))
	s.Data[mapKey] = value
	return nil
}

func (s *CityMap) Has(key kvstore.Key) (bool, error) {
	mapKey := string(byteutils.ConcatBytes(s._Realm, key))
	_, ok := s.Data[mapKey]
	return ok, nil
}

func (s *CityMap) Delete(key kvstore.Key) error {
	mapKey := string(byteutils.ConcatBytes(s._Realm, key))

	if lo.Must(s.Has(key)) {
		s.Data[mapKey] = nil
		delete(s.Data, mapKey)
	} else {
		fmt.Println("COULD NOT FIND KEY TO DELETE")
		fmt.Printf("%s -> %s\n", hexutil.Encode([]byte(mapKey)), string(mapKey))
	}

	return nil
}

func (s *CityMap) DeletePrefix(prefix kvstore.KeyPrefix) error {
	panic("DeletePrefix Needed")
}

func (s *CityMap) Flush() error {
	return nil
}

func (s *CityMap) Close() error {
	s.Data = nil
	return nil
}

func (s *CityMap) Batched() (kvstore.BatchedMutations, error) {
	return &batchedMutations{
		kvStore:          s,
		setOperations:    make(map[string]kvstore.Value),
		deleteOperations: make(map[string]types.Empty),
	}, nil
}

func NewCityMap() *CityMap {
	return &CityMap{
		Data:   make(map[string][]byte),
		_Realm: make([]byte, 0),
	}
}

func NewCityMapWithData(data map[string][]byte) *CityMap {
	return &CityMap{
		Data:   data,
		_Realm: make([]byte, 0),
	}
}

func (s *CityMap) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return &CityMap{
		Data:   s.Data,
		_Realm: realm,
	}, nil
}

func (s *CityMap) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

func (s *CityMap) Realm() kvstore.Realm {
	return byteutils.ConcatBytes(s._Realm)
}

func (c *CityMap) Clone() map[string][]byte {
	dst := make(map[string][]byte, len(c.Data))
	for k, v := range c.Data {
		dst[k] = append([]byte(nil), v...)
	}

	c.Clear()

	return dst
}

type batchedMutations struct {
	kvStore          *CityMap
	setOperations    map[string]kvstore.Value
	deleteOperations map[string]types.Empty
	closed           *atomic.Bool
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	stringKey := byteutils.ConcatBytesToString(key)

	delete(b.deleteOperations, stringKey)
	b.setOperations[stringKey] = value

	return nil
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	stringKey := byteutils.ConcatBytesToString(key)

	delete(b.setOperations, stringKey)
	b.deleteOperations[stringKey] = types.Void

	return nil
}

func (b *batchedMutations) Cancel() {
	b.setOperations = make(map[string]kvstore.Value)
	b.deleteOperations = make(map[string]types.Empty)
}

func (b *batchedMutations) Commit() error {
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
