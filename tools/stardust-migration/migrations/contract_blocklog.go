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
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"

	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

func migrateBlockRegistry(blockIndex uint32, previousL1Commitment *state.L1Commitment, oldState old_kv.KVStoreReader, newState kv.KVStore) blocklog.BlockInfo {
	newBlocks := collections.NewArray(newState, blocklog.PrefixBlockRegistry)
	newBlocks.SetSize(blockIndex + 1)

	oldBlock := old_collections.NewArrayReadOnly(oldState, old_blocklog.PrefixBlockRegistry).GetAt(blockIndex)

	// TODO: Find a good solution for the PreviousAnchor that we don't have. Can it just be zero'd?
	// TODO: Find proper default values for the L1Params
	// Most likely want to pull the params once and use it for all blocks?
	// Probably not super needed, but would be good to have some realistic data inside there.
	// -> Same for the MaxPayloadSize
	const MAX_PAYLOAD_SIZE = 1024
	defaultL1Params := &parameters.L1Params{
		BaseToken: &parameters.BaseToken{
			Name:            isc.BaseTokenCoinInfo.Name,
			TickerSymbol:    isc.BaseTokenCoinInfo.Symbol,
			Unit:            "",
			Subunit:         "",
			Decimals:        isc.BaseTokenCoinInfo.Decimals,
			UseMetricPrefix: false,
			CoinType:        coin.BaseTokenType,
			TotalSupply:     isc.BaseTokenCoinInfo.TotalSupply.Uint64(),
		},
		Protocol: &parameters.Protocol{
			Epoch:                 iotajsonrpc.NewBigInt(0),
			ProtocolVersion:       iotajsonrpc.NewBigInt(0),
			SystemStateVersion:    iotajsonrpc.NewBigInt(0),
			IotaTotalSupply:       iotajsonrpc.NewBigInt(0),
			ReferenceGasPrice:     iotajsonrpc.NewBigInt(0),
			EpochStartTimestampMs: iotajsonrpc.NewBigInt(0),
			EpochDurationMs:       iotajsonrpc.NewBigInt(0),
		},
	}

	oldBlockInfo, err := old_blocklog.BlockInfoFromBytes(oldBlock)
	if err != nil {
		panic(fmt.Errorf("blockRegistry migration error: %v", err))
	}

	if previousL1Commitment == nil {
		previousL1Commitment = &state.L1Commitment{}
	}

	metadata := transaction.StateMetadata{
		SchemaVersion:   isc.SchemaVersion(oldBlockInfo.SchemaVersion),
		L1Commitment:    previousL1Commitment,   // TODO: It's very important that we properly handle the L1Commitment here as it will be used for tracing
		GasCoinObjectID: &iotago.ObjectID{},     // This can probably be zero'd, maybe we pass the real chains GasCoin here otherwise?
		GasFeePolicy:    gas.DefaultFeePolicy(), // TODO: We probably need to grab this from the governance contract
		InitParams:      nil,                    // TODO: This will be fun to figure out
		InitDeposit:     0,                      // TODO: And this too, we probably need to pass this value upon deployment
		PublicURL:       "",
	}

	// The following two structs will need some love regarding IDs, but we probably can fake most of them
	migrationStateAnchor := isc.NewStateAnchor(&iscmove.AnchorWithRef{
		Object: &iscmove.Anchor{
			Assets: iscmove.Referent[iscmove.AssetsBag]{ID: iotago.ObjectID{}, Value: &iscmove.AssetsBag{
				ID:   iotago.ObjectID{},
				Size: 0,
			}},
			ID:            iotago.ObjectID{},
			StateIndex:    blockIndex,
			StateMetadata: metadata.Bytes(),
		},
		ObjectRef: iotago.ObjectRef{
			ObjectID: &iotago.ObjectID{},
			Digest:   iotago.DigestFromBytes([]byte("MIGRATED")),
			Version:  0,
		},
		Owner: &iotago.Address{},
	}, iotago.PackageID{})

	newBlockInfo := blocklog.BlockInfo{
		GasFeeCharged:         coin.Value(oldBlockInfo.GasFeeCharged),
		GasBurned:             oldBlockInfo.GasBurned,
		BlockIndex:            oldBlockInfo.BlockIndex(),
		NumOffLedgerRequests:  oldBlockInfo.NumOffLedgerRequests,
		NumSuccessfulRequests: oldBlockInfo.NumSuccessfulRequests,
		SchemaVersion:         oldBlockInfo.SchemaVersion,
		Timestamp:             oldBlockInfo.Timestamp,
		TotalRequests:         oldBlockInfo.TotalRequests,
		PreviousAnchor:        &migrationStateAnchor,
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

func migrateSingleReceipt(receipt *old_blocklog.RequestReceipt, oldChainID old_isc.ChainID, newChainID isc.ChainID) blocklog.RequestReceipt {
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
		Request:       MigrateSingleRequest(receipt.Request, oldChainID, newChainID),
		Error:         receiptError,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: coin.Value(receipt.GasFeeCharged),
		GasBurnLog:    burnLog,
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
	}
}

func migrateRequestReceipts(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, newChainID isc.ChainID) {
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

			newReceipt := migrateSingleReceipt(oldReceipt, oldChainID, newChainID)
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

func MigrateBlocklogContract(oldChainState old_kv.KVStoreReader, newChainState kv.KVStore, oldChainID old_isc.ChainID, newChainID isc.ChainID, previousL1Commitment *state.L1Commitment) blocklog.BlockInfo {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, blocklog.Contract.Hname())

	//printWarningsForUnprocessableRequests(oldContractState)
	blockIndex := migrateGlobalVars(oldChainState, newChainState)
	blockInfo := migrateBlockRegistry(blockIndex, previousL1Commitment, oldContractState, newContractState)
	migrateRequestLookupIndex(oldContractState, newContractState)
	migrateRequestReceipts(oldContractState, newContractState, oldChainID, newChainID)

	return blockInfo
}
