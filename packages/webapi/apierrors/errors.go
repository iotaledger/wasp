package apierrors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func ChainNotFoundError() *HTTPError {
	return NewHTTPError(http.StatusNotFound, "Chain ID not found", nil)
}

func UserNotFoundError(username string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, fmt.Sprintf("User: %v not found", username), nil)
}

func UserCanNotBeDeleted(username string, explanation string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("User: %v not be deleted. Reason: %v", username, explanation), nil)
}

func BodyIsEmptyError() *HTTPError {
	return InvalidPropertyError("body", errors.New("a valid body is required"))
}

func InvalidPeerPublicKeys(invalidPeerPubKeys []string) *HTTPError {
	joinedKeys := strings.Join(invalidPeerPubKeys, ";")
	return NewHTTPError(http.StatusBadRequest, "invalid peer public keys", errors.New(joinedKeys))
}

func InvalidPropertyError(propertyName string, err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid property: %v", propertyName), err)
}

func PeerNameNotFoundError(name string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, fmt.Sprintf("couldn't find peer with name %s", name), nil)
}

func SelfAsPeerError() *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "cannot add self as a peer", nil)
}

func InvalidPeerName() *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "name must be in slug format (lowecase and hyphens only)", nil)
}

func ContractExecutionError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "Failed to execute contract request", err)
}

func InvalidOffLedgerRequestError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "Supplied offledger request is invalid", err)
}

func NoRecordFoundError(err error) *HTTPError {
	return NewHTTPError(http.StatusNotFound, "Record not found", err)
}

func ReceiptError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "Failed to get receipt", err)
}

func Timeout(msg string) *HTTPError {
	return NewHTTPError(http.StatusRequestTimeout, msg, nil)
}
