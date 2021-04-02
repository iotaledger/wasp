package coretypes

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"time"
)

const DefaultTTL = 1 * time.Hour

type BlobCache interface {
	GetBlob(h hashing.HashValue) ([]byte, bool, error)
	HasBlob(h hashing.HashValue) (bool, error)
	// PutBlob ttl s TimeToLive, expiration time in Unix nanoseconds
	PutBlob(data []byte, ttl ...time.Duration) (hashing.HashValue, error)
}
