package registry_pkg

import (
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// implements BlobCacheProvide interface

// TODO blob cache cleanup

func dbKeyForBlob(h hashing.HashValue) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeBlobCache, h[:])
}

func dbKeyForBlobTTL(h hashing.HashValue) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeBlobCacheTTL)
}

const BlobCacheDefaultTTL = 1 * time.Hour

// PutBlob Writes data into the registry with the key of its hash
// Also stores TTL if provided
func (r *Impl) PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error) {
	h := hashing.HashData(data)
	err := r.store.Set(dbKeyForBlob(h), data)
	if err != nil {
		return hashing.NilHash, err
	}
	nowis := time.Now()
	cleanAfter := nowis.Add(BlobCacheDefaultTTL).UnixNano()
	if len(ttl) > 0 {
		cleanAfter = nowis.Add(ttl[0]).UnixNano()
	}
	if cleanAfter > 0 {
		err = r.store.Set(dbKeyForBlobTTL(h), codec.EncodeInt64(cleanAfter))
		if err != nil {
			return hashing.NilHash, err
		}
	}
	r.log.Infof("data blob has been stored. size: %d bytes, hash: %s", len(data), h)
	return h, nil
}

// Reads data from registry by hash. Returns existence flag
func (r *Impl) GetBlob(h hashing.HashValue) ([]byte, bool, error) {
	ret, err := r.store.Get(dbKeyForBlob(h))
	if err == kvstore.ErrKeyNotFound {
		return nil, false, nil
	}
	return ret, ret != nil && err == nil, err
}

func (r *Impl) HasBlob(h hashing.HashValue) (bool, error) {
	return r.store.Has(dbKeyForBlob(h))
}
