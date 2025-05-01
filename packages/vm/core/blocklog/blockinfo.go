package blocklog

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
)

const (
	blockInfoSchemaVersion0 = iota
	blockInfoSchemaVersionAddedEntropy

	BlockInfoLatestSchemaVersion = blockInfoSchemaVersionAddedEntropy
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
	Entropy               hashing.HashValue
}

func init() {
	bcs.AddCustomEncoder(func(e *bcs.Encoder, bi *BlockInfo) error {
		e.Encode(bi.SchemaVersion)
		e.Encode(bi.BlockIndex)
		e.Encode(bi.Timestamp)
		e.EncodeOptional(bi.PreviousAnchor)
		e.Encode(bi.L1Params)
		e.Encode(bi.TotalRequests)
		e.Encode(bi.NumSuccessfulRequests)
		e.Encode(bi.NumOffLedgerRequests)
		e.WriteCompactUint64(bi.GasBurned)
		e.WriteCompactUint64(bi.GasFeeCharged.Uint64())
		if bi.SchemaVersion >= blockInfoSchemaVersionAddedEntropy {
			e.Encode(bi.Entropy)
		}
		return e.Err()
	})

	bcs.AddCustomDecoder(func(d *bcs.Decoder, bi *BlockInfo) error {
		d.Decode(&bi.SchemaVersion)
		d.Decode(&bi.BlockIndex)
		d.Decode(&bi.Timestamp)
		_ = d.DecodeOptional(&bi.PreviousAnchor)
		d.Decode(&bi.L1Params)
		d.Decode(&bi.TotalRequests)
		d.Decode(&bi.NumSuccessfulRequests)
		d.Decode(&bi.NumOffLedgerRequests)
		bi.GasBurned = d.ReadCompactUint64()
		bi.GasFeeCharged = coin.Value(d.ReadCompactUint64())
		if bi.SchemaVersion >= blockInfoSchemaVersionAddedEntropy {
			d.Decode(&bi.Entropy)
		} else {
			// we are missing entropy information; assign some unique hash for the given block index
			bi.Entropy = hashing.HashData(bcs.MustMarshal(&bi.BlockIndex))
		}
		return d.Err()
	})
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
