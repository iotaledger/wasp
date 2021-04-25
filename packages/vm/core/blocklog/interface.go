// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package blocklog

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
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
	BlockRegistry = "b"
	// functions
	FuncGetBlockInfo       = "getBlockInfo"
	FuncGetLatestBlockInfo = "getLatestBlockInfo"
	// parameters
	ParamBlockIndex = "n"
	ParamBlockInfo  = "i"
)

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
