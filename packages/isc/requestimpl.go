package isc

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

type RequestKind byte

const (
	requestKindOnLedger RequestKind = iota
	requestKindOffLedgerISC
	requestKindOffLedgerEVMTx
	requestKindOffLedgerEVMCall
)

func IsOffledgerKind(b byte) bool {
	switch RequestKind(b) {
	case requestKindOffLedgerISC, requestKindOffLedgerEVMTx:
		return true
	}
	return false
}

func RequestFromBytes(data []byte) (Request, error) {
	return bcs.Unmarshal[Request](data)
}

// region RequestID //////////////////////////////////////////////////////////////////

type RequestID iotago.ObjectID

func (rid *RequestID) AsIotaObjectID() iotago.ObjectID {
	return iotago.ObjectID(*rid)
}

func (rid *RequestID) AsIotaAddress() iotago.Address {
	return iotago.Address(*rid)
}

const RequestIDDigestLen = 6

type RequestRef struct {
	ID   RequestID
	Hash hashing.HashValue
}

const RequestRefKeyLen = iotago.AddressLen + 32

type RequestRefKey [RequestRefKeyLen]byte

func (rrk RequestRefKey) String() string {
	return hexutil.Encode(rrk[:])
}

func RequestRefFromBytes(data []byte) (*RequestRef, error) {
	return bcs.Unmarshal[*RequestRef](data)
}

func RequestRefFromRequest(req Request) *RequestRef {
	return &RequestRef{ID: req.ID(), Hash: RequestHash(req)}
}

func RequestRefsFromRequests(reqs []Request) []*RequestRef {
	rr := make([]*RequestRef, len(reqs))
	for i := range rr {
		rr[i] = RequestRefFromRequest(reqs[i])
	}
	return rr
}

func (ref *RequestRef) AsKey() RequestRefKey {
	var key RequestRefKey
	copy(key[:], ref.Bytes())
	return key
}

func (ref *RequestRef) IsFor(req Request) bool {
	if ref.ID != req.ID() {
		return false
	}
	return ref.Hash == RequestHash(req)
}

func (ref *RequestRef) Bytes() []byte {
	return bcs.MustMarshal(ref)
}

func (ref *RequestRef) String() string {
	return fmt.Sprintf("{requestRef, id=%v, hash=%v}", ref.ID.String(), ref.Hash.Hex())
}

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func RequestIDFromBytes(data []byte) (ret RequestID, err error) {
	return bcs.Unmarshal[RequestID](data)
}

func RequestIDFromEVMTxHash(txHash common.Hash) RequestID {
	return RequestID(txHash)
}

func RequestIDFromString(s string) (ret RequestID, err error) {
	data, err := cryptolib.DecodeHex(s)
	if err != nil {
		return RequestID{}, err
	}

	if len(data) != iotago.AddressLen {
		return ret, errors.New("error parsing requestID: wrong length")
	}

	requestID := RequestID{}
	copy(requestID[:], data)
	return requestID, nil
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	copy(ret[:RequestIDDigestLen], rid[:RequestIDDigestLen])
	// last 2 bytes are the outputindex
	copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], rid[len(rid)-2:])
	return ret
}

func (rid RequestID) Bytes() []byte {
	return rid[:]
}

func (rid RequestID) Equals(other RequestID) bool {
	return rid == other
}

func (rid RequestID) String() string {
	return hexutil.Encode(rid[:])
}

func (rid RequestID) Short() string {
	ridString := rid.String()
	return fmt.Sprintf("%s..%s", ridString[2:6], ridString[len(ridString)-4:])
}

// endregion ////////////////////////////////////////////////////////////

// region RequestMetadata //////////////////////////////////////////////////

type RequestMetadata struct {
	SenderContract ContractIdentity `json:"senderContract"`
	Message        Message          `json:"message"`
	// AllowanceBCS is either empty or a BCS-encoded iscmove.Allowance.
	AllowanceBCS []byte `json:"allowanceBcs"`
	// gas budget
	GasBudget uint64 `json:"gasBudget"`
}

func RequestMetadataFromBytes(data []byte) (*RequestMetadata, error) {
	return bcs.Unmarshal[*RequestMetadata](data)
}

func (meta *RequestMetadata) Bytes() []byte {
	return bcs.MustMarshal(meta)
}
