package blocklog

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
)

const (
	BlockInfoAliasOutputSchemaVersion = 1
	BlockInfoLatestSchemaVersion      = BlockInfoAliasOutputSchemaVersion

	blockInfoPreamble = math.MaxUint64
)

type BlockInfo struct {
	SchemaVersion         uint8
	BlockIndex            uint32 // not persistent. Set from key
	Timestamp             time.Time
	TotalRequests         uint16
	NumSuccessfulRequests uint16 // which didn't panic
	NumOffLedgerRequests  uint16

	// schema < BlockInfoAliasOutputSchemaVersion
	// TODO remove at some point?
	previousL1CommitmentOld *state.L1Commitment  // if not nil => old schema version
	l1CommitmentOld         *state.L1Commitment  // if old schema => nil when not known yet for the current state
	anchorTransactionIDOld  iotago.TransactionID // of the input state

	// schema >= BlockInfoAliasOutputSchemaVersion
	PreviousAliasOutput *isc.AliasOutputWithID // if new schema => always not nil
	AliasOutput         *isc.AliasOutputWithID // if new schema => nil when not known yet for the current state

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

// RequestTimestamp returns timestamp which corresponds to the request with the given index
// Timestamps of requests are incremented by 1 nanosecond in the block. The timestamp of the last one
// is equal to the timestamp pof the block
func (bi *BlockInfo) RequestTimestamp(requestIndex uint16) time.Time {
	return bi.Timestamp.Add(time.Duration(-(bi.TotalRequests - requestIndex - 1)) * time.Nanosecond)
}

func (bi *BlockInfo) AnchorTransactionID() iotago.TransactionID {
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		return bi.anchorTransactionIDOld
	}
	if bi.AliasOutput == nil {
		return iotago.TransactionID{}
	}
	return bi.AliasOutput.TransactionID()
}

func (bi *BlockInfo) L1Commitment() *state.L1Commitment {
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		return bi.l1CommitmentOld
	}
	if bi.AliasOutput == nil {
		return nil
	}
	l1c, err := transaction.L1CommitmentFromAliasOutput(bi.AliasOutput.GetAliasOutput())
	if err != nil {
		panic(err)
	}
	return l1c
}

func (bi *BlockInfo) PreviousL1Commitment() *state.L1Commitment {
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		return bi.previousL1CommitmentOld
	}
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
	ret := fmt.Sprintf("Block index: %d\n", bi.BlockIndex)
	ret += fmt.Sprintf("Timestamp: %d\n", bi.Timestamp.Unix())
	ret += fmt.Sprintf("Total requests: %d\n", bi.TotalRequests)
	ret += fmt.Sprintf("off-ledger requests: %d\n", bi.NumOffLedgerRequests)
	ret += fmt.Sprintf("Succesfull requests: %d\n", bi.NumSuccessfulRequests)
	ret += fmt.Sprintf("Prev AliasOutput: %s\n", bi.PreviousAliasOutput.String())
	ret += fmt.Sprintf("AliasOutput: %s\n", bi.AliasOutput.String())
	ret += fmt.Sprintf("Total base tokens in contracts: %d\n", bi.TotalBaseTokensInL2Accounts)
	ret += fmt.Sprintf("Total base tokens locked in storage deposit: %d\n", bi.TotalStorageDeposit)
	ret += fmt.Sprintf("Gas burned: %d\n", bi.GasBurned)
	ret += fmt.Sprintf("Gas fee charged: %d\n", bi.GasFeeCharged)
	return ret
}

