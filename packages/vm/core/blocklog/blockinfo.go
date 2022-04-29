package blocklog

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type BlockInfo struct {
	BlockIndex             uint32 // not persistent. Set from key
	Timestamp              time.Time
	TotalRequests          uint16
	NumSuccessfulRequests  uint16
	NumOffLedgerRequests   uint16
	PreviousL1Commitment   state.L1Commitment  // always known
	L1Commitment           *state.L1Commitment // nil if not known
	AnchorTransactionID    iotago.TransactionID
	TotalIotasInL2Accounts uint64
	TotalDustDeposit       uint64
	GasBurned              uint64
	GasFeeCharged          uint64
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
	ret += fmt.Sprintf("Timestamp: %d\n", bi.Timestamp.Unix())
	ret += fmt.Sprintf("Total requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("off-ledger requests: %d\n", bi.NumOffLedgerRequests)
	ret += fmt.Sprintf("Succesfull requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("Prev L1 commitment: %s\n", bi.PreviousL1Commitment.String())
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
	if err := bi.PreviousL1Commitment.Write(w); err != nil {
		return err
	}
	if err := util.WriteBoolByte(w, bi.L1Commitment != nil); err != nil {
		return err
	}
	if bi.L1Commitment != nil {
		if err := bi.L1Commitment.Write(w); err != nil {
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
	if err := bi.PreviousL1Commitment.Read(r); err != nil {
		return err
	}
	var knownStateCommitments bool
	if err := util.ReadBoolByte(r, &knownStateCommitments); err != nil {
		return err
	}
	bi.L1Commitment = nil
	if knownStateCommitments {
		bi.L1Commitment = &state.L1Commitment{}
		if err := bi.L1Commitment.Read(r); err != nil {
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

// BlockInfoKey a key to access block info record inside SC state
func BlockInfoKey(index uint32) []byte {
	return []byte(collections.Array32ElemKey(prefixBlockRegistry, index))
}
