// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"math"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog)

var (
	// Funcs
	FuncRetryUnprocessable = coreutil.NewEP1(Contract, "retryUnprocessable",
		coreutil.FieldWithCodec(ParamRequestID, codec.RequestID),
	)

	// Views
	ViewGetBlockInfo = coreutil.NewViewEP12(Contract, "getBlockInfo",
		coreutil.FieldWithCodecOptional(ParamBlockIndex, codec.Uint32),
		coreutil.FieldWithCodec(ParamBlockIndex, codec.Uint32),
		coreutil.FieldWithCodec(ParamBlockInfo, codec.NewCodecEx(BlockInfoFromBytes)),
	)
	ViewGetRequestIDsForBlock = coreutil.NewViewEP12(Contract, "getRequestIDsForBlock",
		coreutil.FieldWithCodecOptional(ParamBlockIndex, codec.Uint32),
		coreutil.FieldWithCodec(ParamBlockIndex, codec.Uint32),
		OutputRequestIDs{},
	)
	ViewGetRequestReceipt = coreutil.NewViewEP11(Contract, "getRequestReceipt",
		coreutil.FieldWithCodec(ParamRequestID, codec.RequestID),
		OutputRequestReceipt{},
	)
	ViewGetRequestReceiptsForBlock = coreutil.NewViewEP12(Contract, "getRequestReceiptsForBlock",
		coreutil.FieldWithCodecOptional(ParamBlockIndex, codec.Uint32),
		coreutil.FieldWithCodec(ParamBlockIndex, codec.Uint32),
		OutputRequestReceipts{},
	)
	ViewIsRequestProcessed = coreutil.NewViewEP11(Contract, "isRequestProcessed",
		coreutil.FieldWithCodec(ParamRequestID, codec.RequestID),
		coreutil.FieldWithCodec(ParamRequestProcessed, codec.Bool),
	)
	ViewGetEventsForRequest = coreutil.NewViewEP11(Contract, "getEventsForRequest",
		coreutil.FieldWithCodec(ParamRequestID, codec.RequestID),
		OutputEvents{},
	)
	ViewGetEventsForBlock = coreutil.NewViewEP12(Contract, "getEventsForBlock",
		coreutil.FieldWithCodecOptional(ParamBlockIndex, codec.Uint32),
		coreutil.FieldWithCodec(ParamBlockIndex, codec.Uint32),
		OutputEvents{},
	)
	ViewHasUnprocessable = coreutil.NewViewEP11(Contract, "hasUnprocessable",
		coreutil.FieldWithCodec(ParamRequestID, codec.RequestID),
		coreutil.FieldWithCodec(ParamUnprocessableRequestExists, codec.Bool),
	)
)

// request parameters
const (
	ParamBlockIndex                 = "n"
	ParamBlockInfo                  = "i"
	ParamContractHname              = "h"
	ParamFromBlock                  = "f"
	ParamToBlock                    = "t"
	ParamRequestID                  = "u"
	ParamRequestIndex               = "r"
	ParamRequestProcessed           = "p"
	ParamRequestRecord              = "d"
	ParamEvent                      = "e"
	ParamStateControllerAddress     = "s"
	ParamUnprocessableRequestExists = "x"
)

const (
	// Array of blockIndex => BlockInfo (pruned)
	// Covered in: TestGetEvents
	prefixBlockRegistry = "a"

	// Map of request.ID().LookupDigest() => []RequestLookupKey (pruned)
	//   LookupDigest = reqID[:6] | outputIndex
	//   RequestLookupKey = blockIndex | requestIndex
	// Covered in: TestGetEvents
	prefixRequestLookupIndex = "b"

	// Map of RequestLookupKey => RequestReceipt (pruned)
	//   RequestLookupKey = blockIndex | requestIndex
	// Covered in: TestGetEvents
	prefixRequestReceipts = "c"

	// Map of EventLookupKey => event (pruned)
	//   EventLookupKey = blockIndex | requestIndex | eventIndex
	// Covered in: TestGetEvents
	prefixRequestEvents = "d"

	// Map of requestID => unprocessableRequestRecord
	// Covered in: TestUnprocessableWithPruning
	prefixUnprocessableRequests = "u"

	// Array of requestID.
	// Temporary list of unprocessable requests that need updating the outputID field
	// Covered in: TestUnprocessableWithPruning
	prefixNewUnprocessableRequests = "U"
)

