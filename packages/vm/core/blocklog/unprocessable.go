package blocklog

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

type unprocessableRequestRecord struct {
	outputID iotago.OutputID
	req      isc.Request
}

func unprocessableRequestRecordFromBytes(data []byte) (*unprocessableRequestRecord, error) {
	return rwutil.ReadFromBytes(data, new(unprocessableRequestRecord))
}

func mustUnprocessableRequestRecordFromBytes(data []byte) *unprocessableRequestRecord {
	rec, err := unprocessableRequestRecordFromBytes(data)
	if err != nil {
		panic(err)
	}
	return rec
}

func (rec *unprocessableRequestRecord) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func (rec *unprocessableRequestRecord) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(rec.outputID[:])
	rec.req = isc.RequestFromReader(rr)
	return rr.Err
}

func (rec *unprocessableRequestRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(rec.outputID[:])
	ww.Write(rec.req)
	return ww.Err
}

func newUnprocessableRequestsArray(state kv.KVStore) *collections.Array {
	return collections.NewArray(state, prefixNewUnprocessableRequests)
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
		// TransactionID is unknown yet, will be filled next block
		outputID: iotago.OutputIDFromTransactionIDAndIndex(iotago.TransactionID{}, outputIndex),
		req:      req,
	}
	unprocessableMap(state).SetAt(req.ID().Bytes(), rec.Bytes())
	newUnprocessableRequestsArray(state).Push(req.ID().Bytes())
}

func updateUnprocessableRequestsOutputID(state kv.KVStore, anchorTxID iotago.TransactionID) {
	newReqs := newUnprocessableRequestsArray(state)
	allReqs := unprocessableMap(state)
	n := newReqs.Len()
	for i := uint32(0); i < n; i++ {
		k := newReqs.GetAt(i)
		rec := mustUnprocessableRequestRecordFromBytes(allReqs.GetAt(k))
		rec.outputID = iotago.OutputIDFromTransactionIDAndIndex(anchorTxID, rec.outputID.Index())
		allReqs.SetAt(k, rec.Bytes())
	}
	newReqs.Erase()
}

func GetUnprocessable(state kv.KVStoreReader, reqID isc.RequestID) (req isc.Request, outputID iotago.OutputID, err error) {
	recData := unprocessableMapR(state).GetAt(reqID.Bytes())
	rec, err := unprocessableRequestRecordFromBytes(recData)
	if err != nil {
		return nil, iotago.OutputID{}, err
	}
	return rec.req, rec.outputID, nil
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
	rec, outputID, err := GetUnprocessable(ctx.StateR(), reqID)
	if err != nil {
		panic(ErrUnprocessableUnexpected)
	}
	recSender := rec.SenderAccount()
	if rec.SenderAccount() == nil || !recSender.Equals(ctx.Request().SenderAccount()) {
		panic(ErrUnprocessableWrongSender)
	}
	ctx.Privileged().RetryUnprocessable(rec, outputID)
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
