package iotajsonrpc

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
)

// in iotago/crates/iotago-types/src/dynamic_field.rs
type DynamicFieldInfo struct {
	Name       iotago.DynamicFieldName                        `json:"name"`
	BcsName    iotago.Base64Data                              `json:"bcsName"`
	Type       serialization.TagJson[iotago.DynamicFieldType] `json:"type"`
	ObjectType string                                         `json:"objectType"`
	ObjectID   iotago.ObjectID                                `json:"objectId"`
	Version    iotago.SequenceNumber                          `json:"version"`
	Digest     iotago.ObjectDigest                            `json:"digest"`
	// ValueJson stores the JSON representation of wrapped dynamic field values
	// (only populated for DynamicField type, not for DynamicObject)
	ValueJson json.RawMessage `json:"valueJson,omitempty"`
}

type DynamicFieldPage = Page[DynamicFieldInfo, iotago.ObjectID]
