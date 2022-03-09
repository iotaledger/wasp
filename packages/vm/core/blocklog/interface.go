// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/state"
	"io"
	"math"
	"time"

	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
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
	BlockIndex              uint32 // not persistent. Set from key
	Timestamp               time.Time
	TotalRequests           uint16
	NumSuccessfulRequests   uint16
	NumOffLedgerRequests    uint16
	PreviousStateCommitment trie.VCommitment
	StateCommitment         trie.VCommitment // nil if not known
	AnchorTransactionID     iotago.TransactionID
	TotalIotasInL2Accounts  uint64
	TotalDustDeposit        uint64
	GasBurned               uint64
	GasFeeCharged           uint64
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
	ret += fmt.Sprintf("Timestamp: %d\n", bi.Timestamp.UnixNano())
	ret += fmt.Sprintf("Total requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("off-ledger requests: %d\n", bi.NumOffLedgerRequests)
	ret += fmt.Sprintf("Succesfull requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("Prev state hash: %s\n", bi.PreviousStateCommitment.String())
	ret += fmt.Sprintf("Anchor tx ID: %s\n", hex.EncodeToString(bi.AnchorTransactionID[:]))
	ret += fmt.Sprintf("Total iotas in contracts: %d\n", bi.TotalIotasInL2Accounts)
	ret += fmt.Sprintf("Total iotas locked in dust deposit: %d\n", bi.TotalDustDeposit)
	ret += fmt.Sprintf("Gas burned: %d\n", bi.GasBurned)
	ret += fmt.Sprintf("Gas fee charged: %d\n", bi.GasFeeCharged)
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
	if _, err := w.Write(bi.PreviousStateCommitment.Bytes()); err != nil {
		return err
	}
	if err := util.WriteBoolByte(w, bi.StateCommitment != nil); err != nil {
		return err
	}
	if bi.StateCommitment != nil {
		if _, err := w.Write(bi.StateCommitment.Bytes()); err != nil {
			return err
		}
	}
	if err := util.WriteUint64(w, bi.TotalIotasInL2Accounts); err != nil {
		return err
	}
	if err := util.WriteUint64(w, bi.TotalDustDeposit); err != nil {
		return err
	}
	if err := util.WriteUint64(w, bi.GasBurned); err != nil {
		return err
	}
	if err := util.WriteUint64(w, bi.GasFeeCharged); err != nil {
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
	bi.PreviousStateCommitment = state.CommitmentModel.NewVectorCommitment()
	if err := bi.PreviousStateCommitment.Read(r); err != nil {
		return err
	}
	var knownStateCommitments bool
	if err := util.ReadBoolByte(r, &knownStateCommitments); err != nil {
		return err
	}
	bi.StateCommitment = nil
	if knownStateCommitments {
		bi.StateCommitment = state.CommitmentModel.NewVectorCommitment()
		if err := bi.StateCommitment.Read(r); err != nil {
			return err
		}
	}
	if err := util.ReadUint64(r, &bi.TotalIotasInL2Accounts); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &bi.TotalDustDeposit); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &bi.GasBurned); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &bi.GasFeeCharged); err != nil {
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

// region RequestReceipt /////////////////////////////////////////////////////

// RequestReceipt represents log record of processed request on the chain
type RequestReceipt struct {
	Request       iscp.Request // TODO request may be big (blobs). Do we want to store it all?
	Error         *iscp.UnresolvedVMError
	GasBudget     uint64
	GasBurned     uint64
	GasFeeCharged uint64
	// not persistent
	BlockIndex   uint32
	RequestIndex uint16
	GasBurnLog   *gas.BurnLog
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
	if ret.Request, err = iscp.RequestDataFromMarshalUtil(mu); err != nil {
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

	iscp.RequestDataToMarshalUtil(r.Request, mu)

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
	ret += fmt.Sprintf("Call data: %s\n", r.Request.String())
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

func BlockInfoKey(index uint32) []byte {
	return []byte(collections.Array32ElemKey(prefixBlockRegistry, index))
}
