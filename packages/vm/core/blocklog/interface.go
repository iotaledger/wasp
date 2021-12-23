// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/util"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlocklog, "Block log contract")

const (
	prefixBlockRegistry = string('a' + iota)
	prefixControlAddresses
	prefixRequestLookupIndex
	prefixRequestReceipts
	prefixRequestEvents
	prefixSmartContractEventsLookup
)

var (
	FuncControlAddresses           = coreutil.ViewFunc("controlAddresses")
	FuncGetBlockInfo               = coreutil.ViewFunc("getBlockInfo")
	FuncGetLatestBlockInfo         = coreutil.ViewFunc("getLatestBlockInfo")
	FuncGetRequestIDsForBlock      = coreutil.ViewFunc("getRequestIDsForBlock")
	FuncGetRequestReceipt          = coreutil.ViewFunc("getRequestReceipt")
	FuncGetRequestReceiptsForBlock = coreutil.ViewFunc("getRequestReceiptsForBlock")
	FuncIsRequestProcessed         = coreutil.ViewFunc("isRequestProcessed")
	FuncGetEventsForRequest        = coreutil.ViewFunc("getEventsForRequest")
	FuncGetEventsForBlock          = coreutil.ViewFunc("getEventsForBlock")
	FuncGetEventsForContract       = coreutil.ViewFunc("getEventsForContract")
)

const (
	// parameters
	ParamBlockIndex             = "n"
	ParamBlockInfo              = "i"
	ParamGoverningAddress       = "g"
	ParamContractHname          = "h"
	ParamFromBlock              = "f"
	ParamToBlock                = "t"
	ParamRequestID              = "u"
	ParamRequestIndex           = "r"
	ParamRequestProcessed       = "p"
	ParamRequestRecord          = "d"
	ParamEvent                  = "e"
	ParamStateControllerAddress = "s"
)

// region BlockInfo //////////////////////////////////////////////////////////////

type BlockInfo struct {
	BlockIndex               uint32 // not persistent. Set from key
	Timestamp                time.Time
	TotalRequests            uint16
	NumSuccessfulRequests    uint16
	NumOffLedgerRequests     uint16
	PreviousStateHash        hashing.HashValue
	AnchorTransactionID      iotago.TransactionID
	DustDepositAnchor        uint64
	DustDepositNativeTokenID uint64
}

func BlockInfoFromBytes(blockIndex uint32, data []byte) (*BlockInfo, error) {
	ret := &BlockInfo{BlockIndex: blockIndex}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

// RequestTimestamp returns timestamp which corresponds to the request with the given index
// Timestamps of requests are incremented by 1 nanosecond in the block. The timestamp of the last one
// is equal to the timestamp pof the block
func (bi *BlockInfo) RequestTimestamp(requestIndex uint16) time.Time {
	return bi.Timestamp.Add(time.Duration(-(bi.TotalRequests - requestIndex - 1)) * time.Nanosecond)
}

func (bi *BlockInfo) Bytes() []byte {
	var buf bytes.Buffer
	_ = bi.Write(&buf)
	return buf.Bytes()
}

func (bi *BlockInfo) String() string {
	ret := fmt.Sprintf("Block index: %d\n", bi.BlockIndex)
	ret += fmt.Sprintf("OutputTimestamp: %v\n", bi.Timestamp)
	ret += fmt.Sprintf("Total requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("Number of succesfull requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("Number of off-ledger requests: %d\n", bi.NumOffLedgerRequests)
	return ret
}

func (bi *BlockInfo) Write(w io.Writer) error {
	if err := util.WriteTime(w, bi.Timestamp); err != nil {
		return err
	}
	if err := util.WriteUint16(w, bi.TotalRequests); err != nil {
		return err
	}
	if err := util.WriteUint16(w, bi.NumSuccessfulRequests); err != nil {
		return err
	}
	if err := util.WriteUint16(w, bi.NumOffLedgerRequests); err != nil {
		return err
	}
	if _, err := w.Write(bi.AnchorTransactionID[:]); err != nil {
		return err
	}
	if _, err := w.Write(bi.PreviousStateHash.Bytes()); err != nil {
		return err
	}
	if err := util.WriteUint64(w, bi.DustDepositAnchor); err != nil {
		return err
	}
	if err := util.WriteUint64(w, bi.DustDepositNativeTokenID); err != nil {
		return err
	}
	return nil
}

func (bi *BlockInfo) Read(r io.Reader) error {
	if err := util.ReadTime(r, &bi.Timestamp); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &bi.TotalRequests); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &bi.NumSuccessfulRequests); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &bi.NumOffLedgerRequests); err != nil {
		return err
	}
	if err := util.ReadTransactionID(r, &bi.AnchorTransactionID); err != nil {
		return err
	}
	if err := util.ReadHashValue(r, &bi.PreviousStateHash); err != nil { // nolint:nolint
		return err
	}
	if err := util.ReadUint64(r, &bi.DustDepositAnchor); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &bi.DustDepositNativeTokenID); err != nil {
		return err
	}
	return nil
}

// endregion //////////////////////////////////////////////////////////

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

