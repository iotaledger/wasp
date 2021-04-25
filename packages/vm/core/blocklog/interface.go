// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
	"io"
	"time"
)

const (
	Name        = "blocklog"
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
		coreutil.ViewFunc(FuncGetBlockInfo, getBlockInfo),
		coreutil.ViewFunc(FuncGetLatestBlockInfo, getLatestBlockInfo),
	})
}

const (
	// state variables
	BlockRegistry    = "b"
	RequestLookupMap = "l"
	// functions
	FuncGetBlockInfo       = "getBlockInfo"
	FuncGetLatestBlockInfo = "getLatestBlockInfo"
	// parameters
	ParamBlockIndex = "n"
	ParamBlockInfo  = "i"
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
	ret += fmt.Sprintf("Timestamp: %v\n", bi.Timestamp)
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

// region RequestBlockReference //////////////////////////////////////
type RequestBlockReference struct {
	BlockIndex   uint32
	RequestIndex uint16
}

func (ref *RequestBlockReference) Write(w io.Writer) error {
	if err := util.WriteUint32(w, ref.BlockIndex); err != nil {
		return err
	}
	if err := util.WriteUint16(w, ref.RequestIndex); err != nil {
		return err
	}
	return nil
}

func (ref *RequestBlockReference) Read(r io.Reader) error {
	if err := util.ReadUint32(r, &ref.BlockIndex); err != nil {
		return err
	}
	if err := util.ReadUint16(r, &ref.RequestIndex); err != nil {
		return err
	}
	return nil
}

// endregion

// region RequestLookupList //////////////////////////////////////////////

type RequestLookupList struct {
	lst []RequestBlockReference
}

func NewRequestLookupList() *RequestLookupList {
	return &RequestLookupList{
		lst: make([]RequestBlockReference, 0),
	}
}

func RequestLookupListFromBytes(data []byte) (*RequestLookupList, error) {
	ret := NewRequestLookupList()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (ll *RequestLookupList) Bytes() []byte {
	var buf bytes.Buffer
	_ = ll.Write(&buf)
	return buf.Bytes()
}

func (ll *RequestLookupList) List() []RequestBlockReference {
	return ll.lst
}

func (ll *RequestLookupList) Append(ref ...RequestBlockReference) {
	ll.lst = append(ll.lst, ref...)
}

func (ll *RequestLookupList) Write(w io.Writer) error {
	if len(ll.lst) > util.MaxUint16 {
		return xerrors.New("RequestLookupList::Write: too long")
	}
	if err := util.WriteUint16(w, uint16(len(ll.lst))); err != nil {
		return err
	}
	for i := range ll.lst {
		if err := ll.lst[i].Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (ll *RequestLookupList) Read(r io.Reader) error {
	var size uint16
	if err := util.ReadUint16(r, &size); err != nil {
		return err
	}
	ll.lst = make([]RequestBlockReference, size)
	for i := range ll.lst {
		if err := ll.lst[i].Read(r); err != nil {
			return err
		}
	}
	return nil
}

// endregion /////////////////////////////////////////////////////////////
