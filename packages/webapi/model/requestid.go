package model

import (
	"encoding/json"

	"github.com/iotaledger/wasp/packages/iscp"
)

// RequestID is the string representation of iscp.RequestID
type RequestID string

func NewRequestID(reqID iscp.RequestID) RequestID {
	return RequestID(reqID.String())
}

func (r RequestID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

func (r *RequestID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*r = RequestID(s)
	_, err := iscp.RequestIDFromString(s)
	return err
}

func (r RequestID) RequestID() iscp.RequestID {
	reqID, err := iscp.RequestIDFromString(string(r))
	if err != nil {
		panic(err)
	}
	return reqID
}
