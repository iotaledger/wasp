package coretypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/mr-tron/base58"
	"io"
)

const RequestIDLength = valuetransaction.IDLength + 2

type RequestID [RequestIDLength]byte

func NewRequestID(txid valuetransaction.ID, index Int16) (ret RequestID) {
	copy(ret[:valuetransaction.IDLength], txid.Bytes())
	copy(ret[valuetransaction.IDLength:], index.Bytes())
	return
}

func NewRequestIDFromBase58(str58 string) (ret RequestID, err error) {
	data, err := base58.Decode(str58)
	if err != nil {
		return
	}
	err = ret.Read(bytes.NewReader(data))
	return
}

func NewRequestIDFromBytes(data []byte) (ret RequestID, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

func (rid *RequestID) Bytes() []byte {
	return rid[:]
}

func (rid *RequestID) TransactionID() *valuetransaction.ID {
	var ret valuetransaction.ID
	copy(ret[:], rid[:valuetransaction.IDLength])
	return &ret
}

func (rid *RequestID) Index() (ret Int16) {
	ret, _ = NewInt16From2Bytes(rid[valuetransaction.IDLength:])
	return
}

func (rid *RequestID) Write(w io.Writer) error {
	_, err := w.Write(rid.Bytes())
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

func (rid *RequestID) String() string {
	return fmt.Sprintf("[%d]%s", rid.Index(), rid.TransactionID().String())
}

func (rid *RequestID) Base58() string {
	return base58.Encode(rid.Bytes())
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
