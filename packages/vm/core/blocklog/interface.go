// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"time"
)

const (
	Name        = coreutil.CoreContractBlocklog
	description = "Block log contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.ViewFunc(FuncGetBlockInfo, viewGetBlockInfo),
		coreutil.ViewFunc(FuncGetLatestBlockInfo, viewGetLatestBlockInfo),
		coreutil.ViewFunc(FuncGetRequestLogRecord, viewGetRequestLogRecord),
		coreutil.ViewFunc(FuncGetRequestLogRecordsForBlock, viewGetRequestLogRecordsForBlock),
		coreutil.ViewFunc(FuncGetRequestIDsForBlock, viewGetRequestIDsForBlock),
		coreutil.ViewFunc(FuncIsRequestProcessed, viewIsRequestProcessed),
	})
}

const (
	// state variables
	StateVarTimestamp          = coreutil.StateVarTimestamp
	StateVarBlockIndex         = coreutil.StateVarBlockIndex
	StateVarBlockRegistry      = "b"
	StateVarRequestLookupIndex = "l"
	StateVarRequestRecords     = "r"
	// functions
	FuncGetBlockInfo                 = "viewGetBlockInfo"
	FuncGetLatestBlockInfo           = "viewGetLatestBlockInfo"
	FuncGetRequestLogRecord          = "viewGetRequestLogRecord"
	FuncGetRequestLogRecordsForBlock = "viewGetRequestLogRecordsForBlock"
	FuncGetRequestIDsForBlock        = "viewGetRequestIDsForBlock"
	FuncIsRequestProcessed           = "viewIsRequestProcessed"

	// parameters
	ParamBlockIndex       = "n"
	ParamRequestIndex     = "r"
	ParamBlockInfo        = "i"
	ParamRequestRecord    = "d"
	ParamRequestID        = "u"
	ParamRequestProcessed = "p"
)

// region BlockInfo //////////////////////////////////////////////////////////////

type BlockInfo struct {
	BlockIndex            uint32 // not persistent. Set from key
	Timestamp             time.Time
	TotalRequests         uint16
	NumSuccessfulRequests uint16
	NumOffLedgerRequests  uint16
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

// RequestLookupKeyList a list of RequestLookupReference of requests with colliding coretypes.RequestLookupDigest
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
	if len(ll) > util.MaxUint16 {
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

// region RequestLogReqcord /////////////////////////////////////////////////////

// RequestLogRecord represents log record of processed request on the chain
type RequestLogRecord struct {
	RequestID coretypes.RequestID
	OffLedger bool
	LogData   []byte
	// not persistent
	BlockIndex   uint32
	RequestIndex uint16
}

func RequestLogRecordFromBytes(data []byte) (*RequestLogRecord, error) {
	return RequestLogRecordFromMarshalutil(marshalutil.New(data))
}

func RequestLogRecordFromMarshalutil(mu *marshalutil.MarshalUtil) (*RequestLogRecord, error) {
	ret := &RequestLogRecord{}
	var err error
	if ret.RequestID, err = coretypes.RequestIDFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.OffLedger, err = mu.ReadBool(); err != nil {
		return nil, err
	}
	var size uint16
	if size, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if ret.LogData, err = mu.ReadBytes(int(size)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *RequestLogRecord) Bytes() []byte {
	mu := marshalutil.New()
	mu.Write(r.RequestID).
		WriteBool(r.OffLedger).
		WriteUint16(uint16(len(r.LogData))).
		WriteBytes(r.LogData)
	return mu.Bytes()
}

func (r *RequestLogRecord) WithBlockData(blockIndex uint32, requestIndex uint16) *RequestLogRecord {
	r.BlockIndex = blockIndex
	r.RequestIndex = requestIndex
	return r
}

func (r *RequestLogRecord) strPrefix() string {
	prefix := "req"
	if !r.OffLedger {
		prefix += "/tx"
	}
	if r.BlockIndex != 0 {
		prefix += fmt.Sprintf("[%d/%d]", r.BlockIndex, r.RequestIndex)
	}
	return prefix
}

func (r *RequestLogRecord) String() string {
	ret := fmt.Sprintf("%s %s", r.strPrefix(), r.RequestID.String())
	if len(r.LogData) > 0 {
		ret += ": '" + string(r.LogData) + "'"
	}
	return ret
}

func (r *RequestLogRecord) Short() string {
	ret := fmt.Sprintf("%s %s", r.strPrefix(), r.RequestID.Short())
	if len(r.LogData) > 0 {
		ret += ": '" + string(r.LogData) + "'"
	}
	return ret
}

// endregion  /////////////////////////////////////////////////////////////
