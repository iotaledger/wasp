// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
	"io"
)

// RequestID is a global ID of any smart contract request.
// In ISCP, each request is a section in the smart contract transaction (sctransaction.Transaction).
// The request ID is a concatenation of the transaction ID and little-endian 2 bytes of uint16 index of the section
type RequestID struct {
	transactionID ledgerstate.TransactionID
	index         uint16
}

// NewRequestID a constructor
func NewRequestID(txid ledgerstate.TransactionID, index uint16) (ret RequestID) {
	ret.transactionID = txid
	ret.index = index
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
func (rid *RequestID) TransactionID() ledgerstate.TransactionID {
	return rid.transactionID
}

// Index of the request ID
func (rid *RequestID) Index() uint16 {
	return rid.index
}

func (rid *RequestID) Write(w io.Writer) error {
	if _, err := w.Write(rid.transactionID[:]); err != nil {
		return err
	}
	if err := util.WriteUint16(w, rid.index); err != nil {
		return err
	}
	return nil
}

func (rid *RequestID) Read(r io.Reader) error {
	if n, err := r.Read(rid.transactionID[:]); err != nil || n != ledgerstate.TransactionIDLength {
		return ErrWrongDataLength
	}
	if err := util.ReadUint16(r, &rid.index); err != nil {
		return err
	}
	return nil
}

func (rid *RequestID) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(rid.transactionID[:])
	util.WriteUint16(&buf, rid.index)
	return buf.Bytes()
}

// String is a human readable representation of the requestID
func (rid *RequestID) String() string {
	return fmt.Sprintf("[%d]%s", rid.Index(), rid.TransactionID().Base58())
}

func (rid *RequestID) Short() string {
	return rid.String()[:8] + ".."
}

func (rid *RequestID) Base58() string {
	return base58.Encode(rid.Bytes())
}

func (rid *RequestID) MarshalJSON() ([]byte, error) {
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
