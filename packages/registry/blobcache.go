package registry

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
)

func dbKeyForBlob(h hashing.HashValue) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeBlob, h[:])
}

// Writes data into the registry with the key of its hash
func (r *Impl) PutBlob(data []byte) (hashing.HashValue, error) {
	h := hashing.HashData(data)
	err := r.dbProvider.GetRegistryPartition().Set(dbKeyForBlob(h), data)
	if err != nil {
		return hashing.HashValue{}, err
	}
	return h, nil
}

// Reads data from registry by hash. Returns existence flag
func (r *Impl) GetBlob(h hashing.HashValue) ([]byte, bool, error) {
	ret, err := r.dbProvider.GetRegistryPartition().Get(dbKeyForBlob(h))
	if err == kvstore.ErrKeyNotFound {
		return nil, false, nil
	}
	return ret, ret != nil && err == nil, err
}
