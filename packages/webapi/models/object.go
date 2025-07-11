package models

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
)

// ObjectTypeJSON is the representation of a Iota object type that is used in the JSON API
type ObjectTypeJSON string

func ToTypeJSON(t iotago.ObjectType) ObjectTypeJSON {
	return ObjectTypeJSON(t.String())
}

func ToIotaObjectJSON(o isc.IotaObject) IotaObjectJSON {
	return IotaObjectJSON{
		ID:   o.ID.ToHex(),
		Type: ToTypeJSON(o.Type),
	}
}

type IotaObjectJSON struct {
	ID   string         `json:"id" swagger:"required,desc(Hex-encoded object ID)"`
	Type ObjectTypeJSON `json:"type" swagger:"required"`
}
