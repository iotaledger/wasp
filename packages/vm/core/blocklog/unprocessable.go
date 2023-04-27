package blocklog

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type unprocessableRequestRecord struct {
	blockIndex  uint32
	outputIndex uint16
	req         isc.Request
}

func (r *unprocessableRequestRecord) bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint32(r.blockIndex)
	mu.WriteUint16(r.outputIndex)
	r.req.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func unprocessableRequestRecordFromBytes(data []byte) (*unprocessableRequestRecord, error) {
	ret := &unprocessableRequestRecord{}
	mu := marshalutil.New(data)
	var err error
	ret.blockIndex, err = mu.ReadUint32()
	if err != nil {
		return nil, err
	}
	ret.outputIndex, err = mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	ret.req, err = isc.NewRequestFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func unprocessableMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixUnprocessableRequests)
}

func unprocessableMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixUnprocessableRequests)
}

// save request reference / address of the sender
func SaveUnprocessable(state kv.KVStore, req isc.OnLedgerRequest, blockIndex uint32, outputIndex uint16) {
	// TODO check if REQID already exists, return an error if it does?
	// save in a map of reqID => {isc.Request, blockIndex, outputIndex}
	rec := unprocessableRequestRecord{
		blockIndex:  blockIndex,
		outputIndex: outputIndex,
		req:         req,
	}
	unprocessableMap(state).SetAt(req.ID().Bytes(), rec.bytes())
}

func GetUnprocessable(state kv.KVStoreReader, reqID isc.RequestID) (req isc.Request, blockIndex uint32, outputIndex uint16, err error) {
	recData := unprocessableMapR(state).GetAt(reqID.Bytes())
	rec, err := unprocessableRequestRecordFromBytes(recData)
	if err != nil {
		return nil, 0, 0, err
	}
	return rec.req, rec.blockIndex, rec.outputIndex, nil
}

func HasUnprocessable(state kv.KVStoreReader, reqID isc.RequestID) bool {
	return unprocessableMapR(state).HasAt(reqID.Bytes())
}

func RemoveUnprocessable(state kv.KVStore, reqID isc.RequestID) {
	unprocessableMap(state).DelAt(reqID.Bytes())
}

// ---- entrypoints

// view used to check if a given requestID exists on the unprocessable list
func viewHasUnprocessable(ctx isc.SandboxView) dict.Dict {
	reqID := ctx.Params().MustGetRequestID(ParamRequestID)
	exists := HasUnprocessable(ctx.StateR(), reqID)
	return dict.Dict{
		ParamUnprocessableRequestExists: codec.EncodeBool(exists),
	}
}

func retryUnprocessable(ctx isc.Sandbox) dict.Dict {
	reqID := ctx.Params().MustGetRequestID(ParamRequestID)
	exists := HasUnprocessable(ctx.StateR(), reqID)
	// TODO add VMERRORS for these errors?
	if !exists {
		panic("request does not exist on the unprocessable list")
	}
	rec, blockIndex, outputIndex, err := GetUnprocessable(ctx.StateR(), reqID)
	if err != nil {
		panic(fmt.Errorf("error getting unprocessable request: %s", err.Error()))
	}
	if !rec.SenderAccount().Equals(ctx.Request().SenderAccount()) {
		panic("unprocessable request sender does not match the retry sender")
	}
	ctx.Privileged().RetryUnprocessable(rec, blockIndex, outputIndex)
	return nil
}
