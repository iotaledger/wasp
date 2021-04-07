package coretypes

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"time"
)

//inMemoryBlobCache is supposed to be used as BlobCache in tests through
//factory method NewInMemoryBlobCache
//NOTE: Implements coretypes.BlobCache
type inMemoryBlobCache struct {
	b map[hashing.HashValue][]byte
}

//NewInMemoryBlobCache is a factory method for inMemoryBlobCache
func NewInMemoryBlobCache() *inMemoryBlobCache {
	return &inMemoryBlobCache{make(map[hashing.HashValue][]byte)}
}

func (d *inMemoryBlobCache) GetBlob(h hashing.HashValue) ([]byte, bool, error) {
	ret, ok := d.b[h]
	if !ok {
		return nil, false, nil
	}
	return ret, true, nil
}

func (d *inMemoryBlobCache) HasBlob(h hashing.HashValue) (bool, error) {
	_, ok := d.b[h]
	return ok, nil
}

func (d *inMemoryBlobCache) PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error) {
	h := hashing.HashData(data)
	c := make([]byte, len(data))
	copy(c, data)
	d.b[h] = c
	return h, nil
}
