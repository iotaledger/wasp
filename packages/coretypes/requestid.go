// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
	"io"
)

// RequestIDLength size of the RequestID in bytes
const RequestIDLength = valuetransaction.IDLength + 2

// RequestID is a global ID of any smart contract request.
// In ISCP, each request is a section in the smart contract transaction (sctransaction.Transaction).
// The request ID is a concatenation of the transaction ID and little-endian 2 bytes of uint16 index of the section
type RequestID [RequestIDLength]byte

// NewRequestID a constructor
func NewRequestID(txid valuetransaction.ID, index uint16) (ret RequestID) {
	copy(ret[:valuetransaction.IDLength], txid.Bytes())
	copy(ret[valuetransaction.IDLength:], util.Uint16To2Bytes(index))
	return
}

// NewRequestIDFromBase58 a constructor
func NewRequestIDFromBase58(str58 string) (ret RequestID, err error) {
	data, err := base58.Decode(str58)
	if err != nil {
		return
	}
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewRequestIDFromBytes a constructor
func NewRequestIDFromBytes(data []byte) (ret RequestID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// TransactionID of the request ID (copy)
func (rid *RequestID) TransactionID() *valuetransaction.ID {
	var ret valuetransaction.ID
	copy(ret[:], rid[:valuetransaction.IDLength])
	return &ret
}

// Index of the request ID
func (rid *RequestID) Index() uint16 {
	return util.MustUint16From2Bytes(rid[valuetransaction.IDLength:])
}

func (rid *RequestID) Write(w io.Writer) error {
	_, err := w.Write(rid[:])
	return err
}

func (rid *RequestID) Read(r io.Reader) error {
	n, err := r.Read(rid[:])
	if err != nil {
		return err
	}
	if n != RequestIDLength {
		return ErrWrongDataLength
	}
	return nil
}

// String is a human readable representation of the requestID
func (rid *RequestID) String() string {
	return fmt.Sprintf("[%d]%s", rid.Index(), rid.TransactionID().String())
}

func (rid *RequestID) Base58() string {
	return base58.Encode(rid[:])
}

func (rid *RequestID) Short() string {
	return rid.String()[:8] + ".."
}

func (rid RequestID) MarshalJSON() ([]byte, error) {
	return json.Marshal(rid.Base58())
}

func (rid *RequestID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r, err := NewRequestIDFromBase58(s)
	*rid = r
	return err
}
