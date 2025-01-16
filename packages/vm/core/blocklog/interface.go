// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog)

var (
	// Views
	ViewGetBlockInfo = coreutil.NewViewEP12(Contract, "getBlockInfo",
		coreutil.FieldOptional[uint32](),
		coreutil.Field[uint32](),
		coreutil.Field[*BlockInfo](),
	)
	ViewGetRequestIDsForBlock = coreutil.NewViewEP12(Contract, "getRequestIDsForBlock",
		coreutil.FieldOptional[uint32](),
		coreutil.Field[uint32](),
		coreutil.Field[[]isc.RequestID](),
	)
	ViewGetRequestReceipt = coreutil.NewViewEP11(Contract, "getRequestReceipt",
		coreutil.Field[isc.RequestID](),
		OutputRequestReceipt{},
	)
	ViewGetRequestReceiptsForBlock = coreutil.NewViewEP11(Contract, "getRequestReceiptsForBlock",
		coreutil.FieldOptional[uint32](),
		OutputRequestReceipts{},
	)
	ViewIsRequestProcessed = coreutil.NewViewEP11(Contract, "isRequestProcessed",
		coreutil.Field[isc.RequestID](),
		coreutil.Field[bool](),
	)
	ViewGetEventsForRequest = coreutil.NewViewEP11(Contract, "getEventsForRequest",
		coreutil.Field[isc.RequestID](),
		coreutil.Field[[]*isc.Event](),
	)
	ViewGetEventsForBlock = coreutil.NewViewEP12(Contract, "getEventsForBlock",
		coreutil.FieldOptional[uint32](),
		coreutil.Field[uint32](),
		coreutil.Field[[]*isc.Event](),
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

	var buf bytes.Buffer
	enc := bcs.NewEncoder(&buf)

	enc.MustEncode(rec.BlockIndex)
	enc.MustEncode(rec.RequestIndex)
	enc.MustEncode(rec)

	return buf.Bytes()
}

func (OutputRequestReceipt) Decode(r []byte) (*RequestReceipt, error) {
	if len(r) == 0 {
		return nil, nil
	}

	rr := bytes.NewReader(r)
	dec := bcs.NewDecoder(rr)

	blockIndex := bcs.Decode[uint32](dec)
	reqIndex := bcs.Decode[uint16](dec)

	if dec.Err() != nil {
		return nil, dec.Err()
	}

	rec, err := RequestReceiptFromReader(rr, blockIndex, reqIndex)
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
	return bcs.MustMarshal(res)
}

func (OutputRequestReceipts) Decode(r []byte) (*RequestReceiptsResponse, error) {
	res, err := bcs.Unmarshal[*RequestReceiptsResponse](r)
	if err != nil {
		return nil, err
	}

	for i := range res.Receipts {
		res.Receipts[i].BlockIndex = res.BlockIndex
		res.Receipts[i].RequestIndex = uint16(i)
	}

	return res, nil
}
