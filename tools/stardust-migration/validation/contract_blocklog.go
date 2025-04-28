package validation

import (
	"fmt"
	"reflect"
	"strings"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

const blockRetentionPeriod = 10000

func oldBlocklogContractContentToStr(chainState old_kv.KVStoreReader, firstIndex, lastIndex uint32, short bool) string {
	contractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())

	var receiptsStr, blockRegistryStr, requestLookupIndexStr string
	GoAllAndWait(func() {
		receiptsStr = oldReceiptsToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old receipts preview:\n%v", utils.MultilinePreview(receiptsStr))
	}, func() {
		blockRegistryStr = oldBlockRegistryToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old block registry preview:\n%v", utils.MultilinePreview(blockRegistryStr))
	}, func() {
		requestLookupIndexStr = oldRequestLookupIndex(contractState, firstIndex, lastIndex, short)
		cli.DebugLogf("Old request lookup index preview:\n%v", utils.MultilinePreview(requestLookupIndexStr))
	})

	return receiptsStr + "\n" + blockRegistryStr + "\n" + requestLookupIndexStr
}

func oldReceiptsToStr(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)
	var requestStr strings.Builder
	receiptKeysCount := 0

	GoAllAndWait(func() {

		cli.DebugLogf("Retrieving old receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)
		reqCount := 0
		printProgress, done := NewProgressPrinter("blocklog_old", "receipts (requests)", "receipts", lastIndex-firstAvailableBlockIndex+1)
		defer done()

		for blockIndex := firstAvailableBlockIndex; blockIndex < lastIndex; blockIndex++ {
			_, requests, err := old_blocklog.GetRequestsInBlock(contractState, blockIndex)
			if err != nil {
				if strings.Contains(err.Error(), "request not found") {
					requestStr.WriteString(fmt.Sprintf("Req in block: %v: MISSING\n", blockIndex))
					continue
				}
				panic(err)
			}

			for _, req := range requests {
				// 																						^ There is no concept of RequestIndexes anymore, snip it.
				str := fmt.Sprintf("Type:%s,ID:%s,BaseToken:%d\n", reflect.TypeOf(req), req.ID().OutputID().TransactionID().ToHex(), req.Assets().BaseTokens)
				requestStr.WriteString(str)
				reqCount++
			}

			printProgress()
		}

		if uint32(reqCount) < max(lastIndex-firstAvailableBlockIndex, 1)-1 {
			panic(fmt.Sprintf("Not enough receipts found in the blocks range [%v; %v]: %v", firstIndex, lastIndex, reqCount))
		}

		cli.DebugLogf("Retrieved %v requests", reqCount)

		if firstAvailableBlockIndex > 0 && (lastIndex-firstAvailableBlockIndex) > blockRetentionPeriod {
			firstUnavailableBlockIndex := firstAvailableBlockIndex - 1
			_, _, err := old_blocklog.GetRequestsInBlock(contractState, firstUnavailableBlockIndex)
			if err == nil {
				panic(fmt.Sprintf("Block %v should be unavailable, but it is available", firstUnavailableBlockIndex))
			}
			if !strings.Contains(err.Error(), "block not found") {
				panic(err)
			}

			requestStr.WriteString(fmt.Sprintf("Last pruned block (%v): unavailable\n", firstUnavailableBlockIndex))
		}
	}, func() {
		printProgress, done := NewProgressPrinter("blocklog_old", "receipts (keys)", "keys", 0)
		defer done()

		contractState.IterateSorted(old_blocklog.PrefixRequestReceipts, func(key old_kv.Key, value []byte) bool {
			printProgress()
			key = utils.MustRemovePrefix(key, old_blocklog.PrefixRequestReceipts)
			if key == "" || key[0] == '.' {
				receiptKeysCount++
			}
			return true
		})

		if uint32(receiptKeysCount) < (lastIndex - firstAvailableBlockIndex) {
			panic(fmt.Sprintf("Not enough receipt keys found in the blocks range [%v; %v]: %v", firstIndex, lastIndex, receiptKeysCount))
		}
	})

	requestStr.WriteString(fmt.Sprintf("Receipts keys: %v\n", receiptKeysCount))

	receipts := old_collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestReceipts)
	requestStr.WriteString(fmt.Sprintf("Receipts map len: %v\n", receipts.Len()))

	return requestStr.String()
}

