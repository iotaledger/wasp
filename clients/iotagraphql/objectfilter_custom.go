// Package iotagraphql provides GraphQL client types for the IOTA network.
package iotagraphql

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// MarshalJSON provides custom JSON marshaling for ObjectFilter to skip zero values.
// This is needed because genqlient generates non-pointer fields for optional GraphQL input fields,
// which means they always have a value (even if it's the zero value).
// The zero-value for iotago.Address (all zeros) was being sent to the GraphQL API and
// filtering out results incorrectly.
func (f ObjectFilter) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})

	if f.Type != "" {
		result["type"] = f.Type
	}

	// Check if Owner is non-zero
	zeroOwner := iotago.Address{}
	if f.Owner != zeroOwner {
		result["owner"] = f.Owner
	}

	if len(f.ObjectIds) > 0 {
		result["objectIds"] = f.ObjectIds
	}

	if len(f.ObjectKeys) > 0 {
		result["objectKeys"] = f.ObjectKeys
	}

	return json.Marshal(result)
}
