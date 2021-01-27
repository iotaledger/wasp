package dbprovider

import "bytes"

const (
	ObjectTypeDBSchemaVersion byte = iota
	ObjectTypeChainRecord
	ObjectTypeDistributedKeyData
	ObjectTypeSolidState
	ObjectTypeStateUpdateBatch
	ObjectTypeProcessedRequestId
	ObjectTypeSolidStateIndex
	ObjectTypeStateVariable
	ObjectTypeProgramMetadata
	ObjectTypeNodeIdentity
	ObjectTypeBlob
)

// MakeKey makes key within the partition. It consists to one byte for object type
// and arbitrary byte fragments concatenated together
func MakeKey(objType byte, keyBytes ...[]byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(objType)
	for _, b := range keyBytes {
		buf.Write(b)
	}
	return buf.Bytes()
}
