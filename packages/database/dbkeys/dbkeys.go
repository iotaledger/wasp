package dbkeys

import (
	"bytes"
)

const (
	ObjectTypeDBSchemaVersion = byte(iota) + 'A'
	ObjectTypeChainRecord
	ObjectTypeCommitteeRecord
	ObjectTypeDistributedKeyData
	ObjectTypeState
	ObjectTypeTrie
	ObjectTypeBlock
	ObjectTypeNodeIdentity
	ObjectTypeBlobCache
	ObjectTypeBlobCacheTTL
	ObjectTypeTrustedPeer
)

// MakeKey makes key within the partition. It consists of one byte for object type
// and arbitrary byte fragments concatenated together
func MakeKey(objType byte, keyBytes ...[]byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(objType)
	for _, b := range keyBytes {
		buf.Write(b)
	}
	return buf.Bytes()
}
