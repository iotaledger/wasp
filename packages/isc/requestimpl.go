package isc

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type RequestKind rwutil.Kind

const (
	requestKindOnLedger RequestKind = iota
	requestKindOffLedgerISC
	requestKindOffLedgerEVMTx
	requestKindOffLedgerEVMCall
)

func RequestFromBytes(data []byte) (Request, error) {
	return RequestFromMarshalUtil(marshalutil.New(data))
}

func RequestFromMarshalUtil(mu *marshalutil.MarshalUtil) (Request, error) {
	rr := rwutil.NewMuReader(mu)
	return RequestFromReader(rr), rr.Err
}

func RequestFromReader(rr *rwutil.Reader) (ret Request) {
	kind := rr.ReadKind()
	switch RequestKind(kind) {
	case requestKindOnLedger:
		ret = new(onLedgerRequestData)
	case requestKindOffLedgerISC:
		ret = new(offLedgerRequestData)
	case requestKindOffLedgerEVMTx:
		ret = new(evmOffLedgerTxRequest)
	case requestKindOffLedgerEVMCall:
		ret = new(evmOffLedgerCallRequest)
	default:
		if rr.Err == nil {
			rr.Err = errors.New("invalid Request kind")
			return nil
		}
	}
	rr.PushBack().WriteKind(kind)
	rr.Read(ret)
	return ret
}

// region RequestID //////////////////////////////////////////////////////////////////

type RequestID iotago.OutputID

const RequestIDDigestLen = 6

type RequestRef struct {
	ID   RequestID
	Hash hashing.HashValue
}

const RequestRefKeyLen = iotago.OutputIDLength + 32

type RequestRefKey [RequestRefKeyLen]byte

func (rrk RequestRefKey) String() string {
	return iotago.EncodeHex(rrk[:])
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

func (rr *RequestRef) AsKey() RequestRefKey {
	var key RequestRefKey
	copy(key[:], rr.Bytes())
	return key
}

func (rr *RequestRef) IsFor(req Request) bool {
	if rr.ID != req.ID() {
		return false
	}
	return rr.Hash == RequestHash(req)
}

func (rr *RequestRef) Bytes() []byte {
	return append(rr.Hash[:], rr.ID[:]...)
}

func (rr *RequestRef) String() string {
	return fmt.Sprintf("{requestRef, id=%v, hash=%v}", rr.ID.String(), rr.Hash.Hex())
}

func RequestRefFromBytes(data []byte) (*RequestRef, error) {
	reqID, err := RequestIDFromBytes(data[hashing.HashSize:])
	if err != nil {
		return nil, err
	}
	ret := &RequestRef{
		ID: reqID,
	}
	copy(ret.Hash[:], data[:hashing.HashSize])

	return ret, nil
}

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid iotago.TransactionID, index uint16) RequestID {
	return RequestID(iotago.OutputIDFromTransactionIDAndIndex(txid, index))
}

func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	outputIDData, err := mu.ReadBytes(iotago.OutputIDLength)
	if err != nil {
		return RequestID{}, err
	}

	outputID := iotago.OutputID{}
	copy(outputID[:], outputIDData)
	return RequestID(outputID), nil
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

func RequestIDFromString(s string) (ret RequestID, err error) {
	data, err := iotago.DecodeHex(s)
	if err != nil {
		return RequestID{}, err
	}

	if len(data) != iotago.OutputIDLength {
		return ret, errors.New("error parsing requestID: wrong length")
	}

	requestID := RequestID{}
	copy(requestID[:], data)
	return requestID, nil
}

func (rid RequestID) OutputID() iotago.OutputID {
	return iotago.OutputID(rid)
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
	return iotago.EncodeHex(rid[:])
}

func (rid RequestID) Short() string {
	ridString := rid.String()
	return fmt.Sprintf("%s..%s", ridString[2:6], ridString[len(ridString)-4:])
}

func ShortRequestIDs(ids []RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}

func ShortRequestIDsFromRequests(reqs []Request) []string {
	requestIDs := make([]RequestID, len(reqs))
	for i := range reqs {
		requestIDs[i] = reqs[i].ID()
	}
	return ShortRequestIDs(requestIDs)
}

// endregion ////////////////////////////////////////////////////////////

// region RequestMetadata //////////////////////////////////////////////////

type RequestMetadata struct {
	SenderContract Hname `json:"senderContract"`
	// ID of the target smart contract
	TargetContract Hname `json:"targetContract"`
	// entry point code
	EntryPoint Hname `json:"entryPoint"`
	// request arguments
	Params dict.Dict `json:"params"`
	// Allowance intended to the target contract to take. Nil means zero allowance
	Allowance *Assets `json:"allowance"`
	// gas budget
	GasBudget uint64 `json:"gasBudget"`
}

func requestMetadataFromFeatureSet(set iotago.FeatureSet) (*RequestMetadata, error) {
	metadataFeatBlock := set.MetadataFeature()
	if metadataFeatBlock == nil {
		// IMPORTANT: this cannot return an empty `&RequestMetadata{}` object because that could cause `isInternalUTXO` check to fail
		return nil, nil
	}
	return RequestMetadataFromBytes(metadataFeatBlock.Data)
}

func RequestMetadataFromBytes(data []byte) (*RequestMetadata, error) {
	ret := &RequestMetadata{}
	err := ret.ReadFromMarshalUtil(marshalutil.New(data))
	return ret, err
}

// returns nil if nil pointer receiver is cloned
func (p *RequestMetadata) Clone() *RequestMetadata {
	if p == nil {
		return nil
	}

	return &RequestMetadata{
		SenderContract: p.SenderContract,
		TargetContract: p.TargetContract,
		EntryPoint:     p.EntryPoint,
		Params:         p.Params.Clone(),
		Allowance:      p.Allowance.Clone(),
		GasBudget:      p.GasBudget,
	}
}

func (p *RequestMetadata) Bytes() []byte {
	mu := marshalutil.New()
	p.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (p *RequestMetadata) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(p.SenderContract).
		Write(p.TargetContract).
		Write(p.EntryPoint).
		WriteUint64(p.GasBudget)
	p.Params.WriteToMarshalUtil(mu)
	p.Allowance.WriteToMarshalUtil(mu)
}

func (p *RequestMetadata) ReadFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if p.SenderContract, err = HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.TargetContract, err = HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.EntryPoint, err = HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.GasBudget, err = mu.ReadUint64(); err != nil {
		return err
	}
	if p.Params, err = dict.FromMarshalUtil(mu); err != nil {
		return err
	}
	if p.Allowance, err = AssetsFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}
