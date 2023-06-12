package blocklog

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
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
	ret.req, err = isc.RequestFromMarshalUtil(mu)
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

var (
	ErrUnprocessableAlreadyExist = coreerrors.Register("request does not exist on the unprocessable list").Create()
	ErrUnprocessableUnexpected   = coreerrors.Register("unexpected error getting unprocessable request from the state").Create()
	ErrUnprocessableWrongSender  = coreerrors.Register("unprocessable request sender does not match the retry sender").Create()
)

func retryUnprocessable(ctx isc.Sandbox) dict.Dict {
	reqID := ctx.Params().MustGetRequestID(ParamRequestID)
	exists := HasUnprocessable(ctx.StateR(), reqID)
	if !exists {
		panic(ErrUnprocessableAlreadyExist)
	}
	rec, blockIndex, outputIndex, err := GetUnprocessable(ctx.StateR(), reqID)
	if err != nil {
		panic(ErrUnprocessableUnexpected)
	}
	if !rec.SenderAccount().Equals(ctx.Request().SenderAccount()) {
		panic(ErrUnprocessableWrongSender)
	}
	ctx.Privileged().RetryUnprocessable(rec, blockIndex, outputIndex)
	return nil
}

func UnprocessableRequestsAddedInBlock(block state.Block) ([]isc.Request, error) {
	var respErr error
	requests := []isc.Request{}
	kvStore := subrealm.NewReadOnly(block.MutationsReader(), kv.Key(Contract.Hname().Bytes()))
	unprocessableMapR(kvStore).Iterate(func(_, recData []byte) bool {
		rec, err := unprocessableRequestRecordFromBytes(recData)
		if err != nil {
			respErr = err
			return false
		}
		requests = append(requests, rec.req)
		return true
	})
	return requests, respErr
}

func HasUnprocessableRequestBeenRemovedInBlock(block state.Block, requestID isc.RequestID) bool {
	keyBytes := Contract.Hname().Bytes()
	keyBytes = append(keyBytes, collections.MapElemKey(prefixUnprocessableRequests, requestID.Bytes())...)
	_, wasRemoved := block.Mutations().Dels[kv.Key(keyBytes)]
	return wasRemoved
}
