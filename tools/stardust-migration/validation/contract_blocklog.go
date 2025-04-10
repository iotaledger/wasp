package validation

import (
	"fmt"
	"reflect"
	"strings"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

const blockRetentionPeriod = 10000

func OldBlocklogContractContentToStr(chainState old_kv.KVStoreReader, chainID old_isc.ChainID, firstIndex, lastIndex uint32) string {
	contractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())

	var receiptsStr, blockRegistryStr, requestLookupIndexStr string
	GoAllAndWait(func() {
		receiptsStr = oldReceiptsToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old receipts preview:\n%v", utils.MultilinePreview(receiptsStr))
	}, func() {
		blockRegistryStr = oldBlockRegistryToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old block registry preview:\n%v", utils.MultilinePreview(blockRegistryStr))
	}, func() {
		requestLookupIndexStr = oldRequestLookupIndex(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old request lookup index preview:\n%v", utils.MultilinePreview(requestLookupIndexStr))
	})

	return receiptsStr + blockRegistryStr + requestLookupIndexStr
}

func oldReceiptsToStr(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)

	cli.DebugLogf("Retrieving old receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)
	var requestStr strings.Builder
	reqCount := 0
	printProgress, done := NewProgressPrinter("old_blocklog", "receipts", "receipts", lastIndex-firstAvailableBlockIndex+1)
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

	if reqCount == 0 {
		panic(fmt.Sprintf("No requests found in the blocks range [%v; %v]", firstIndex, lastIndex))
	}

	cli.DebugLogf("Retrieved %v requests", reqCount)

	if firstAvailableBlockIndex > 0 {
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

	return requestStr.String()
}

func NewBlocklogContractContentToStr(chainState kv.KVStoreReader, chainID isc.ChainID, firstIndex, lastIndex uint32) string {
	contractState := newstate.GetContactStateReader(chainState, blocklog.Contract.Hname())

	var receiptsStr, blockRegistryStr, requestLookupIndexStr string
	GoAllAndWait(func() {
		receiptsStr = newReceiptsToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New receipts preview:\n%v", utils.MultilinePreview(receiptsStr))
	}, func() {
		blockRegistryStr = newBlockRegistryToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New block registry preview:\n%v", utils.MultilinePreview(blockRegistryStr))
	}, func() {
		requestLookupIndexStr = newRequestLookupIndex(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New request lookup index preview:\n%v", utils.MultilinePreview(requestLookupIndexStr))
	})

	return receiptsStr + blockRegistryStr + requestLookupIndexStr
}

func newReceiptsToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)

	cli.DebugLogf("Retrieving new receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)
	var requestStr strings.Builder
	reqCount := 0
	printProgress, done := NewProgressPrinter("new_blocklog", "receipts", "receipts", lastIndex-firstAvailableBlockIndex+1)
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

	if firstAvailableBlockIndex > 0 {
		firstUnavailableBlockIndex := firstAvailableBlockIndex - 1
		_, _, err := blocklog.NewStateReader(contractState).GetRequestsInBlock(firstUnavailableBlockIndex)
		requestStr.WriteString(fmt.Sprintf("Last pruned block (%v): %v\n",
			firstUnavailableBlockIndex, lo.Ternary(err == nil, "AVAILABLE", "unavailable")))

	}

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
		printProgress, done := NewProgressPrinter("old_blocklog", "block registry (blocks)", "blocks", lastIndex-firstAvailBlockIndex+1)
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
		if firstAvailBlockIndex > 0 {
			firstUnavailableBlockIndex := firstAvailBlockIndex - 1
			blockBytes := blocks.GetAt(firstUnavailableBlockIndex)
			if blockBytes != nil {
				panic(fmt.Sprintf("Block %v should be unavailable, but it is available", firstUnavailableBlockIndex))
			}
			blocksStr.WriteString(fmt.Sprintf("Block (last pruned): %v: unavailable\n", firstUnavailableBlockIndex))
		}
	}, func() {
		cli.DebugLogf("Retrieving old block registry keys: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
		printProgress, done := NewProgressPrinter("old_blocklog", "block registry (keys)", "keys", 0)
		defer done()

		contractState.Iterate(old_blocklog.PrefixBlockRegistry, func(key old_kv.Key, value []byte) bool {
			blockRegistryKeys++
			printProgress()
			return true
		})

		if blockRegistryKeys == 0 {
			panic(fmt.Sprintf("No block registry keys found in the blocks range [%v; %v]", firstIndex, lastIndex))
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
		printProgress, done := NewProgressPrinter("new_blocklog", "block registry (blocks)", "blocks", lastIndex-firstAvailBlockIndex+1)
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
		if firstAvailBlockIndex > 0 {
			firstUnavailableBlockIndex := firstAvailBlockIndex - 1
			blockBytes := blocks.GetAt(firstUnavailableBlockIndex)
			blocksStr.WriteString(fmt.Sprintf("Block (last pruned): %v: %v\n",
				firstUnavailableBlockIndex, lo.Ternary(blockBytes == nil, "unavailable", "AVAILABLE")))
		}
	}, func() {
		cli.DebugLogf("Retrieving new block registry keys: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
		printProgress, done := NewProgressPrinter("blocklog", "block registry (keys)", "keys", 0)
		defer done()

		contractState.Iterate(kv.Key(blocklog.PrefixBlockRegistry), func(key kv.Key, value []byte) bool {
			blockRegistryKeys++
			printProgress()
			return true
		})

		cli.DebugLogf("Retrieved %v new block registry keys", blockRegistryKeys)
	})

	blocksStr.WriteString(fmt.Sprintf("Block registry keys: %v\n", blockRegistryKeys))

	return blocksStr.String()
}

func oldRequestLookupIndex(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	firstAvailBlockIndex := max(getFirstAvailableBlockIndex(firstIndex, lastIndex), 1)
	lookupTable := old_collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestLookupIndex)
	cli.DebugLogf("Retrieving old request lookup index: blocks = [%v; %v]", firstAvailBlockIndex, lastIndex)

	// The table is too big to be stringified - so just checking stringifying requests from last 10000 blocks,
	// plus gathering general statistics

	var indexStr = strings.Builder{}
	var elementsCount = 0

	GoAllAndWait(func() {
		requestsCount := 0
		printProgress, done := NewProgressPrinter("old_blocklog", "lookup (requests)", "requests", lastIndex-firstAvailBlockIndex+1)
		defer done()

		for blockIndex := max(firstAvailBlockIndex, 1); blockIndex < lastIndex; blockIndex++ {
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

		if requestsCount == 0 {
			panic(fmt.Sprintf("No requests found in the blocks range [%v; %v]", firstIndex, lastIndex))
		}

		cli.DebugLogf("Retrieved %v old request lookup entries", requestsCount)
	}, func() {
		printProgress, done := NewProgressPrinter("old_blocklog", "lookup (elements)", "keys", 0)
		defer done()

		contractState.Iterate(old_blocklog.PrefixRequestLookupIndex, func(key old_kv.Key, value []byte) bool {
			key = utils.MustRemovePrefix(key, old_blocklog.PrefixRequestLookupIndex)
			if len(key) == 0 {
				return true
			}

			lookupKeyList := lo.Must(old_blocklog.RequestLookupKeyListFromBytes(value))
			elementsCount += len(lookupKeyList)
			printProgress()

			return true
		})

		if elementsCount == 0 {
			panic(fmt.Sprintf("No request lookup index keys found in the blocks range [%v; %v]", firstIndex, lastIndex))
		}

		cli.DebugLogf("Retrieved %v old request lookup index keys", elementsCount)
	})

	indexStr.WriteString(fmt.Sprintf("Request lookup index keys: %v\n", elementsCount))

	return indexStr.String()
}

func newRequestLookupIndex(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	firstAvailBlockIndex := max(getFirstAvailableBlockIndex(firstIndex, lastIndex), 1)
	lookupTable := collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestLookupIndex)
	cli.DebugLogf("Retrieving new request lookup index: blocks = [%v; %v]", firstAvailBlockIndex, lastIndex)

	// The table is too big to be stringified - so just checking stringifying requests from last 10000 blocks,
	// plus gathering general statistics

	var indexStr strings.Builder
	var elementsCount = 0

	GoAllAndWait(func() {
		requestsCount := 0
		printProgress, done := NewProgressPrinter("new_blocklog", "lookup (requests)", "requests", lastIndex-firstAvailBlockIndex+1)
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
		printProgress, done := NewProgressPrinter("blocklog", "lookup (elements)", "keys", 0)
		defer done()

		contractState.Iterate(kv.Key(blocklog.PrefixRequestLookupIndex), func(key kv.Key, value []byte) bool {
			key = utils.MustRemovePrefix(key, old_blocklog.PrefixRequestLookupIndex)
			if len(key) == 0 {
				return true
			}

			lookupKeyList := lo.Must(blocklog.RequestLookupKeyListFromBytes(value))
			elementsCount += len(lookupKeyList)
			printProgress()

			return true
		})

		if elementsCount == 0 {
			panic(fmt.Sprintf("No request lookup index keys found in the blocks range [%v; %v]", firstIndex, lastIndex))
		}

		cli.DebugLogf("Retrieved %v new request lookup index keys", elementsCount)
	})

	indexStr.WriteString(fmt.Sprintf("Request lookup index keys: %v\n", elementsCount))

	return indexStr.String()
}
