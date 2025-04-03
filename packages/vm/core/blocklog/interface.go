// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"
	"reflect"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog)

var (
	// Views
	ViewGetBlockInfo = coreutil.NewViewEP12(Contract, "getBlockInfo",
		coreutil.FieldOptional[uint32]("blockIndex"),
		coreutil.Field[uint32]("blockIndex"),
		coreutil.Field[*BlockInfo]("blockInfo"),
	)
	ViewGetRequestIDsForBlock = coreutil.NewViewEP12(Contract, "getRequestIDsForBlock",
		coreutil.FieldOptional[uint32]("blockIndex"),
		coreutil.Field[uint32]("blockIndex"),
		coreutil.Field[[]isc.RequestID]("requestIDsInBlock"),
	)
	ViewGetRequestReceipt = coreutil.NewViewEP11(Contract, "getRequestReceipt",
		coreutil.Field[isc.RequestID]("requestID"),
		OutputRequestReceipt{},
	)
	ViewGetRequestReceiptsForBlock = coreutil.NewViewEP11(Contract, "getRequestReceiptsForBlock",
		coreutil.FieldOptional[uint32]("blockIndex"),
		OutputRequestReceipts{},
	)
	ViewIsRequestProcessed = coreutil.NewViewEP11(Contract, "isRequestProcessed",
		coreutil.Field[isc.RequestID]("requestID"),
		coreutil.Field[bool]("isProcessed"),
	)
	ViewGetEventsForRequest = coreutil.NewViewEP11(Contract, "getEventsForRequest",
		coreutil.Field[isc.RequestID]("requestID"),
		coreutil.Field[[]*isc.Event]("events"),
	)
	ViewGetEventsForBlock = coreutil.NewViewEP12(Contract, "getEventsForBlock",
		coreutil.FieldOptional[uint32]("blockIndex"),
		coreutil.Field[uint32]("blockIndex"),
		coreutil.Field[[]*isc.Event]("events"),
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

func (OutputRequestReceipt) Name() string {
	return "requestReceipt"
}

func (OutputRequestReceipt) Type() reflect.Type {
	return reflect.TypeOf(OutputRequestReceipt{})
}

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

func (OutputRequestReceipts) Name() string {
	return "requestReceipts"
}

func (OutputRequestReceipts) Type() reflect.Type {
	return reflect.TypeOf(RequestReceiptsResponse{})
}

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