func newBlocklogContractContentToStr(chainState kv.KVStoreReader, firstIndex, lastIndex uint32, short bool) string {
	contractState := newstate.GetContactStateReader(chainState, blocklog.Contract.Hname())

	var receiptsStr, blockRegistryStr, requestLookupIndexStr string
	GoAllAndWait(func() {
		receiptsStr = newReceiptsToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New receipts preview:\n%v", utils.MultilinePreview(receiptsStr))
	}, func() {
		blockRegistryStr = newBlockRegistryToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New block registry preview:\n%v", utils.MultilinePreview(blockRegistryStr))
	}, func() {
		requestLookupIndexStr = newRequestLookupIndex(contractState, firstIndex, lastIndex, short)
		cli.DebugLogf("New request lookup index preview:\n%v", utils.MultilinePreview(requestLookupIndexStr))
	})

	return receiptsStr + "\n" + blockRegistryStr + "\n" + requestLookupIndexStr
}

func newReceiptsToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var requestStr strings.Builder
	receiptKeysCount := 0

	GoAllAndWait(func() {
		var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)

		cli.DebugLogf("Retrieving new receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)
		reqCount := 0
		printProgress, done := NewProgressPrinter("blocklog_new", "receipts", "receipts", lastIndex-firstAvailableBlockIndex+1)
		defer done()

		for blockIndex := firstAvailableBlockIndex; blockIndex < lastIndex; blockIndex++ {
			_, requests, err := blocklog.NewStateReader(contractState).GetRequestsInBlock(blockIndex)
			if err != nil {
				if strings.Contains(err.Error(), "request not found") {
					requestStr.WriteString(fmt.Sprintf("Req in block: %v: MISSING\n", blockIndex))
					continue
				}
				panic(err)
			}

			for _, req := range requests {
				str := fmt.Sprintf("Type:%s,ID:%s,BaseToken:%d\n", reflect.TypeOf(req), req.ID(), req.Assets().BaseTokens()/1000) // Base token conversion 9=>6
				requestStr.WriteString(str)
				reqCount++
			}

			printProgress()
		}

		cli.DebugLogf("Retrieved %v requests", reqCount)

		if firstAvailableBlockIndex > 0 && (lastIndex-firstAvailableBlockIndex) > blockRetentionPeriod {
			firstUnavailableBlockIndex := firstAvailableBlockIndex - 1
			_, _, err := blocklog.NewStateReader(contractState).GetRequestsInBlock(firstUnavailableBlockIndex)
			requestStr.WriteString(fmt.Sprintf("Last pruned block (%v): %v\n",
				firstUnavailableBlockIndex, lo.Ternary(err == nil, "AVAILABLE", "unavailable")))

		}
	}, func() {
		printProgress, done := NewProgressPrinter("blocklog_new", "receipts (keys)", "keys", 0)
		defer done()

		contractState.IterateSorted(kv.Key(blocklog.PrefixRequestReceipts), func(key kv.Key, value []byte) bool {
			printProgress()
			key = utils.MustRemovePrefix(key, blocklog.PrefixRequestReceipts)
			if key == "" || key[0] == '.' {
				receiptKeysCount++
			}
			return true
		})
	})

	requestStr.WriteString(fmt.Sprintf("Receipts keys: %v\n", receiptKeysCount))

	receipts := collections.NewMapReadOnly(contractState, blocklog.PrefixRequestReceipts)
	requestStr.WriteString(fmt.Sprintf("Receipts map len: %v\n", receipts.Len()))

	return requestStr.String()
}

func getFirstAvailableBlockIndex(firstIndexHint uint32, lastIndex uint32) uint32 {
	if lastIndex < blockRetentionPeriod {
		return 0
	}
	return max(lastIndex-blockRetentionPeriod+1, firstIndexHint)
}

