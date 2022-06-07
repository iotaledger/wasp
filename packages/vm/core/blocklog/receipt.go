package blocklog

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// region RequestReceipt /////////////////////////////////////////////////////

// RequestReceipt represents log record of processed request on the chain
type RequestReceipt struct {
	// TODO request may be big (blobs). Do we want to store it all?
	Request       iscp.Request            `json:"request"`
	Error         *iscp.UnresolvedVMError `json:"error"`
	GasBudget     uint64                  `json:"gasBudget"`
	GasBurned     uint64                  `json:"gasBurned"`
	GasFeeCharged uint64                  `json:"gasFeeCharged"`
	// not persistent
	BlockIndex   uint32       `json:"blockIndex"`
	RequestIndex uint16       `json:"requestIndex"`
	GasBurnLog   *gas.BurnLog `json:"-"`
}

func RequestReceiptFromBytes(data []byte) (*RequestReceipt, error) {
	return RequestReceiptFromMarshalUtil(marshalutil.New(data))
}

func RequestReceiptFromMarshalUtil(mu *marshalutil.MarshalUtil) (*RequestReceipt, error) {
	ret := &RequestReceipt{}

	var err error

	if ret.GasBudget, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.GasBurned, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.GasFeeCharged, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.Request, err = iscp.NewRequestFromMarshalUtil(mu); err != nil {
		return nil, err
	}

	if isError, err := mu.ReadBool(); err != nil {
		return nil, err
	} else if !isError {
		return ret, nil
	}

	if ret.Error, err = iscp.UnresolvedVMErrorFromMarshalUtil(mu); err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *RequestReceipt) Bytes() []byte {
	mu := marshalutil.New()

	mu.WriteUint64(r.GasBudget).
		WriteUint64(r.GasBurned).
		WriteUint64(r.GasFeeCharged)

	r.Request.WriteToMarshalUtil(mu)

	if r.Error == nil {
		mu.WriteBool(false)
	} else {
		mu.WriteBool(true)
		mu.WriteBytes(r.Error.Bytes())
	}

	return mu.Bytes()
}

func (r *RequestReceipt) WithBlockData(blockIndex uint32, requestIndex uint16) *RequestReceipt {
	r.BlockIndex = blockIndex
	r.RequestIndex = requestIndex
	return r
}

func (r *RequestReceipt) String() string {
	ret := fmt.Sprintf("ID: %s\n", r.Request.ID().String())
	ret += fmt.Sprintf("Err: %v\n", r.Error)
	ret += fmt.Sprintf("Block/Request index: %d / %d\n", r.BlockIndex, r.RequestIndex)
	ret += fmt.Sprintf("Gas budget / burned / fee charged: %d / %d /%d\n", r.GasBudget, r.GasBurned, r.GasFeeCharged)
	ret += fmt.Sprintf("Call data: %s\n", r.Request)
	return ret
}

func (r *RequestReceipt) Short() string {
	prefix := "tx"
	if r.Request.IsOffLedger() {
		prefix = "api"
	}

	ret := fmt.Sprintf("%s/%s", prefix, r.Request.ID())

	if r.Error != nil {
		ret += fmt.Sprintf(": Err: %v", r.Error)
	}

	return ret
}

func (r *RequestReceipt) LookupKey() RequestLookupKey {
	return NewRequestLookupKey(r.BlockIndex, r.RequestIndex)
}

// endregion  /////////////////////////////////////////////////////////////

// region RequestLookupKey /////////////////////////////////////////////

// RequestLookupReference globally unique reference to the request: block index and index of the request within block
type RequestLookupKey [6]byte

func NewRequestLookupKey(blockIndex uint32, requestIndex uint16) RequestLookupKey {
	ret := RequestLookupKey{}
	copy(ret[:4], util.Uint32To4Bytes(blockIndex))
	copy(ret[4:6], util.Uint16To2Bytes(requestIndex))
	return ret
}

func (k RequestLookupKey) BlockIndex() uint32 {
	return util.MustUint32From4Bytes(k[:4])
}

func (k RequestLookupKey) RequestIndex() uint16 {
	return util.MustUint16From2Bytes(k[4:6])
}

func (k RequestLookupKey) Bytes() []byte {
	return k[:]
}

func (k *RequestLookupKey) Write(w io.Writer) error {
	_, err := w.Write(k[:])
	return err
}

func (k *RequestLookupKey) Read(r io.Reader) error {
	n, err := r.Read(k[:])
	if err != nil || n != 6 {
		return io.EOF
	}
	return nil
}

// endregion ///////////////////////////////////////////////////////////

// region RequestLookupKeyList //////////////////////////////////////////////

// RequestLookupKeyList a list of RequestLookupReference of requests with colliding iscp.RequestLookupDigest
type RequestLookupKeyList []RequestLookupKey

func RequestLookupKeyListFromBytes(data []byte) (RequestLookupKeyList, error) {
	rdr := bytes.NewReader(data)
	var size uint16
	if err := util.ReadUint16(rdr, &size); err != nil {
		return nil, err
	}
	ret := make(RequestLookupKeyList, size)
	for i := uint16(0); i < size; i++ {
		if err := ret[i].Read(rdr); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (ll RequestLookupKeyList) Bytes() []byte {
	if len(ll) > math.MaxUint16 {
		panic("RequestLookupKeyList::Write: too long")
	}
	var buf bytes.Buffer
	_ = util.WriteUint16(&buf, uint16(len(ll)))
	for i := range ll {
		_ = ll[i].Write(&buf)
	}
	return buf.Bytes()
}

// endregion /////////////////////////////////////////////////////////////