func (bi *BlockInfo) Bytes() []byte {
	mu := marshalutil.New()
	if bi.SchemaVersion >= BlockInfoAliasOutputSchemaVersion {
		// old version starts with the timestamp. Assuming here that it can never take this value
		mu.WriteUint64(blockInfoPreamble)
		mu.WriteUint8(bi.SchemaVersion)
	}
	mu.WriteTime(bi.Timestamp)
	mu.WriteUint16(bi.TotalRequests)
	mu.WriteUint16(bi.NumSuccessfulRequests)
	mu.WriteUint16(bi.NumOffLedgerRequests)
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		mu.WriteBytes(bi.anchorTransactionIDOld[:])
	}
	mu.WriteBytes(bi.TransactionSubEssenceHash[:])
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		mu.WriteBytes(bi.previousL1CommitmentOld.Bytes())
		mu.WriteBool(bi.l1CommitmentOld != nil)
		if bi.l1CommitmentOld != nil {
			mu.WriteBytes(bi.l1CommitmentOld.Bytes())
		}
	} else {
		mu.WriteBool(bi.PreviousAliasOutput != nil)
		if bi.PreviousAliasOutput != nil {
			mu.WriteBytes(bi.PreviousAliasOutput.Bytes())
		}
		mu.WriteBool(bi.AliasOutput != nil)
		if bi.AliasOutput != nil {
			mu.WriteBytes(bi.AliasOutput.Bytes())
		}
	}
	mu.WriteUint64(bi.TotalBaseTokensInL2Accounts)
	mu.WriteUint64(bi.TotalStorageDeposit)
	mu.WriteUint64(bi.GasBurned)
	mu.WriteUint64(bi.GasFeeCharged)
	return mu.Bytes()
}

//nolint:gocyclo,revive,govet
func BlockInfoFromBytes(blockIndex uint32, data []byte) (*BlockInfo, error) {
	mu := marshalutil.New(data)
	var err error
	bi := &BlockInfo{BlockIndex: blockIndex}
	if p, err := mu.ReadUint64(); err != nil {
		return nil, err
	} else {
		if p != blockInfoPreamble {
			// old version
			bi.SchemaVersion = 0
			mu.ReadSeek(-marshalutil.Uint64Size)
		} else {
			if bi.SchemaVersion, err = mu.ReadUint8(); err != nil {
				return nil, err
			}
			if bi.SchemaVersion == 0 || bi.SchemaVersion > BlockInfoLatestSchemaVersion {
				return nil, fmt.Errorf("BlockInfoFromBytes: unexpected schema version: %d", bi.SchemaVersion)
			}
		}
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
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		if buf, err := mu.ReadBytes(iotago.TransactionIDLength); err != nil {
			return nil, err
		} else {
			copy(bi.anchorTransactionIDOld[:], buf)
		}
	}
	if buf, err := mu.ReadBytes(TransactionEssenceHashLength); err != nil {
		return nil, err
	} else {
		copy(bi.TransactionSubEssenceHash[:], buf)
	}
	if bi.SchemaVersion < BlockInfoAliasOutputSchemaVersion {
		if bi.previousL1CommitmentOld, err = state.L1CommitmentFromMarshalUtil(mu); err != nil {
			return nil, err
		}
		if hasL1Commitment, err := mu.ReadBool(); err != nil {
			return nil, err
		} else if hasL1Commitment {
			if bi.l1CommitmentOld, err = state.L1CommitmentFromMarshalUtil(mu); err != nil {
				return nil, err
			}
		}
	} else {
		if hasPreviousAliasOutput, err := mu.ReadBool(); err != nil {
			return nil, err
		} else if hasPreviousAliasOutput {
			if bi.PreviousAliasOutput, err = isc.NewAliasOutputWithIDFromMarshalUtil(mu); err != nil {
				return nil, err
			}
		}
		if hasAliasOutput, err := mu.ReadBool(); err != nil {
			return nil, err
		} else if hasAliasOutput {
			if bi.AliasOutput, err = isc.NewAliasOutputWithIDFromMarshalUtil(mu); err != nil {
				return nil, err
			}
		}
	}
	if bi.TotalBaseTokensInL2Accounts, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if bi.TotalStorageDeposit, err = mu.ReadUint64(); err != nil {
		return nil, err
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
	return []byte(collections.Array32ElemKey(prefixBlockRegistry, index))
}
