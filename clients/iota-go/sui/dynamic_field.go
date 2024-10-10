package sui

import (
	"github.com/iotaledger/wasp/clients/iota-go/sui/serialization"
)

type DynamicFieldType struct {
	DynamicField  *serialization.EmptyEnum `json:"DynamicField"`
	DynamicObject *serialization.EmptyEnum `json:"DynamicObject"`
}

func (d DynamicFieldType) Tag() string {
	return ""
}

func (d DynamicFieldType) Content() string {
	return ""
}

type DynamicFieldName struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}
