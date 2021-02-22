package registry

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"time"
)

// implements BlobCacheProvide interface

// TODO blob cache cleanup

func dbKeyForBlob(h hashing.HashValue) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeBlobCache, h[:])
}

func dbKeyForBlobTTL(h hashing.HashValue) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeBlobCacheTTL)
}

// PutBlob Writes data into the registry with the key of its hash
// Also stores TTL if provided
func (r *Impl) PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error) {
	h := hashing.HashData(data)
	err := r.dbProvider.GetRegistryPartition().Set(dbKeyForBlob(h), data)
	if err != nil {
		return hashing.NilHash, err
	}
	nowis := time.Now()
	cleanAfter := nowis.Add(coretypes.DefaultTTL).UnixNano()
	if len(ttl) > 0 {
		cleanAfter = nowis.Add(ttl[0]).UnixNano()
	}
	if cleanAfter > 0 {
		err = r.dbProvider.GetRegistryPartition().Set(dbKeyForBlobTTL(h), codec.EncodeInt64(cleanAfter))
		if err != nil {
			return hashing.NilHash, err
		}
	}
	r.log.Infof("data blob has been stored. size: %d bytes, hash: %s", len(data), h)
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

func (r *Impl) HasBlob(h hashing.HashValue) (bool, error) {
	return r.dbProvider.GetRegistryPartition().Has(dbKeyForBlob(h))
}