// region RequestLookupKey /////////////////////////////////////////////

// EventLookupKey is a globally unique reference to the event:
// block index + index of the request within block + index of the event within the request
type EventLookupKey [8]byte

func NewEventLookupKey(blockIndex uint32, requestIndex, eventIndex uint16) EventLookupKey {
	ret := EventLookupKey{}
	copy(ret[:4], util.Uint32To4Bytes(blockIndex))
	copy(ret[4:6], util.Uint16To2Bytes(requestIndex))
	copy(ret[6:8], util.Uint16To2Bytes(eventIndex))
	return ret
}

func (k EventLookupKey) BlockIndex() uint32 {
	return util.MustUint32From4Bytes(k[:4])
}

func (k EventLookupKey) RequestIndex() uint16 {
	return util.MustUint16From2Bytes(k[4:6])
}

func (k EventLookupKey) RequestEventIndex() uint16 {
	return util.MustUint16From2Bytes(k[6:8])
}

func (k EventLookupKey) Bytes() []byte {
	return k[:]
}

func (k *EventLookupKey) Write(w io.Writer) error {
	_, err := w.Write(k[:])
	return err
}

func EventLookupKeyFromBytes(r io.Reader) (*EventLookupKey, error) {
	k := EventLookupKey{}
	n, err := r.Read(k[:])
	if err != nil || n != 8 {
		return nil, io.EOF
	}
	return &k, nil
}

// endregion ///////////////////////////////////////////////////////////

// region RequestLogReqcord /////////////////////////////////////////////////////

// RequestReceipt represents log record of processed request on the chain
type RequestReceipt struct {
	RequestData iscp.RequestData
	Error       string
	// not persistent
	BlockIndex   uint32
	RequestIndex uint16
}

func RequestReceiptFromBytes(data []byte) (*RequestReceipt, error) {
	return RequestReceiptFromMarshalUtil(marshalutil.New(data))
}

func RequestReceiptFromMarshalUtil(mu *marshalutil.MarshalUtil) (*RequestReceipt, error) {
	ret := &RequestReceipt{}
	var err error
	if ret.RequestData, err = iscp.RequestDataFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	var size uint16
	if size, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	strBytes, err := mu.ReadBytes(int(size))
	if err != nil {
		return nil, err
	}
	ret.Error = string(strBytes)
	return ret, nil
}

func (r *RequestReceipt) Bytes() []byte {
	mu := marshalutil.New()

	mu.WriteBytes(r.RequestData.Bytes()).
		WriteUint16(uint16(len(r.Error))).
		WriteBytes([]byte(r.Error))
	return mu.Bytes()
}

func (r *RequestReceipt) WithBlockData(blockIndex uint32, requestIndex uint16) *RequestReceipt {
	r.BlockIndex = blockIndex
	r.RequestIndex = requestIndex
	return r
}

func (r *RequestReceipt) String() string {
	if len(r.Error) > 0 {
		return fmt.Sprintf("%s\n Error: '%s'", r.RequestData.String(), r.Error)
	}
	return r.RequestData.String()
}

func (r *RequestReceipt) Short() string {
	prefix := "tx"
	if r.RequestData.IsOffLedger() {
		prefix = "api"
	}
	ret := fmt.Sprintf("%s/%s", prefix, r.RequestData.ID())
	if len(r.Error) > 0 {
		ret += ": '" + r.Error + "'"
	}
	return ret
}

// endregion  /////////////////////////////////////////////////////////////

// region ControlAddresses ///////////////////////////////////////////////

type ControlAddresses struct {
	StateAddress     iotago.Address
	GoverningAddress iotago.Address
	SinceBlockIndex  uint32
}

func ControlAddressesFromBytes(data []byte) (*ControlAddresses, error) {
	return ControlAddressesFromMarshalUtil(marshalutil.New(data))
}

func ControlAddressesFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ControlAddresses, error) {
	ret := &ControlAddresses{}
	var err error

	if ret.StateAddress, err = iscp.AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.GoverningAddress, err = iscp.AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.SinceBlockIndex, err = mu.ReadUint32(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (ca *ControlAddresses) Bytes() []byte {
	mu := marshalutil.New()

	mu.WriteBytes(iscp.BytesFromAddress(ca.StateAddress)).
		WriteBytes(iscp.BytesFromAddress(ca.GoverningAddress)).
		WriteUint32(ca.SinceBlockIndex)
	return mu.Bytes()
}

func (ca *ControlAddresses) String() string {
	var ret string
	if ca.StateAddress.Equal(ca.GoverningAddress) {
		ret = fmt.Sprintf("ControlAddresses(%s), block: %d", ca.StateAddress.Bech32(iscp.Bech32Prefix), ca.SinceBlockIndex)
	} else {
		ret = fmt.Sprintf("ControlAddresses(%s, %s), block: %d",
			ca.StateAddress.Bech32(iscp.Bech32Prefix), ca.GoverningAddress.Bech32(iscp.Bech32Prefix), ca.SinceBlockIndex)
	}
	return ret
}

// endregion /////////////////////////////////////////////////////////////
