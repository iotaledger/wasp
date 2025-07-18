package blocklog

import (
	"bytes"
	"fmt"
	"io"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

// region RequestReceipt /////////////////////////////////////////////////////

// RequestReceipt represents log record of processed request on the chain
type RequestReceipt struct {
	Request       isc.Request            `json:"request"`
	Error         *isc.UnresolvedVMError `json:"error" bcs:"optional"`
	GasBudget     uint64                 `json:"gasBudget" bcs:"compact"`
	GasBurned     uint64                 `json:"gasBurned" bcs:"compact"`
	GasFeeCharged coin.Value             `json:"gasFeeCharged" bcs:"compact"`
	GasBurnLog    *gas.BurnLog           `json:"-" bcs:"optional"`
	// not persistent
	BlockIndex   uint32 `json:"blockIndex" bcs:"-"`
	RequestIndex uint16 `json:"requestIndex" bcs:"-"`
}

func RequestReceiptFromBytes(data []byte, blockIndex uint32, reqIndex uint16) (*RequestReceipt, error) {
	return RequestReceiptFromReader(bytes.NewReader(data), blockIndex, reqIndex)
}

func RequestReceiptFromReader(r io.Reader, blockIndex uint32, reqIndex uint16) (*RequestReceipt, error) {
	rec, err := bcs.UnmarshalStream[*RequestReceipt](r)
	if err != nil {
		return nil, err
	}
	rec.BlockIndex = blockIndex
	rec.RequestIndex = reqIndex
	return rec, nil
}

func RequestReceiptsFromBlock(block state.Block) ([]*RequestReceipt, error) {
	state := NewStateReaderFromBlockMutations(block)
	_, recs, err := state.GetRequestReceiptsInBlock(block.StateIndex())
	return recs, err
}

func (rec *RequestReceipt) Bytes() []byte {
	return bcs.MustMarshal(rec)
}

func (rec *RequestReceipt) String() string {
	ret := fmt.Sprintf("ID: %s\n", rec.Request.ID().String())
	ret += fmt.Sprintf("Err: %v\n", rec.Error)
	ret += fmt.Sprintf("Block/Request index: %d / %d\n", rec.BlockIndex, rec.RequestIndex)
	ret += fmt.Sprintf("Gas budget / burned / fee charged: %d / %d /%d\n", rec.GasBudget, rec.GasBurned, rec.GasFeeCharged)
	ret += fmt.Sprintf("Call data: %s\n", rec.Request)
	ret += fmt.Sprintf("burn log: %s\n", rec.GasBurnLog)
	return ret
}

func (rec *RequestReceipt) Short() string {
	prefix := "tx"
	if rec.Request.IsOffLedger() {
		prefix = "api"
	}

	ret := fmt.Sprintf("%s/%s", prefix, rec.Request.ID())

	if rec.Error != nil {
		ret += fmt.Sprintf(": Err: %v", rec.Error)
	}

	return ret
}

func (rec *RequestReceipt) LookupKey() RequestLookupKey {
	return NewRequestLookupKey(rec.BlockIndex, rec.RequestIndex)
}

func (rec *RequestReceipt) ToISCReceipt(resolvedError *isc.VMError) *isc.Receipt {
	return &isc.Receipt{
		Request:       rec.Request.Bytes(),
		Error:         rec.Error,
		GasBudget:     rec.GasBudget,
		GasBurned:     rec.GasBurned,
		GasFeeCharged: rec.GasFeeCharged,
		BlockIndex:    rec.BlockIndex,
		RequestIndex:  rec.RequestIndex,
		ResolvedError: resolvedError.Error(),
		GasBurnLog:    rec.GasBurnLog,
	}
}

// endregion  /////////////////////////////////////////////////////////////

// region RequestLookupKey /////////////////////////////////////////////

// RequestLookupKey globally unique reference to the request: block index and index of the request within block
type RequestLookupKey [6]byte

func NewRequestLookupKey(blockIndex uint32, requestIndex uint16) RequestLookupKey {
	ret := RequestLookupKey{}
	copy(ret[:4], codec.Encode[uint32](blockIndex))
	copy(ret[4:6], codec.Encode[uint16](requestIndex))
	return ret
}

func (k *RequestLookupKey) BlockIndex() uint32 {
	return codec.MustDecode[uint32](k[:4])
}

func (k *RequestLookupKey) RequestIndex() uint16 {
	return codec.MustDecode[uint16](k[4:6])
}

func (k *RequestLookupKey) Bytes() []byte {
	return k[:]
}

// endregion ///////////////////////////////////////////////////////////

// region RequestLookupKeyList //////////////////////////////////////////////

// RequestLookupKeyList a list of RequestLookupReference of requests with colliding isc.RequestLookupDigest
type RequestLookupKeyList []RequestLookupKey

func RequestLookupKeyListFromBytes(data []byte) (ret RequestLookupKeyList, err error) {
	return bcs.Unmarshal[RequestLookupKeyList](data)
}

func (ll RequestLookupKeyList) Bytes() []byte {
	return bcs.MustMarshal(&ll)
}

// endregion /////////////////////////////////////////////////////////////
