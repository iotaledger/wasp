package graphqltypes

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// DynamicFieldInfo represents dynamic field information from iotago-types.
type DynamicFieldInfo struct {
	Name       iotago.DynamicFieldName `json:"name"`
	BcsName    iotago.Base64Data       `json:"bcsName"`
	Type       iotago.DynamicFieldType `json:"type"`
	ObjectType string                  `json:"objectType"`
	ObjectID   iotago.ObjectID         `json:"objectId"`
	Version    iotago.SequenceNumber   `json:"version"`
	Digest     iotago.ObjectDigest     `json:"digest"`
	// ValueJSON stores the JSON representation of wrapped dynamic field values
	// (only populated for DynamicField type, not for DynamicObject)
	ValueJSON json.RawMessage `json:"valueJson,omitempty"`
}

type DynamicFieldPage = Page[DynamicFieldInfo, iotago.ObjectID]
