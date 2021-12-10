package iscp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// region Request //////////////////////////////////////////////////////
// Request has two main implementations
// - RequestOnLedger
// - RequestOffLedger
type Request interface {
	// index == 0 for off ledger requests
	ID() RequestID
	// is request on transaction of not
	IsOffLedger() bool
	// true or false for on-ledger requests, always true for off-ledger
	IsFeePrepaid() bool
	// arguments of the call with the flag if they are ready. No arguments mean empty dictionary and true
	Params() (dict.Dict, bool)
	// account of the sender
	SenderAccount() *AgentID
	// address of the sender for all requests,
	SenderAddress() ledgerstate.Address
	// returns contract/entry point pair
	Target() RequestTarget
	// Timestamp returns a request TX timestamp, if such TX exist, otherwise zero is returned.
	Timestamp() time.Time
	// Bytes returns binary representation of the request
	Bytes() []byte
	// Hash returns the hash of the request (used for consensus)
	Hash() [32]byte
	// String representation of the request (humanly readable)
	String() string
}

type RequestTarget struct {
	Contract   Hname
	EntryPoint Hname
}

func NewRequestTarget(contract, entryPoint Hname) RequestTarget {
	return RequestTarget{
		Contract:   contract,
		EntryPoint: entryPoint,
	}
}

func TakeRequestIDs(reqs ...Request) []RequestID {
	ret := make([]RequestID, len(reqs))
	for i := range reqs {
		ret[i] = reqs[i].ID()
	}
	return ret
}

// endregion ///////////////////////////////////////////////////////////

// region RequestID ///////////////////////////////////////////////////////////////
type RequestID ledgerstate.OutputID

const RequestIDDigestLen = 6

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be uniques
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid ledgerstate.TransactionID, index uint16) RequestID {
	return RequestID(ledgerstate.NewOutputID(txid, index))
}

func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	ret, err := ledgerstate.OutputIDFromMarshalUtil(mu)
	return RequestID(ret), err
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

func RequestIDFromBase58(b58 string) (ret RequestID, err error) {
	var oid ledgerstate.OutputID
	oid, err = ledgerstate.OutputIDFromBase58(b58)
	if err != nil {
		return
	}
	ret = RequestID(oid)
	return
}

func RequestIDFromString(s string) (ret RequestID, err error) {
	if !strings.HasPrefix(s, "[") {
		return RequestIDFromBase58(s)
	}
	parts := strings.Split(s[1:], "]")
	if len(parts) != 2 {
		err = xerrors.New("RequestIDFromString: wrong format")
		return
	}
	index, err2 := strconv.ParseUint(parts[0], 10, 16)
	if err2 != nil {
		err = xerrors.Errorf("RequestIDFromString: %v", err2)
		return
	}
	txid, err3 := ledgerstate.TransactionIDFromBase58(parts[1])
	if err3 != nil {
		err = xerrors.Errorf("RequestIDFromString: %v", err3)
		return
	}
	return NewRequestID(txid, uint16(index)), nil
}

func (rid RequestID) OutputID() ledgerstate.OutputID {
	return ledgerstate.OutputID(rid)
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	copy(ret[:RequestIDDigestLen], rid[:RequestIDDigestLen])
	copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], util.Uint16To2Bytes(rid.OutputID().OutputIndex()))
	return ret
}

// Base58 returns a base58 encoded version of the request id.
func (rid RequestID) Base58() string {
	return ledgerstate.OutputID(rid).Base58()
}

func (rid RequestID) Bytes() []byte {
	return rid[:]
}

func (rid RequestID) String() string {
	return OID(rid.OutputID())
}

func (rid RequestID) Short() string {
	txid := rid.OutputID().TransactionID().Base58()
	return fmt.Sprintf("[%d]%s", rid.OutputID().OutputIndex(), txid[:6]+"..")
}

func OID(o ledgerstate.OutputID) string {
	return fmt.Sprintf("[%d]%s", o.OutputIndex(), o.TransactionID().Base58())
}

func ShortRequestIDs(ids []RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}

// endregion ////////////////////////////////////////////////////////////////////////////////////
