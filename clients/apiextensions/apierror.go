// Package apiextensions provides extended API functionality for interacting with the Wasp node API.
package apiextensions

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
)

type APIDetailError struct {
	Message string
	Error   string
}

type APIError struct {
	Error       string
	DetailError *APIDetailError
}

func AsAPIError(err error) (*APIError, bool) {
	genericError, ok := err.(*apiclient.GenericOpenAPIError)

	if !ok {
		return nil, false
	}

	apiError := APIError{
		Error: genericError.Error(),
	}

	var detailError APIDetailError
	if json.Unmarshal(genericError.Body(), &detailError) == nil {
		apiError.DetailError = &detailError
	}

	return &apiError, true
}
