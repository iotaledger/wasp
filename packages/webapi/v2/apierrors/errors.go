package apierrors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func ChainNotFoundError(chainID string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, fmt.Sprintf("Chain ID: %v not found", chainID), nil)
}

func UserNotFoundError(username string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, fmt.Sprintf("User: %v not found", username), nil)
}

func BodyIsEmptyError() *HTTPError {
	return InvalidPropertyError("body", errors.New("A valid body is required"))
}

func InvalidPeerPublicKeys(invalidPeerPubKeys []string) *HTTPError {
	joinedKeys := strings.Join(invalidPeerPubKeys, ";")
	return NewHTTPError(http.StatusBadRequest, "invalid peer public keys", errors.New(joinedKeys))
}

func InvalidPropertyError(propertyName string, err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid property: %v", propertyName), err)
}

func ContractExecutionError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "Failed to execute contract request", err)
}

func InvalidOffLedgerRequestError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "Supplied offledger request is invalid", err)
}

func ReceiptError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "Failed to get receipt", err)
}

func InternalServerError(err error) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, "Unknown error has occoured", err)
}
