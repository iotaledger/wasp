package migrations

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

var m003CleanupBlockInfo = Migration{
	Contract: blocklog.Contract,
	Apply: func(state kv.KVStore, log *logger.Logger) error {
		registry := collections.NewArray32(state, blocklog.PrefixBlockRegistry)

		for i := uint32(0); i < registry.MustLen(); i++ {
			biBinOld := registry.MustGetAt(i)
			biNew, err := m003ConvertBlockInfo(biBinOld)
			if err != nil {
				return err
			}
			registry.MustSetAt(i, biNew.Bytes())
		}
		return nil
	},
}

func m003ConvertBlockInfo(oldBin []byte) (*blocklog.BlockInfo, error) {
	biOld, err := blockInfoPreCleanupFromBytes(oldBin)
	if err != nil {
		return nil, err
	}
	return &blocklog.BlockInfo{
		SchemaVersion:         blocklog.BlockInfoDeprecateSDAssumptionsVersion,
		Timestamp:             biOld.Timestamp,
		TotalRequests:         biOld.TotalRequests,
		NumSuccessfulRequests: biOld.NumSuccessfulRequests,
		NumOffLedgerRequests:  biOld.NumOffLedgerRequests,
		PreviousAliasOutput:   biOld.PreviousAliasOutput,
		AliasOutput:           biOld.AliasOutput,
		GasBurned:             biOld.GasBurned,
		GasFeeCharged:         biOld.GasFeeCharged,
	}, nil
}

type transactionEssenceHash [transactionEssenceHashLength]byte

const transactionEssenceHashLength = 32

type blockInfoPreCleanup struct {
	SchemaVersion         uint8
	BlockIndex            uint32 // not persistent. Set from key
	Timestamp             time.Time
	TotalRequests         uint16
	NumSuccessfulRequests uint16 // which didn't panic
	NumOffLedgerRequests  uint16

	// schema < BlockInfoAliasOutputSchemaVersion
	previousL1CommitmentOld *state.L1Commitment  // if not nil => old schema version
	l1CommitmentOld         *state.L1Commitment  // if old schema => nil when not known yet for the current state
	anchorTransactionIDOld  iotago.TransactionID // of the input state

	// schema >= BlockInfoAliasOutputSchemaVersion
	PreviousAliasOutput *isc.AliasOutputWithID // if new schema => always not nil
	AliasOutput         *isc.AliasOutputWithID // if new schema => nil when not known yet for the current state

	TransactionSubEssenceHash   transactionEssenceHash // always known even without state commitment. Needed for fraud proofs
	TotalBaseTokensInL2Accounts uint64
	TotalStorageDeposit         uint64
	GasBurned                   uint64
	GasFeeCharged               uint64
}

const (
	blockInfoPreamble                 = math.MaxUint64
	blockInfoAliasOutputSchemaVersion = 1
)

//nolint:gocyclo
func blockInfoPreCleanupFromBytes(data []byte) (*blockInfoPreCleanup, error) {
	mu := marshalutil.New(data)
	var err error
	bi := &blockInfoPreCleanup{}
	p, err := mu.ReadUint64()
	if err != nil {
		return nil, err
	}
	if p != blockInfoPreamble {
		// old version
		bi.SchemaVersion = 0
		mu.ReadSeek(-marshalutil.Uint64Size)
	} else {
		if bi.SchemaVersion, err = mu.ReadUint8(); err != nil {
			return nil, err
		}
		if bi.SchemaVersion == 0 {
			return nil, fmt.Errorf("BlockInfoFromBytes: unexpected schema version: %d", bi.SchemaVersion)
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
	if bi.SchemaVersion < blockInfoAliasOutputSchemaVersion {
		buf, err2 := mu.ReadBytes(iotago.TransactionIDLength)
		if err2 != nil {
			return nil, err2
		}
		copy(bi.anchorTransactionIDOld[:], buf)
	}
	buf, err2 := mu.ReadBytes(transactionEssenceHashLength)
	if err2 != nil {
		return nil, err2
	}
	copy(bi.TransactionSubEssenceHash[:], buf)
	if bi.SchemaVersion < blockInfoAliasOutputSchemaVersion {
		if bi.previousL1CommitmentOld, err = state.L1CommitmentFromMarshalUtil(mu); err != nil {
			return nil, err
		}
		if hasL1Commitment, err2 := mu.ReadBool(); err2 != nil {
			return nil, err2
		} else if hasL1Commitment {
			if bi.l1CommitmentOld, err2 = state.L1CommitmentFromMarshalUtil(mu); err2 != nil {
				return nil, err2
			}
		}
	} else {
		if hasPreviousAliasOutput, err2 := mu.ReadBool(); err2 != nil {
			return nil, err2
		} else if hasPreviousAliasOutput {
			if bi.PreviousAliasOutput, err2 = isc.NewAliasOutputWithIDFromMarshalUtil(mu); err2 != nil {
				return nil, err2
			}
		}
		if hasAliasOutput, err2 := mu.ReadBool(); err2 != nil {
			return nil, err2
		} else if hasAliasOutput {
			if bi.AliasOutput, err2 = isc.NewAliasOutputWithIDFromMarshalUtil(mu); err2 != nil {
				return nil, err2
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