func oldBlockRegistryToStr(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	firstAvailBlockIndex := getFirstAvailableBlockIndex(firstIndex, lastIndex)
	var blocksStr strings.Builder
	blockRegistryKeys := 0

	GoAllAndWait(func() {
		cli.DebugLogf("Retrieving old blocks: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
		blocks := old_collections.NewArrayReadOnly(contractState, old_blocklog.PrefixBlockRegistry)
		printProgress, done := NewProgressPrinter("blocklog_old", "block registry (blocks)", "blocks", lastIndex-firstAvailBlockIndex+1)
		defer done()

		blocksStr.WriteString(fmt.Sprintf("Block registry virtual len: %v\n", blocks.Len()))

		for blockIndex := firstAvailBlockIndex; blockIndex < lastIndex; blockIndex++ {
			blockBytes := blocks.GetAt(blockIndex)
			if blockBytes == nil {
				blocksStr.WriteString(fmt.Sprintf("Block %v: MISSING\n", blockIndex))
				continue
			}

			block := lo.Must(old_blocklog.BlockInfoFromBytes(blockBytes))
			blocksStr.WriteString(fmt.Sprintf("Block %v: %s\n", blockIndex, oldBlockToStr(block)))
			printProgress()
		}

		cli.DebugLogf("Retrieved %v blocks", lastIndex-firstAvailBlockIndex)
		if firstAvailBlockIndex > 0 && (lastIndex-firstAvailBlockIndex) > blockRetentionPeriod {
			firstUnavailableBlockIndex := firstAvailBlockIndex - 1
			blockBytes := blocks.GetAt(firstUnavailableBlockIndex)
			if blockBytes != nil {
				panic(fmt.Sprintf("Block %v should be unavailable, but it is available", firstUnavailableBlockIndex))
			}
			blocksStr.WriteString(fmt.Sprintf("Block (last pruned): %v: unavailable\n", firstUnavailableBlockIndex))
		}
	}, func() {
		if firstAvailBlockIndex != 0 {
			// Corresponding migration does not work with -i option, so just skipping it
			return
		}

		cli.DebugLogf("Retrieving old block registry keys: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
		printProgress, done := NewProgressPrinter("blocklog_old", "block registry (keys)", "keys", 0)
		defer done()

		contractState.IterateSorted(old_blocklog.PrefixBlockRegistry, func(key old_kv.Key, value []byte) bool {
			printProgress()
			key = utils.MustRemovePrefix(key, old_blocklog.PrefixBlockRegistry)
			if key == "" || key[0] == '#' {
				blockRegistryKeys++
			}
			return true
		})

		if uint32(blockRegistryKeys) < (lastIndex - firstAvailBlockIndex) {
			panic(fmt.Sprintf("Not enough block registry keys found in the blocks range [%v; %v]: %v", firstIndex, lastIndex, blockRegistryKeys))
		}

		cli.DebugLogf("Retrieved %v old block registry keys", blockRegistryKeys)
	})

	blocksStr.WriteString(fmt.Sprintf("Block registry keys: %v\n", blockRegistryKeys))

	return blocksStr.String()
}

func newBlockRegistryToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	firstAvailBlockIndex := getFirstAvailableBlockIndex(firstIndex, lastIndex)
	var blocksStr strings.Builder
	blockRegistryKeys := 0

	GoAllAndWait(func() {
		cli.DebugLogf("Retrieving new blocks: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
		blocks := blocklog.NewStateReader(contractState).GetBlockRegistry()
		printProgress, done := NewProgressPrinter("blocklog_new", "block registry (blocks)", "blocks", lastIndex-firstAvailBlockIndex+1)
		defer done()

		blocksStr.WriteString(fmt.Sprintf("Block registry virtual len: %v\n", blocks.Len()))

		for blockIndex := firstAvailBlockIndex; blockIndex < lastIndex; blockIndex++ {
			blockBytes := blocks.GetAt(blockIndex)
			if blockBytes == nil {
				blocksStr.WriteString(fmt.Sprintf("Block %v: MISSING\n", blockIndex))
				continue
			}

			block := lo.Must(blocklog.BlockInfoFromBytes(blockBytes))
			blocksStr.WriteString(fmt.Sprintf("Block %v: %s\n", blockIndex, newBlockToStr(block)))
			printProgress()
		}

		cli.DebugLogf("Retrieved %v blocks", lastIndex-firstAvailBlockIndex)
		if firstAvailBlockIndex > 0 && (lastIndex-firstAvailBlockIndex) > blockRetentionPeriod {
			firstUnavailableBlockIndex := firstAvailBlockIndex - 1
			blockBytes := blocks.GetAt(firstUnavailableBlockIndex)
			blocksStr.WriteString(fmt.Sprintf("Block (last pruned): %v: %v\n",
				firstUnavailableBlockIndex, lo.Ternary(blockBytes == nil, "unavailable", "AVAILABLE")))
		}
	}, func() {
		if firstAvailBlockIndex != 0 {
			// Corresponding migration does not work with -i option, so just skipping it
			return
		}

		cli.DebugLogf("Retrieving new block registry keys: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
		printProgress, done := NewProgressPrinter("blocklog_new", "block registry (keys)", "keys", 0)
		defer done()

		contractState.IterateSorted(kv.Key(blocklog.PrefixBlockRegistry), func(key kv.Key, value []byte) bool {
			printProgress()
			key = utils.MustRemovePrefix(key, blocklog.PrefixBlockRegistry)
			if key == "" || key[0] == '#' {
				blockRegistryKeys++
			}
			return true
		})

		cli.DebugLogf("Retrieved %v new block registry keys", blockRegistryKeys)
	})

	blocksStr.WriteString(fmt.Sprintf("Block registry keys: %v\n", blockRegistryKeys))

	return blocksStr.String()
}

func oldRequestLookupIndex(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32, short bool) string {
	firstAvailBlockIndex := max(getFirstAvailableBlockIndex(firstIndex, lastIndex), 1)
	lookupTable := old_collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestLookupIndex)
	cli.DebugLogf("Retrieving old request lookup index: blocks = [%v; %v]", firstAvailBlockIndex, lastIndex)

	// The table is too big to be stringified - so just checking stringifying requests from last 10000 blocks,
	// plus gathering general statistics

	var indexStr = strings.Builder{}
	var keysCount uint32 = 0
	var listElementsCount = 0

	GoAllAndWait(func() {
		requestsCount := 0
		printProgress, done := NewProgressPrinter("blocklog_old", "lookup (requests)", "requests", lastIndex-firstAvailBlockIndex+1)
		defer done()

		for blockIndex := firstAvailBlockIndex; blockIndex < lastIndex; blockIndex++ {
			printProgress()
			_, reqs := lo.Must2(old_blocklog.GetRequestsInBlock(contractState, blockIndex))
			requestsCount += len(reqs)

			for reqIdx, req := range reqs {
				lookupKey := req.ID().LookupDigest()
				lookupKeyListBytes := lookupTable.GetAt(lookupKey[:])
				if lookupKeyListBytes == nil {
					panic(fmt.Sprintf("Req lookup %v/%v: NO LOOKUP RECORD\n", blockIndex, reqIdx))
				}

				lookupKeyList := lo.Must(old_blocklog.RequestLookupKeyListFromBytes(lookupKeyListBytes))
				found := false
				for _, lookupRecord := range lookupKeyList {
					if lookupRecord.BlockIndex() == blockIndex || lookupRecord.RequestIndex() == uint16(reqIdx) {
						indexStr.WriteString(fmt.Sprintf("Req lookup %v/%v: found\n", blockIndex, reqIdx))
						found = true
						break
					}
				}

				if !found {
					panic(fmt.Sprintf("Req lookup %v/%v: REQUEST NOT FOUND\n", blockIndex, reqIdx))
				}
			}
		}

		if uint32(requestsCount) < max(lastIndex-firstAvailBlockIndex, 1)-1 && lastIndex != 0 {
			panic(fmt.Sprintf("Not enough requests found in the blocks range [%v; %v]: %v", firstAvailBlockIndex, lastIndex, requestsCount))
		}

		cli.DebugLogf("Retrieved %v old request lookup entries", requestsCount)
	}, func() {
		if short {
			cli.DebugLogf("Skipping old request lookup index keys retrieval")
			return
		}

		printProgress, done := NewProgressPrinter("blocklog_old", "lookup (elements)", "keys", 0)
		defer done()

		contractState.IterateSorted(old_blocklog.PrefixRequestLookupIndex, func(key old_kv.Key, value []byte) bool {
			printProgress()
			key = utils.MustRemovePrefix(key, old_blocklog.PrefixRequestLookupIndex)
			if key == "" {
				keysCount++
				listElementsCount++
			} else if key[0] == '.' {
				keysCount++
				lookupKeyList := lo.Must(old_blocklog.RequestLookupKeyListFromBytes(value))
				listElementsCount += len(lookupKeyList)
			}
			return true
		})

		if lastIndex != 0 {
			if keysCount != lookupTable.Len()+1 { // +1 because of length key
				panic(fmt.Sprintf("Request lookup index keys count mismatch: %v != %v", keysCount, lookupTable.Len()+1))
			}
			if uint32(listElementsCount) < (lastIndex - firstAvailBlockIndex) {
				panic(fmt.Sprintf("Not enough request lookup index keys found in the blocks range [%v; %v]: %v",
					firstIndex, lastIndex, listElementsCount))
			}
		}

		cli.DebugLogf("Retrieved %v old request lookup index keys", listElementsCount)
	})

	indexStr.WriteString(fmt.Sprintf("Request lookup index keys: %v\n", listElementsCount))
	indexStr.WriteString(fmt.Sprintf("Request lookup index map len: %v\n", lookupTable.Len()))

	return indexStr.String()
}

func newRequestLookupIndex(contractState kv.KVStoreReader, firstIndex, lastIndex uint32, short bool) string {
	firstAvailBlockIndex := max(getFirstAvailableBlockIndex(firstIndex, lastIndex), 1)
	lookupTable := collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestLookupIndex)
	cli.DebugLogf("Retrieving new request lookup index: blocks = [%v; %v]", firstAvailBlockIndex, lastIndex)

	// The table is too big to be stringified - so just checking stringifying requests from last 10000 blocks,
	// plus gathering general statistics

	var indexStr strings.Builder
	var elementsCount = 0

	GoAllAndWait(func() {
		requestsCount := 0
		printProgress, done := NewProgressPrinter("blocklog_new", "lookup (requests)", "requests", lastIndex-firstAvailBlockIndex+1)
		defer done()
		r := blocklog.NewStateReader(contractState)

		for blockIndex := firstAvailBlockIndex; blockIndex < lastIndex; blockIndex++ {
			printProgress()
			_, reqs := lo.Must2(r.GetRequestsInBlock(blockIndex))
			requestsCount += len(reqs)

			for reqIdx, req := range reqs {
				lookupKey := req.ID().LookupDigest()
				lookupKeyListBytes := lookupTable.GetAt(lookupKey[:])
				if lookupKeyListBytes == nil {
					indexStr.WriteString(fmt.Sprintf("Req lookup %v/%v: NO LOOKUP RECORD\n", blockIndex, reqIdx))
					continue
				}

				lookupKeyList := lo.Must(blocklog.RequestLookupKeyListFromBytes(lookupKeyListBytes))
				found := false
				for _, lookupRecord := range lookupKeyList {
					if lookupRecord.BlockIndex() == blockIndex || lookupRecord.RequestIndex() == uint16(reqIdx) {
						indexStr.WriteString(fmt.Sprintf("Req lookup %v/%v: found\n", blockIndex, reqIdx))
						found = true
						break
					}
				}

				if !found {
					indexStr.WriteString(fmt.Sprintf("Req lookup %v/%v: REQUEST NOT FOUND\n", blockIndex, reqIdx))
				}
			}
		}

		cli.DebugLogf("Retrieved %v new request lookup entries", requestsCount)
	}, func() {
		if short {
			cli.DebugLogf("Skipping old request lookup index keys retrieval")
			return
		}

		printProgress, done := NewProgressPrinter("blocklog_new", "lookup (elements)", "keys", 0)
		defer done()

		contractState.IterateSorted(kv.Key(blocklog.PrefixRequestLookupIndex), func(key kv.Key, value []byte) bool {
			printProgress()
			if key == kv.Key(blocklog.PrefixRequestLookupIndex) {
				elementsCount++
			} else {
				lookupKeyList := lo.Must(blocklog.RequestLookupKeyListFromBytes(value))
				elementsCount += len(lookupKeyList)
			}
			return true
		})

		if lastIndex != 0 && elementsCount == 0 {
			panic(fmt.Sprintf("No request lookup index keys found in the blocks range [%v; %v]", firstIndex, lastIndex))
		}

		cli.DebugLogf("Retrieved %v new request lookup index keys", elementsCount)
	})

	indexStr.WriteString(fmt.Sprintf("Request lookup index keys: %v\n", elementsCount))
	indexStr.WriteString(fmt.Sprintf("Request lookup index map len: %v\n", lookupTable.Len()))

	return indexStr.String()
}
