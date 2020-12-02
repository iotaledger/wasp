package coret

import (
	"bytes"
	"encoding/json"
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/mr-tron/base58"
	"io"
)

const RequestIDLength = valuetransaction.IDLength + 2

// RequestID is a global ID of any request. It is a concatenation of trasaction ID and 2 bytes of index of the section
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

func (rid *RequestID) Bytes() []byte {
	return rid[:]
}

// TransactionID of the request ID
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
