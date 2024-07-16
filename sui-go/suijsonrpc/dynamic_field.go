package suijsonrpc

import (
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/serialization"
)

// in sui/crates/sui-types/src/dynamic_field.rs
type DynamicFieldInfo struct {
	Name       sui.DynamicFieldName                        `json:"name"`
	BcsName    sui.Base58                                  `json:"bcsName"`
	Type       serialization.TagJson[sui.DynamicFieldType] `json:"type"`
	ObjectType string                                      `json:"objectType"`
	ObjectID   sui.ObjectID                                `json:"objectId"`
	Version    sui.SequenceNumber                          `json:"version"`
	Digest     sui.ObjectDigest                            `json:"digest"`
}

type DynamicFieldPage = Page[DynamicFieldInfo, sui.ObjectID]