type OutputRequestIDs struct{}

func (OutputRequestIDs) Encode(reqIDs []isc.RequestID) []byte {
	return codec.SliceToArray(codec.RequestID, reqIDs)
}

func (OutputRequestIDs) Decode(r []byte) ([]isc.RequestID, error) {
	return codec.SliceFromArray(codec.RequestID, r)
}

type OutputRequestReceipt struct{}

func (OutputRequestReceipt) Encode(rec *RequestReceipt) []byte {
	if rec == nil {
		return nil
	}
	return []byte{
		ParamRequestRecord: rec.Bytes(),
		ParamBlockIndex:    codec.Uint32.Encode(rec.BlockIndex),
		ParamRequestIndex:  codec.Uint16.Encode(rec.RequestIndex),
	}
}

func (OutputRequestReceipt) Decode(r []byte) (*RequestReceipt, error) {
	if r.IsEmpty() {
		return nil, nil
	}
	blockIndex, err := codec.Uint32.Decode(r[ParamBlockIndex])
	if err != nil {
		return nil, err
	}
	reqIndex, err := codec.Uint16.Decode(r[ParamRequestIndex])
	if err != nil {
		return nil, err
	}
	rec, err := RequestReceiptFromBytes(r[ParamRequestRecord], blockIndex, reqIndex)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

type OutputRequestReceipts struct{}

func (OutputRequestReceipts) Encode(receipts []*RequestReceipt) []byte {
	ret := dict.New()
	requestReceipts := collections.NewArray(ret, ParamRequestRecord)
	for _, receipt := range receipts {
		requestReceipts.Push(receipt.Bytes())
	}
	return ret
}

func (OutputRequestReceipts) Decode(r []byte) ([]*RequestReceipt, error) {
	receipts := collections.NewArrayReadOnly(r, ParamRequestRecord)
	ret := make([]*RequestReceipt, receipts.Len())
	var err error
	blockIndex, err := codec.Uint32.Decode(r.Get(ParamBlockIndex))
	if err != nil {
		return nil, err
	}
	for i := range ret {
		ret[i], err = RequestReceiptFromBytes(receipts.GetAt(uint32(i)), blockIndex, uint16(i))
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

type OutputEvents struct{}

func (OutputEvents) Encode(events []*isc.Event) []byte {
	return codec.SliceToArray(codec.NewCodecEx(isc.EventFromBytes), events)
}

func (OutputEvents) Decode(r []byte) ([]*isc.Event, error) {
	return codec.SliceFromArray(codec.NewCodecEx(isc.EventFromBytes), r)
}

type BlockRange struct {
	From uint32
	To   uint32
}

type EventsForContractQuery struct {
	Contract   isc.Hname
	BlockRange *BlockRange
}

type InputEventsForContract struct{}

func (InputEventsForContract) Encode(q EventsForContractQuery) []byte {
	r := []byte{
		ParamContractHname: codec.Hname.Encode(q.Contract),
	}
	if q.BlockRange != nil {
		r[ParamFromBlock] = codec.Uint32.Encode(q.BlockRange.From)
		r[ParamToBlock] = codec.Uint32.Encode(q.BlockRange.To)
	}
	return r
}

func (InputEventsForContract) Decode(d []byte) (ret EventsForContractQuery, err error) {
	ret.Contract, err = codec.Hname.Decode(d[ParamContractHname])
	if err != nil {
		return
	}
	ret.BlockRange = &BlockRange{}
	ret.BlockRange.From, err = codec.Uint32.Decode(d[ParamFromBlock], 0)
	if err != nil {
		return
	}
	ret.BlockRange.To, err = codec.Uint32.Decode(d[ParamToBlock], math.MaxUint32)
	return
}
