package graphqltypes

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type DynamicFieldType string

const (
	DynamicFieldTypeDynamicField  DynamicFieldType = "DynamicField"
	DynamicFieldTypeDynamicObject DynamicFieldType = "DynamicObject"
)

type DynamicFieldInfo struct {
	Name       iotago.DynamicFieldName
	BcsName    []byte
	Type       DynamicFieldType
	ObjectType string
	ObjectID   iotago.ObjectID
	Version    uint64
	Digest     iotago.ObjectDigest
	// ValueJson stores the JSON representation of wrapped dynamic field values
	// (only populated for DynamicField type, not for DynamicObject)
	ValueJson json.RawMessage
}
