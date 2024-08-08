// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog)

var (
	// Views
	ViewGetBlockInfo = coreutil.NewViewEP12(Contract, "getBlockInfo",
		coreutil.FieldWithCodecOptional(codec.Uint32),
		coreutil.FieldWithCodec(codec.Uint32),
		coreutil.FieldWithCodec(codec.NewCodecEx(BlockInfoFromBytes)),
	)
	ViewGetRequestIDsForBlock = coreutil.NewViewEP12(Contract, "getRequestIDsForBlock",
		coreutil.FieldWithCodecOptional(codec.Uint32),
		coreutil.FieldWithCodec(codec.Uint32),
		coreutil.FieldArrayWithCodec(codec.RequestID),
	)
	ViewGetRequestReceipt = coreutil.NewViewEP11(Contract, "getRequestReceipt",
		coreutil.FieldWithCodec(codec.RequestID),
		OutputRequestReceipt{},
	)
	ViewGetRequestReceiptsForBlock = coreutil.NewViewEP11(Contract, "getRequestReceiptsForBlock",
		coreutil.FieldWithCodecOptional(codec.Uint32),
		OutputRequestReceipts{},
	)
	ViewIsRequestProcessed = coreutil.NewViewEP11(Contract, "isRequestProcessed",
		coreutil.FieldWithCodec(codec.RequestID),
		coreutil.FieldWithCodec(codec.Bool),
	)
	ViewGetEventsForRequest = coreutil.NewViewEP11(Contract, "getEventsForRequest",
		coreutil.FieldWithCodec(codec.RequestID),
		coreutil.FieldArrayWithCodec(codec.NewCodecEx(isc.EventFromBytes)),
	)
	ViewGetEventsForBlock = coreutil.NewViewEP12(Contract, "getEventsForBlock",
		coreutil.FieldWithCodecOptional(codec.Uint32),
		coreutil.FieldWithCodec(codec.Uint32),
		coreutil.FieldArrayWithCodec(codec.NewCodecEx(isc.EventFromBytes)),
	)
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
)

type OutputRequestReceipt struct{}

func (OutputRequestReceipt) Encode(rec *RequestReceipt) []byte {
	if rec == nil {
		return nil
	}
	ww := rwutil.NewBytesWriter()
	ww.WriteUint32(rec.BlockIndex)
	ww.WriteUint16(rec.RequestIndex)
	ww.Write(rec)
	if ww.Err != nil {
		panic(ww.Err)
	}
	return ww.Bytes()
}

func (OutputRequestReceipt) Decode(r []byte) (*RequestReceipt, error) {
	if len(r) == 0 {
		return nil, nil
	}
	rr := rwutil.NewBytesReader(r)
	blockIndex := rr.ReadUint32()
	reqIndex := rr.ReadUint16()
	if rr.Err != nil {
		return nil, rr.Err
	}
	rec, err := RequestReceiptFromBytes(rr.Bytes(), blockIndex, reqIndex)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

type RequestReceiptsResponse struct {
	BlockIndex uint32
	Receipts   []*RequestReceipt
}

type OutputRequestReceipts struct{}

func (OutputRequestReceipts) Encode(res *RequestReceiptsResponse) []byte {
	ww := rwutil.NewBytesWriter()
	ww.WriteUint32(res.BlockIndex)
	ww.WriteUint16(uint16(len(res.Receipts)))
	for _, receipt := range res.Receipts {
		ww.Write(receipt)
	}
	if ww.Err != nil {
		panic(ww.Err)
	}
	return ww.Bytes()
}

func (OutputRequestReceipts) Decode(r []byte) (*RequestReceiptsResponse, error) {
	rr := rwutil.NewBytesReader(r)
	blockIndex := rr.ReadUint32()
	n := rr.ReadUint16()
	ret := make([]*RequestReceipt, n)
	for i := uint16(0); i < n; i++ {
		rec := new(RequestReceipt)
		rr.Read(rec)
		rec.BlockIndex = blockIndex
		rec.RequestIndex = i
	}
	return &RequestReceiptsResponse{
		BlockIndex: blockIndex,
		Receipts:   ret,
	}, nil
}
