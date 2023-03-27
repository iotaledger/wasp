package blocklog

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
)

const (
	BlockInfoLatestSchemaVersion = 0
)

type BlockInfo struct {
	SchemaVersion         uint8
	Timestamp             time.Time
	TotalRequests         uint16
	NumSuccessfulRequests uint16 // which didn't panic
	NumOffLedgerRequests  uint16
	PreviousAliasOutput   *isc.AliasOutputWithID // if new schema => always not nil
	GasBurned             uint64
	GasFeeCharged         uint64
}

// RequestTimestamp returns timestamp which corresponds to the request with the given index
// Timestamps of requests are incremented by 1 nanosecond in the block. The timestamp of the last one
// is equal to the timestamp pof the block
func (bi *BlockInfo) RequestTimestamp(requestIndex uint16) time.Time {
	return bi.Timestamp.Add(time.Duration(-(bi.TotalRequests - requestIndex - 1)) * time.Nanosecond)
}

func (bi *BlockInfo) PreviousL1Commitment() *state.L1Commitment {
	if bi.PreviousAliasOutput == nil {
		return nil
	}
	l1c, err := transaction.L1CommitmentFromAliasOutput(bi.PreviousAliasOutput.GetAliasOutput())
	if err != nil {
		panic(err)
	}
	return l1c
}

func (bi *BlockInfo) String() string {
	ret := fmt.Sprintf("Block index: %d\n", bi.BlockIndex())
	ret += fmt.Sprintf("SchemaVersion: %d\n", bi.SchemaVersion)
	ret += fmt.Sprintf("Timestamp: %d\n", bi.Timestamp.Unix())
	ret += fmt.Sprintf("Total requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("off-ledger requests: %d\n", bi.NumOffLedgerRequests)
	ret += fmt.Sprintf("Succesfull requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("Prev AliasOutput: %s\n", bi.PreviousAliasOutput.String())
	ret += fmt.Sprintf("Gas burned: %d\n", bi.GasBurned)
	ret += fmt.Sprintf("Gas fee charged: %d\n", bi.GasFeeCharged)
	return ret
}

func (bi *BlockInfo) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint8(bi.SchemaVersion)
	mu.WriteTime(bi.Timestamp)
	mu.WriteUint16(bi.TotalRequests)
	mu.WriteUint16(bi.NumSuccessfulRequests)
	mu.WriteUint16(bi.NumOffLedgerRequests)
	mu.WriteBool(bi.PreviousAliasOutput != nil)
	if bi.PreviousAliasOutput != nil {
		mu.WriteBytes(bi.PreviousAliasOutput.Bytes())
	}
	mu.WriteUint64(bi.GasBurned)
	mu.WriteUint64(bi.GasFeeCharged)
	return mu.Bytes()
}

func BlockInfoFromBytes(data []byte) (*BlockInfo, error) {
	mu := marshalutil.New(data)
	var err error
	bi := &BlockInfo{}
	if bi.SchemaVersion, err = mu.ReadUint8(); err != nil {
		return nil, err
	}
	if bi.Timestamp, err = mu.ReadTime(); err != nil {
		return nil, err
	}
	if bi.TotalRequests, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if bi.NumSuccessfulRequests, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if bi.NumOffLedgerRequests, err = mu.ReadUint16(); err != nil {
		return nil, err
	}
	if hasPreviousAliasOutput, err2 := mu.ReadBool(); err2 != nil {
		return nil, err2
	} else if hasPreviousAliasOutput {
		if bi.PreviousAliasOutput, err2 = isc.NewAliasOutputWithIDFromMarshalUtil(mu); err2 != nil {
			return nil, err2
		}
	}
	if bi.GasBurned, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if bi.GasFeeCharged, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if done, err := mu.DoneReading(); err != nil {
		return nil, err
	} else if !done {
		return nil, errors.New("BlockInfoFromBytes: remaining bytes")
	}
	return bi, nil
}

// BlockInfoKey a key to access block info record inside SC state
func BlockInfoKey(index uint32) []byte {
	return []byte(collections.Array32ElemKey(PrefixBlockRegistry, index))
}

func (bi *BlockInfo) BlockIndex() uint32 {
	if bi.PreviousAliasOutput == nil {
		return 0
	}
	return bi.PreviousAliasOutput.GetStateIndex() + 1
}
