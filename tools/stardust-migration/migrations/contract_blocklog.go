package migrations

import (
	"fmt"
	"log"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/util/bcs"

	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/parameters"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

func migrateBlockRegistry(oldState old_kv.KVStoreReader, newState kv.KVStore) {
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

	fmt.Printf("Found first unpruned block at index: %d from %d blocks in total.\n", lastUnprunedBlock, oldBlocks.Len())

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

		newBlocks.SetAt(i, bcs.MustMarshal(&newBlockInfo))
	}

	fmt.Printf("blockRegistry: oldState Len: %d, newState Len: %d\n", oldBlocks.Len(), newBlocks.Len())
}

func MigrateBlocklogContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) {
	log.Print("Migrating blocklog contract\n")

	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, blocklog.Contract.Hname())

	printWarningsForUnprocessableRequests(oldContractState)
	migrateBlockRegistry(oldContractState, newContractState)
	
	log.Print("Migrated blocklog contract\n")
}

func printWarningsForUnprocessableRequests(oldState old_kv.KVStoreReader) {
	// No need to migrate. Just print a warning if there are any

	log.Printf("Listing Unprocessable Requests...\n")

	count := IterateByPrefix(oldState, old_blocklog.PrefixUnprocessableRequests, func(k isc.RequestID, v []byte) {
		log.Printf("Warning: unprocessable request found %v", k.String())
	})

	log.Printf("Listing Unprocessable Requests completed (found %v records)\n", count)
}
