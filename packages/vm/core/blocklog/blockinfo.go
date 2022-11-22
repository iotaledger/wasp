package blocklog

import (
	"bytes"
	"fmt"
	"io"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type BlockInfo struct {
	BlockIndex                  uint32 // not persistent. Set from key
	Timestamp                   time.Time
	TotalRequests               uint16
	NumSuccessfulRequests       uint16 // which didn't panic
	NumOffLedgerRequests        uint16
	PreviousL1Commitment        state.L1Commitment     // always known
	L1Commitment                *state.L1Commitment    // nil when not known yet for the current state
	AnchorTransactionID         iotago.TransactionID   // of the input state
	TransactionSubEssenceHash   TransactionEssenceHash // always known even without state commitment. Needed for fraud proofs
	TotalBaseTokensInL2Accounts uint64
	TotalStorageDeposit         uint64
	GasBurned                   uint64
	GasFeeCharged               uint64
}

// TransactionEssenceHash is a blake2b 256 bit hash of the essence of the transaction
// Used to calculate sub-essence hash
type TransactionEssenceHash [TransactionEssenceHashLength]byte

const TransactionEssenceHashLength = 32

func CalcTransactionEssenceHash(essence *iotago.TransactionEssence) (ret TransactionEssenceHash) {
	h, err := essence.SigningMessage()
	if err != nil {
		panic(err)
	}
	copy(ret[:], h)
	return
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
	ret += fmt.Sprintf("Anchor tx ID: %s\n", iotago.EncodeHex(bi.AnchorTransactionID[:]))
	ret += fmt.Sprintf("Total base tokens in contracts: %d\n", bi.TotalBaseTokensInL2Accounts)
	ret += fmt.Sprintf("Total base tokens locked in storage deposit: %d\n", bi.TotalStorageDeposit)
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
	if _, err := w.Write(bi.TransactionSubEssenceHash[:]); err != nil {
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
	if err := util.WriteUint64(w, bi.TotalBaseTokensInL2Accounts); err != nil {
		return err
	}
	if err := util.WriteUint64(w, bi.TotalStorageDeposit); err != nil {
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
	if err := ReadTransactionSubEssenceHash(r, &bi.TransactionSubEssenceHash); err != nil {
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
	if err := util.ReadUint64(r, &bi.TotalBaseTokensInL2Accounts); err != nil {
		return err
	}
	if err := util.ReadUint64(r, &bi.TotalStorageDeposit); err != nil {
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

func ReadTransactionSubEssenceHash(r io.Reader, h *TransactionEssenceHash) error {
	n, err := r.Read(h[:])
	if err != nil {
		return err
	}

	if n != TransactionEssenceHashLength {
		return fmt.Errorf("error while reading transaction subessence hash: read %d bytes, expected %d bytes",
			n, TransactionEssenceHashLength)
	}
	return nil
}

// BlockInfoKey a key to access block info record inside SC state
func BlockInfoKey(index uint32) []byte {
	return []byte(collections.Array32ElemKey(prefixBlockRegistry, index))
}
