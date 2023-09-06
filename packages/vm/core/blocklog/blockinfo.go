package blocklog

import (
	"fmt"
	"io"
	"time"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util/rwutil"
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
	PreviousAliasOutput   *isc.AliasOutputWithID // nil for block #0
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
	ret := "{\n"
	ret += fmt.Sprintf("\tBlock index: %d\n", bi.BlockIndex())
	ret += fmt.Sprintf("\tSchemaVersion: %d\n", bi.SchemaVersion)
	ret += fmt.Sprintf("\tTimestamp: %d\n", bi.Timestamp.Unix())
	ret += fmt.Sprintf("\tTotal requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("\toff-ledger requests: %d\n", bi.NumOffLedgerRequests)
	ret += fmt.Sprintf("\tSuccessful requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("\tPrev AliasOutput: %s\n", bi.PreviousAliasOutput.String())
	ret += fmt.Sprintf("\tGas burned: %d\n", bi.GasBurned)
	ret += fmt.Sprintf("\tGas fee charged: %d\n", bi.GasFeeCharged)
	ret += "}\n"
	return ret
}

func (bi *BlockInfo) Bytes() []byte {
	return rwutil.WriteToBytes(bi)
}

func BlockInfoFromBytes(data []byte) (*BlockInfo, error) {
	return rwutil.ReadFromBytes(data, new(BlockInfo))
}

// BlockInfoKey a key to access block info record inside SC state
func BlockInfoKey(index uint32) []byte {
	return []byte(collections.ArrayElemKey(PrefixBlockRegistry, index))
}

func (bi *BlockInfo) BlockIndex() uint32 {
	if bi.PreviousAliasOutput == nil {
		return 0
	}
	return bi.PreviousAliasOutput.GetStateIndex() + 1
}

func (bi *BlockInfo) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	bi.SchemaVersion = rr.ReadUint8()
	bi.Timestamp = time.Unix(0, rr.ReadInt64())
	bi.TotalRequests = rr.ReadUint16()
	bi.NumSuccessfulRequests = rr.ReadUint16()
	bi.NumOffLedgerRequests = rr.ReadUint16()
	hasPreviousAliasOutput := rr.ReadBool()
	if hasPreviousAliasOutput {
		bi.PreviousAliasOutput = &isc.AliasOutputWithID{}
		rr.Read(bi.PreviousAliasOutput)
	}
	bi.GasBurned = rr.ReadGas64()
	bi.GasFeeCharged = rr.ReadGas64()
	return rr.Err
}

func (bi *BlockInfo) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint8(bi.SchemaVersion)
	ww.WriteInt64(bi.Timestamp.UnixNano())
	ww.WriteUint16(bi.TotalRequests)
	ww.WriteUint16(bi.NumSuccessfulRequests)
	ww.WriteUint16(bi.NumOffLedgerRequests)
	ww.WriteBool(bi.PreviousAliasOutput != nil)
	if bi.PreviousAliasOutput != nil {
		ww.Write(bi.PreviousAliasOutput)
	}
	ww.WriteGas64(bi.GasBurned)
	ww.WriteGas64(bi.GasFeeCharged)
	return ww.Err
}
