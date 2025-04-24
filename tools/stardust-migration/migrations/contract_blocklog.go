package migrations

import (
	"fmt"
	"math"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_coreutil "github.com/nnikolash/wasp-types-exported/packages/isc/coreutil"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_rwutil "github.com/nnikolash/wasp-types-exported/packages/util/rwutil"
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

func migrateBlockRegistry(blockIndex uint32, blockKeepAmount int32, chainOwner *cryptolib.Address, stateMetadata *transaction.StateMetadata, oldState old_kv.KVStoreReader, newState kv.KVStore) blocklog.BlockInfo {
	oldBlocks := old_collections.NewArrayReadOnly(oldState, old_blocklog.PrefixBlockRegistry)
	newBlocks := collections.NewArray(newState, blocklog.PrefixBlockRegistry)
	newBlocks.SetSize(blockIndex + 1)

	if blockKeepAmount > 0 && blockIndex >= uint32(blockKeepAmount) {
		// Anyway this migration already won't work with "-i" option, so we can just
		// skip iteration over the store and just straight copy pruning logic.
		// TODO: Is this deterministic? When pruning config changed, is it ALWAYS used by pruning algorythm immedatelly or (sometimes) on next block?
		prunedBlockIndex := blockIndex - uint32(blockKeepAmount)
		newBlocks.PruneAt(prunedBlockIndex)
	}

	oldBlock := oldBlocks.GetAt(blockIndex)

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
				Digest:   iotago.DigestFromBytes([]byte("MIGRATEDMIGRATEDMIGRATEDMIGRATED")),
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

func migrateRequestReceipts(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, enableValidations bool) {
	oldRequests := old_collections.NewMapReadOnly(oldState, old_blocklog.PrefixRequestReceipts)

	progress := NewProgressPrinter(500)
	var firstBlockIndex uint32 = math.MaxUint32
	var lastBlockIndex uint32 = 0
	var pruneBlockIndex uint32 = 0
	var totalPrunedRequests uint16 = 0
	oldRequests.Iterate(func(k, v []byte) bool {
		// Usually oldState contains mutations for just one block. But not always - e.g. with option "-i" it will contain multiple blocks.
		// We might ignore that and say we dont care about that option. But why not to implement this is it is easy?
		lookupKey := lo.Must(old_rwutil.ReadFromBytes(k, &old_blocklog.RequestLookupKey{}))

		if v == nil {
			// Request deleted - this means that block was pruned
			if totalPrunedRequests != 0 && pruneBlockIndex != lookupKey.BlockIndex() {
				panic(fmt.Errorf("unexpected multiple blocks pruned"))
			}

			pruneBlockIndex = lookupKey.BlockIndex()
			totalPrunedRequests++
			return true
		} else {
			if lookupKey.BlockIndex() < firstBlockIndex {
				firstBlockIndex = lookupKey.BlockIndex()
			}
			if lookupKey.BlockIndex() > lastBlockIndex {
				lastBlockIndex = lookupKey.BlockIndex()
			}
		}

		progress.Print()
		return true
	})

	if firstBlockIndex > lastBlockIndex {
		panic(fmt.Errorf("requestReceipts migration error: no receipts found"))
	}

	cli.DebugLogf("Migrating request receipts (%d)", oldRequests.Len())
	progress = NewProgressPrinter(500)

	// We need to go through blocks and through requests in the same way they were processed
	// to ensure, that list in values of lookup table will be generated in the same order as they
	// will appear after tracing.
	for blockIndex := firstBlockIndex; blockIndex <= lastBlockIndex; blockIndex++ {
		oldRequestsInBlock := lo.Must(getRequestReceiptsInBlock(oldState, blockIndex))

		for _, oldReceipt := range oldRequestsInBlock {
			newReceipt := migrateSingleReceipt(&oldReceipt, oldChainID)
			newLookupKey := blocklog.NewRequestLookupKey(newReceipt.BlockIndex, newReceipt.RequestIndex)
			blocklog.NewStateWriter(newState).SaveRequestReceipt(&newReceipt, newLookupKey)
			progress.Print()
		}
	}

	if totalPrunedRequests != 0 {
		// Some block was pruned - we need to prune it in the new state too.
		blocklog.NewStateWriter(newState).PruneRequestLogRecordsByBlockIndex(pruneBlockIndex, totalPrunedRequests)
		// TODO: who should call pruneEventsByBlockIndex?
	}

	newRequests := collections.NewMapReadOnly(newState, blocklog.PrefixRequestReceipts)
	if enableValidations && oldRequests.Len() != newRequests.Len() {
		panic(fmt.Errorf("requestReceipts migration error: old and new receipts count mismatch: %v != %v", oldRequests.Len(), newRequests.Len()))
	}
}

func getRequestReceiptsInBlock(partition old_kv.KVStoreReader, blockIndex uint32) ([]old_blocklog.RequestReceipt, error) {
	blockInfo, ok := old_blocklog.GetBlockInfo(partition, blockIndex)
	if !ok {
		return nil, fmt.Errorf("block not found: %d", blockIndex)
	}
	reqs := make([]old_blocklog.RequestReceipt, blockInfo.TotalRequests)
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		recBin, ok := getRequestRecordDataByRef(partition, blockIndex, reqIdx)
		if !ok {
			return nil, fmt.Errorf("request not found: %d/%d", blockIndex, reqIdx)
		}
		rec, err := old_blocklog.RequestReceiptFromBytes(recBin, blockIndex, reqIdx)
		if err != nil {
			return nil, err
		}
		reqs[reqIdx] = *rec
	}
	return reqs, nil
}

func getRequestRecordDataByRef(partition old_kv.KVStoreReader, blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := old_blocklog.NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := old_collections.NewMapReadOnly(partition, old_blocklog.PrefixRequestReceipts)
	recBin := lookupTable.GetAt(lookupKey[:])
	if recBin == nil {
		return nil, false
	}
	return recBin, true
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

func MigrateBlocklogContract(oldChainState old_kv.KVStoreReader, newChainState kv.KVStore, oldChainID old_isc.ChainID, stateMetadata *transaction.StateMetadata, chainOwner *cryptolib.Address, blockKeepAmount int32, enableValidations bool) blocklog.BlockInfo {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, blocklog.Contract.Hname())

	//printWarningsForUnprocessableRequests(oldContractState)
	blockIndex := migrateGlobalVars(oldChainState, newChainState)
	blockInfo := migrateBlockRegistry(blockIndex, blockKeepAmount, chainOwner, stateMetadata, oldContractState, newContractState)
	if blockIndex != 0 { // no requests on origin block
		migrateRequestReceipts(oldContractState, newContractState, oldChainID, enableValidations)
	}

	return blockInfo
}
