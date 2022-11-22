package dbkeys

import (
	"bytes"
)

const (
	ObjectTypeDBSchemaVersion    = byte(iota) + 'A'
	ObjectTypeChainRecord        // deprecated
	ObjectTypeCommitteeRecord    // unused
	ObjectTypeDistributedKeyData // unused
	ObjectTypeState
	ObjectTypeTrie
	ObjectTypeBlock
	ObjectTypeNodeIdentity     // deprecated
	ObjectTypeBlobCache        // unused
	ObjectTypeBlobCacheTTL     // unused
	ObjectTypeTrustedPeer      // deprecated
	ObjectTypeConsensusJournal // deprecated
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
