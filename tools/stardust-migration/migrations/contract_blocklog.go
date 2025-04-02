package migrations

import (
	"fmt"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_coreutil "github.com/nnikolash/wasp-types-exported/packages/isc/coreutil"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"

	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

// Fakes the previous anchor id to be the state index (for debugging)
func Uint32ToAddress(value uint32) iotago.ObjectID {
	var addr iotago.Address

	addr[0] = uint8(value >> 24)
	addr[1] = uint8(value >> 16)
	addr[2] = uint8(value >> 8)
	addr[3] = uint8(value)

	return addr
}

func migrateBlockRegistry(blockIndex uint32, chainOwner *cryptolib.Address, stateMetadata *transaction.StateMetadata, oldState old_kv.KVStoreReader, newState kv.KVStore) blocklog.BlockInfo {
	oldBlocks := old_collections.NewArrayReadOnly(oldState, old_blocklog.PrefixBlockRegistry)
	newBlocks := collections.NewArray(newState, blocklog.PrefixBlockRegistry)
	newBlocks.SetSize(blockIndex + 1)

	oldBlock := oldBlocks.GetAt(blockIndex)

	// TODO: Just poll latest L1Params once so we can save some space here and have more accurate info.
	defaultL1Params := &parameters.L1Params{
		BaseToken: &parameters.IotaCoinInfo{
			CoinType:    coin.BaseTokenType,
			Name:        "Iota",
			Symbol:      "IOTA",
			Description: "IOTA",
			IconURL:     "http://iota.org",
			Decimals:    parameters.BaseTokenDecimals,
			TotalSupply: 9978371123948460000,
		},
		Protocol: &parameters.Protocol{
			Epoch:                 iotajsonrpc.NewBigInt(100),
			ProtocolVersion:       iotajsonrpc.NewBigInt(1),
			SystemStateVersion:    iotajsonrpc.NewBigInt(1),
			ReferenceGasPrice:     iotajsonrpc.NewBigInt(1000),
			EpochStartTimestampMs: iotajsonrpc.NewBigInt(1734538812318),
			EpochDurationMs:       iotajsonrpc.NewBigInt(86400000),
		},
	}

	oldBlockInfo, err := old_blocklog.BlockInfoFromBytes(oldBlock)
	if err != nil {
		panic(fmt.Errorf("blockRegistry migration error: %v", err))
	}

	// This is a reference to the PREVIOUS anchor.
	// Which means its state index - 1, and the past StateMetadata too!
	var migrationPreviousAnchor *isc.StateAnchor = nil

	iotaObjectID := iotago.ObjectID{1, 0, 7, 4}

	// Block 0 has no past anchor, so that's a nil.
	if blockIndex != 0 {
		previousAnchor := isc.NewStateAnchor(&iscmove.AnchorWithRef{
			Object: &iscmove.Anchor{
				Assets: iscmove.Referent[iscmove.AssetsBag]{ID: iotaObjectID, Value: &iscmove.AssetsBag{
					ID:   iotago.ObjectID{},
					Size: 0,
				}},
				ID:            iotaObjectID,
				StateIndex:    blockIndex - 1,
				StateMetadata: stateMetadata.Bytes(),
			},
			ObjectRef: iotago.ObjectRef{
				ObjectID: &iotaObjectID,
				Digest:   iotago.DigestFromBytes([]byte("MIGRATEDMIGRATEDMIGRATEDMIGRATED")), // TODO: fix this dummy ID
				Version:  iotago.SequenceNumber(blockIndex - 1),
			},
			Owner: chainOwner.AsIotaAddress(),
		}, iotaObjectID)
		migrationPreviousAnchor = &previousAnchor
	}

	newBlockInfo := blocklog.BlockInfo{
		GasFeeCharged:         ConvertOldCoinDecimalsToNew(oldBlockInfo.GasFeeCharged),
		GasBurned:             oldBlockInfo.GasBurned,
		BlockIndex:            oldBlockInfo.BlockIndex(),
		NumOffLedgerRequests:  oldBlockInfo.NumOffLedgerRequests,
		NumSuccessfulRequests: oldBlockInfo.NumSuccessfulRequests,
		SchemaVersion:         allmigrations.SchemaVersionMigratedRebased,
		Timestamp:             oldBlockInfo.Timestamp,
		TotalRequests:         oldBlockInfo.TotalRequests,
		PreviousAnchor:        migrationPreviousAnchor,
		L1Params:              defaultL1Params,
	}

	newBlocks.SetAt(blockIndex, newBlockInfo.Bytes())

	return newBlockInfo
}

func migrateRequestLookupIndex(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	oldLookup := old_collections.NewMapReadOnly(oldState, old_blocklog.PrefixRequestLookupIndex)
	newLookup := collections.NewMap(newState, blocklog.PrefixRequestLookupIndex)

	progress := NewProgressPrinter(500)
	oldLookup.Iterate(func(elemKey, oldIndex []byte) bool {
		// TODO: should the key be also migrated (re-encoded)?
		if oldIndex == nil {
			newLookup.DelAt(elemKey)
		} else {
			oldLookupKeys, err := old_blocklog.RequestLookupKeyListFromBytes(oldIndex)
			if err != nil {
				panic(fmt.Errorf("requestLookupIndex migration error: %v", err))
			}

			newLookupKeys := blocklog.RequestLookupKeyList{}
			for _, l := range oldLookupKeys {
				newLookupKeys = append(newLookupKeys, blocklog.NewRequestLookupKey(l.BlockIndex(), l.RequestIndex()))
			}

			// TODO: Check if we can take over the original key, I assume so. But double check it
			newLookup.SetAt(elemKey, newLookupKeys.Bytes())
		}

		progress.Print()
		return true
	})
}

func migrateSingleReceipt(receipt *old_blocklog.RequestReceipt, oldChainID old_isc.ChainID) blocklog.RequestReceipt {
	var burnLog *gas.BurnLog

	if receipt.GasBurnLog != nil {
		burnLog = &gas.BurnLog{}

		for _, b := range receipt.GasBurnLog.Records {
			burnLog.Records = append(burnLog.Records, gas.BurnRecord{
				Code:      gas.BurnCode(b.Code),
				GasBurned: b.GasBurned,
			})
		}
	}

	var receiptError *isc.UnresolvedVMError
	if receipt.Error != nil {
		errorParams := make([]isc.VMErrorParam, len(receipt.Error.Params))

		for i, e := range receipt.Error.Params {
			errorParams[i] = isc.VMErrorParam(e)
		}

		receiptError = &isc.UnresolvedVMError{
			ErrorCode: isc.VMErrorCode{
				ID:         receipt.Error.ErrorCode.ID,
				ContractID: OldHnameToNewHname(receipt.Error.ErrorCode.ContractID),
			},
			Params: errorParams,
		}
	}

	return blocklog.RequestReceipt{
		Request:       MigrateSingleRequest(receipt.Request, oldChainID),
		Error:         receiptError,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: ConvertOldCoinDecimalsToNew(receipt.GasFeeCharged),
		GasBurnLog:    burnLog,
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
	}
}

func migrateRequestReceipts(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID) {
	oldRequests := old_collections.NewMapReadOnly(oldState, old_blocklog.PrefixRequestReceipts)
	newRequests := collections.NewMap(newState, blocklog.PrefixRequestReceipts)

	cli.DebugLogf("Migrating request receipts (%d)", oldRequests.Len())

	progress := NewProgressPrinter(500)
	oldRequests.Iterate(func(elemKey []byte, value []byte) bool {
		// TODO: should the key be also migrated (re-encoded)?
		if value == nil {
			newRequests.DelAt(elemKey)
		} else {
			// TODO: Validate if this is fine. BlockIndex and ReqIndex is 0 here, as we don't persist these values in the db
			// So in my understanding, using 0 here is fine. If not, we need to iterate the whole request lut again and combine the tables.
			// I added a solution in commit: 96504e6165ed4056a3e8a50281215f3d7eb7c015, for now I go without.
			oldReceipt, err := old_blocklog.RequestReceiptFromBytes(value, 0, 0)
			if err != nil {
				panic(fmt.Errorf("requestReceipt migration error: %v", err))
			}

			newReceipt := migrateSingleReceipt(oldReceipt, oldChainID)
			newRequests.SetAt(elemKey, newReceipt.Bytes())
		}

		progress.Print()

		return true
	})
}

func printWarningsForUnprocessableRequests(oldState old_kv.KVStoreReader) {
	// No need to migrate. Just print a warning if there are any

	cli.DebugLogf("Listing Unprocessable Requests...")

	count := IterateByPrefix(oldState, old_blocklog.PrefixUnprocessableRequests, func(k isc.RequestID, v []byte) {
		if v != nil {
			cli.DebugLogf("Warning: unprocessable request found %v", k.String())
		}
	})

	cli.DebugLogf("Listing Unprocessable Requests completed (found %v records)", count)
}

func migrateGlobalVars(oldState old_kv.KVStoreReader, newState kv.KVStore) uint32 {
	oldTimeBytes := oldState.Get(old_kv.Key(old_coreutil.StatePrefixTimestamp))
	oldTime := lo.Must(old_codec.DecodeTime(oldTimeBytes))
	newState.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Encode(oldTime))

	oldBlockIndex := old_codec.MustDecodeUint32(oldState.Get(old_kv.Key(old_coreutil.StatePrefixBlockIndex)))
	newState.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(oldBlockIndex))

	return oldBlockIndex
}

func MigrateBlocklogContract(oldChainState old_kv.KVStoreReader, newChainState kv.KVStore, oldChainID old_isc.ChainID, stateMetadata *transaction.StateMetadata, chainOwner *cryptolib.Address) blocklog.BlockInfo {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, blocklog.Contract.Hname())

	//printWarningsForUnprocessableRequests(oldContractState)
	blockIndex := migrateGlobalVars(oldChainState, newChainState)
	blockInfo := migrateBlockRegistry(blockIndex, chainOwner, stateMetadata, oldContractState, newContractState)
	migrateRequestLookupIndex(oldContractState, newContractState)
	migrateRequestReceipts(oldContractState, newContractState, oldChainID)

	return blockInfo
}
