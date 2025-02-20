package migrations

import (
	"fmt"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/vm/gas"

	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

func migrateBlockRegistry(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	cli.Log("Migrating block registry")
	oldBlocks := old_collections.NewArrayReadOnly(oldState, old_blocklog.PrefixBlockRegistry)
	newBlocks := collections.NewArray(newState, blocklog.PrefixBlockRegistry)
	newBlocks.SetSize(oldBlocks.Len())

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
		MaxPayloadSize: MAX_PAYLOAD_SIZE,
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

	lastUnprunedBlock := -1

	// Due to pruning, we have an array length of > 4 million, but only 10-20k entries with block data.
	// This will search backwards to find the oldest block to prune, so we don't have to iterate 4 million empty items.
	for i := oldBlocks.Len() - 1; i > 0; i-- {
		oldData := oldBlocks.GetAt(i)
		if len(oldData) > 0 {
			lastUnprunedBlock = int(i)
			continue
		}

		break
	}

	cli.Logf("blockRegistry: Found first unpruned block at index: %d from %d blocks in total.\n", lastUnprunedBlock, oldBlocks.Len())
	progress := NewProgressPrinter(500)

	for i := uint32(lastUnprunedBlock); i < oldBlocks.Len(); i++ {
		oldData := oldBlocks.GetAt(i)
		if len(oldData) == 0 {
			panic(fmt.Errorf("blockRegistry migration error: %d has empty data!\n", i))
		}

		oldBlockInfo, err := old_blocklog.BlockInfoFromBytes(oldData)
		if err != nil {
			panic(fmt.Errorf("blockRegistry migration error: %v", err))
		}

		newBlockInfo := blocklog.BlockInfo{
			GasFeeCharged:         coin.Value(oldBlockInfo.GasFeeCharged),
			GasBurned:             oldBlockInfo.GasBurned,
			BlockIndex:            oldBlockInfo.BlockIndex(),
			NumOffLedgerRequests:  oldBlockInfo.NumOffLedgerRequests,
			NumSuccessfulRequests: oldBlockInfo.NumSuccessfulRequests,
			SchemaVersion:         oldBlockInfo.SchemaVersion,
			Timestamp:             oldBlockInfo.Timestamp,
			TotalRequests:         oldBlockInfo.TotalRequests,
			PreviousAnchor:        nil,
			L1Params:              defaultL1Params,
		}

		newBlocks.SetAt(i, newBlockInfo.Bytes())
		progress.Print()
	}

	cli.Logf("\nblockRegistry: oldState Len: %d, newState Len: %d\n", oldBlocks.Len(), newBlocks.Len())
}

func migrateRequestLookupIndex(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	cli.Log("Migrating request lookup index")

	oldLookup := old_collections.NewMapReadOnly(oldState, old_blocklog.PrefixRequestLookupIndex)
	newLookup := collections.NewMap(newState, blocklog.PrefixRequestLookupIndex)

	progress := NewProgressPrinter(500)
	oldLookup.IterateKeys(func(elemKey []byte) bool {
		oldIndex := oldLookup.GetAt(elemKey)
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
		progress.Print()
		return true
	})
}

func migrateSingleRequest(req old_isc.Request) isc.Request {
	switch req.(type) {
	case old_isc.OnLedgerRequest:
		break

	case old_isc.OffLedgerRequest:
		break

	case old_isc.UnsignedOffLedgerRequest:
		break

	case old_isc.ImpersonatedOffLedgerRequest:
		break

	default:
		panic(fmt.Errorf("migrateSingleRequest: invalid request type: %T", req))
	}

	return nil
}

func migrateSingleReceipt(receipt *old_blocklog.RequestReceipt) blocklog.RequestReceipt {
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
		Request:       migrateSingleRequest(receipt.Request),
		Error:         receiptError,
		GasBudget:     receipt.GasBudget,
		GasBurned:     receipt.GasBurned,
		GasFeeCharged: coin.Value(receipt.GasFeeCharged),
		GasBurnLog:    burnLog,
		BlockIndex:    receipt.BlockIndex,
		RequestIndex:  receipt.RequestIndex,
	}
}

type lutCollection struct {
	k []byte
	v []byte
}

func migrateRequestReceipts(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	oldRequests := old_collections.NewMapReadOnly(oldState, old_blocklog.PrefixRequestReceipts)
	oldLookup := old_collections.NewMapReadOnly(oldState, old_blocklog.PrefixRequestLookupIndex)

	cli.Logf("Migrating request receipts (%d)\n", oldRequests.Len())

	_ = collections.NewMap(newState, blocklog.PrefixRequestReceipts)

	progress := NewProgressPrinter(500)

	oldLUT := make([]old_blocklog.RequestLookupKeyList, 0)
	oldLookup.Iterate(func(elemKey []byte, value []byte) bool {
		oldLookupKeys, err := old_blocklog.RequestLookupKeyListFromBytes(value)
		if err != nil {
			panic(fmt.Errorf("requestReceipts migration error: %v", err))
		}

		oldLUT = append(oldLUT, oldLookupKeys)

		progress.Print()

		return true
	})

	progress = NewProgressPrinter(500)

	for _, l := range oldLUT {
		for _, k := range l {
			oldReceipt, err := old_blocklog.RequestReceiptFromBytes(oldRequests.GetAt(k.Bytes()), k.BlockIndex(), k.RequestIndex())
			if err != nil {
				panic(fmt.Errorf("requestReceipt migration error: %v", err))
			}

			migrateSingleReceipt(oldReceipt)
			progress.Print()
		}
	}
}

func printWarningsForUnprocessableRequests(oldState old_kv.KVStoreReader) {
	// No need to migrate. Just print a warning if there are any

	cli.Logf("Listing Unprocessable Requests...\n")

	count := IterateByPrefix(oldState, old_blocklog.PrefixUnprocessableRequests, func(k isc.RequestID, v []byte) {
		cli.Logf("Warning: unprocessable request found %v", k.String())
	})

	cli.Logf("Listing Unprocessable Requests completed (found %v records)\n", count)
}

func MigrateBlocklogContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) {
	cli.Log("Migrating blocklog contract\n")

	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, blocklog.Contract.Hname())

	printWarningsForUnprocessableRequests(oldContractState)
	//migrateBlockRegistry(oldContractState, newContractState)
	//migrateRequestLookupIndex(oldContractState, newContractState)
	migrateRequestReceipts(oldContractState, newContractState)

	cli.Log("Migrated blocklog contract\n")
}
