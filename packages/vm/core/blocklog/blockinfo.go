package blocklog

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
)

const (
	BlockInfoLatestSchemaVersion = 0
)

type BlockInfo struct {
	SchemaVersion         uint8
	BlockIndex            uint32
	Timestamp             time.Time
	PreviousAnchor        *isc.StateAnchor `bcs:"optional"`
	L1Params              *parameters.L1Params
	TotalRequests         uint16
	NumSuccessfulRequests uint16 // which didn't panic
	NumOffLedgerRequests  uint16
	GasBurned             uint64     `bcs:"compact"`
	GasFeeCharged         coin.Value `bcs:"compact"`
}

// RequestTimestamp returns timestamp which corresponds to the request with the given index
// Timestamps of requests are incremented by 1 nanosecond in the block. The timestamp of the last one
// is equal to the timestamp pof the block
func (bi *BlockInfo) RequestTimestamp(requestIndex uint16) time.Time {
	return bi.Timestamp.Add(time.Duration(-(bi.TotalRequests - requestIndex - 1)) * time.Nanosecond)
}

func (bi *BlockInfo) String() string {
	ret := "{\n"
	ret += fmt.Sprintf("\tBlock index: %d\n", bi.BlockIndex)
	ret += fmt.Sprintf("\tSchemaVersion: %d\n", bi.SchemaVersion)
	ret += fmt.Sprintf("\tTimestamp: %d\n", bi.Timestamp.Unix())
	if bi.PreviousAnchor != nil {
		ret += fmt.Sprintf("\tPackageID: %v\n", bi.PreviousAnchor.ISCPackage())
		ret += fmt.Sprintf("\tAnchor: %v\n", bi.PreviousAnchor.Anchor().String())
	}
	ret += fmt.Sprintf("\tL1Params: %v\n", bi.L1Params.String())
	ret += fmt.Sprintf("\tTotal requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("\toff-ledger requests: %d\n", bi.NumOffLedgerRequests)
	ret += fmt.Sprintf("\tSuccessful requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("\tPrev L1Commitment: %v\n", bi.PreviousL1Commitment())
	ret += fmt.Sprintf("\tGas burned: %d\n", bi.GasBurned)
	ret += fmt.Sprintf("\tGas fee charged: %d\n", bi.GasFeeCharged)
	ret += "}\n"
	return ret
}

func (bi *BlockInfo) PreviousL1Commitment() *state.L1Commitment {
	if bi.PreviousAnchor == nil {
		return nil
	}

	return lo.Must(transaction.L1CommitmentFromAnchor(bi.PreviousAnchor))
}

func (bi *BlockInfo) Bytes() []byte {
	return bcs.MustMarshal(bi)
}

func BlockInfoFromBytes(data []byte) (*BlockInfo, error) {
	return bcs.Unmarshal[*BlockInfo](data)
}

// BlockInfoKey a key to access block info record inside SC state
func BlockInfoKey(index uint32) []byte {
	return []byte(collections.ArrayElemKey(prefixBlockRegistry, index))
}
