package solo

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"time"
)

type dummyBlobCache struct {
	b map[hashing.HashValue][]byte
}

func newDummyBlobCache() *dummyBlobCache {
	return &dummyBlobCache{make(map[hashing.HashValue][]byte)}
}

func (d *dummyBlobCache) GetBlob(h hashing.HashValue) ([]byte, bool, error) {
	ret, ok := d.b[h]
	if !ok {
		return nil, false, nil
	}
	return ret, true, nil
}

func (d *dummyBlobCache) HasBlob(h hashing.HashValue) (bool, error) {
	_, ok := d.b[h]
	return ok, nil
}

func (d *dummyBlobCache) PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error) {
	h := hashing.HashData(data)
	c := make([]byte, len(data))
	copy(c, data)
	d.b[h] = c
	return h, nil
}
