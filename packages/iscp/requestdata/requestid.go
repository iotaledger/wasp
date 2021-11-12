package requestdata

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp/requestdata/iotago"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
)

type RequestID iotago.UTXOInput

const RequestIDDigestLen = 6

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid ledgerstate.TransactionID, index uint16) RequestID {
	return RequestID{}
}

func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	//ret, err := ledgerstate.OutputIDFromMarshalUtil(mu)
	return RequestID{}, nil
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

func RequestIDFromBase58(b58 string) (ret RequestID, err error) {
	//var oid ledgerstate.OutputID
	//oid, err = ledgerstate.OutputIDFromBase58(b58)
	//if err != nil {
	//	return
	//}
	//ret = RequestID(oid)
	ret = RequestID{}
	return
}

func (rid RequestID) OutputID() iotago.UTXOInput {
	return iotago.UTXOInput(rid)
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	//copy(ret[:RequestIDDigestLen], rid[:RequestIDDigestLen])
	//copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], util.Uint16To2Bytes(rid.OutputID().OutputIndex()))
	return ret
}

// Base58 returns a base58 encoded version of the request id.
func (rid RequestID) Base58() string {
	//return ledgerstate.OutputID(rid).Base58()
	return ""
}

func (rid RequestID) Bytes() []byte {
	return nil
}

func (rid RequestID) String() string {
	return OID(rid.OutputID())
}

func (rid RequestID) Short() string {
	//txid := rid.OutputID().TransactionID().Base58()
	//return fmt.Sprintf("[%d]%s", rid.OutputID().TransactionOutputIndex, txid[:6]+"..")
	return ""
}

func OID(o iotago.UTXOInput) string {
	return fmt.Sprintf("[%d]%s", o.TransactionOutputIndex, "") // o.TransactionID.Base58())
}

func ShortRequestIDs(ids []RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}
