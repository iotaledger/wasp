package blocklog

import (
	"errors"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// region RequestReceipt /////////////////////////////////////////////////////

// RequestReceipt represents log record of processed request on the chain
type RequestReceipt struct {
	Request       isc.Request            `json:"request"`
	Error         *isc.UnresolvedVMError `json:"error"`
	GasBudget     uint64                 `json:"gasBudget"`
	GasBurned     uint64                 `json:"gasBurned"`
	GasFeeCharged uint64                 `json:"gasFeeCharged"`
	SDCharged     uint64                 `json:"storageDepositCharged"`
	// not persistent
	BlockIndex   uint32       `json:"blockIndex"`
	RequestIndex uint16       `json:"requestIndex"`
	GasBurnLog   *gas.BurnLog `json:"-"`
}

func RequestReceiptFromBytes(data []byte, blockIndex uint32, reqIndex uint16) (*RequestReceipt, error) {
	rec, err := rwutil.ReadFromBytes(data, new(RequestReceipt))
	if err != nil {
		return nil, err
	}
	rec.BlockIndex = blockIndex
	rec.RequestIndex = reqIndex
	return rec, nil
}

func RequestReceiptsFromBlock(block state.Block) ([]*RequestReceipt, error) {
	receipts := []*RequestReceipt{}
	partition := subrealm.NewReadOnly(block.MutationsReader(), kv.Key(Contract.Hname().Bytes()))

	blockInfo, ok := GetBlockInfo(partition, block.StateIndex())
	if !ok {
		return nil, errors.New("inconsistency: BlockInfo not found in block mutations")
	}
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		recBin, found := getRequestRecordDataByRef(partition, block.StateIndex(), reqIdx)
		if !found {
			return nil, errors.New("inconsistency: request log record wasn't found by exact reference")
		}
		receipt, err := RequestReceiptFromBytes(recBin, block.StateIndex(), reqIdx)
		if err != nil {
			return nil, fmt.Errorf("cannot deserialize requestReceipt: %w", err)
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func (rec *RequestReceipt) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func (rec *RequestReceipt) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rec.GasBudget = rr.ReadGas64()
	rec.GasBurned = rr.ReadGas64()
	rec.GasFeeCharged = rr.ReadGas64()
	rec.SDCharged = rr.ReadAmount64()
	rec.Request = isc.RequestFromReader(rr)
	hasError := rr.ReadBool()
	if hasError {
		rec.Error = new(isc.UnresolvedVMError)
		rr.Read(rec.Error)
	}
	if len(rr.Bytes()) != 0 {
		rec.GasBurnLog = new(gas.BurnLog)
		rr.Read(rec.GasBurnLog)
	}

	return rr.Err
}

func (rec *RequestReceipt) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteGas64(rec.GasBudget)
	ww.WriteGas64(rec.GasBurned)
	ww.WriteGas64(rec.GasFeeCharged)
	ww.WriteAmount64(rec.SDCharged)
	ww.Write(rec.Request)
	ww.WriteBool(rec.Error != nil)
	if rec.Error != nil {
		ww.Write(rec.Error)
	}
	if rec.GasBurnLog != nil {
		ww.Write(rec.GasBurnLog)
	}
	return ww.Err
}

func (rec *RequestReceipt) String() string {
	ret := fmt.Sprintf("ID: %s\n", rec.Request.ID().String())
	ret += fmt.Sprintf("Err: %v\n", rec.Error)
	ret += fmt.Sprintf("Block/Request index: %d / %d\n", rec.BlockIndex, rec.RequestIndex)
	ret += fmt.Sprintf("Gas budget / burned / fee charged: %d / %d /%d\n", rec.GasBudget, rec.GasBurned, rec.GasFeeCharged)
	ret += fmt.Sprintf("Storage deposit charged: %d\n", rec.SDCharged)
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

// RequestLookupReference globally unique reference to the request: block index and index of the request within block
type RequestLookupKey [6]byte

func NewRequestLookupKey(blockIndex uint32, requestIndex uint16) RequestLookupKey {
	ret := RequestLookupKey{}
	copy(ret[:4], codec.EncodeUint32(blockIndex))
	copy(ret[4:6], codec.EncodeUint16(requestIndex))
	return ret
}

func (k RequestLookupKey) BlockIndex() uint32 {
	return codec.MustDecodeUint32(k[:4])
}

func (k RequestLookupKey) RequestIndex() uint16 {
	return codec.MustDecodeUint16(k[4:6])
}

func (k RequestLookupKey) Bytes() []byte {
	return k[:]
}

func (k *RequestLookupKey) Read(r io.Reader) error {
	return rwutil.ReadN(r, k[:])
}

func (k *RequestLookupKey) Write(w io.Writer) error {
	return rwutil.WriteN(w, k[:])
}

// endregion ///////////////////////////////////////////////////////////

// region RequestLookupKeyList //////////////////////////////////////////////

// RequestLookupKeyList a list of RequestLookupReference of requests with colliding isc.RequestLookupDigest
type RequestLookupKeyList []RequestLookupKey

func RequestLookupKeyListFromBytes(data []byte) (ret RequestLookupKeyList, err error) {
	rr := rwutil.NewBytesReader(data)
	size := rr.ReadSize16()
	ret = make(RequestLookupKeyList, size)
	for i := range ret {
		rr.Read(&ret[i])
	}
	return ret, rr.Err
}

func (ll RequestLookupKeyList) Bytes() []byte {
	ww := rwutil.NewBytesWriter()
	ww.WriteSize16(len(ll))
	for i := range ll {
		ww.Write(&ll[i])
	}
	return ww.Bytes()
}

// endregion /////////////////////////////////////////////////////////////
