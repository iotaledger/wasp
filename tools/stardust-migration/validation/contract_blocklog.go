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
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

const blockRetentionPeriod = 10000

func OldBlocklogContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, firstIndex, lastIndex uint32) string {
	var receiptsStr, blockRegistryStr, requestLookupIndexStr string
	GoAllAndWait(func() {
		receiptsStr = OldReceiptsContentToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old receipts preview:\n%v\n", utils.MultilinePreview(receiptsStr))
	}, func() {
		blockRegistryStr = oldBlockRegistryToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old block registry preview:\n%v\n", utils.MultilinePreview(blockRegistryStr))
	}, func() {
		requestLookupIndexStr = oldRequestLookupIndex(contractState, firstIndex, lastIndex)
		cli.DebugLogf("Old request lookup index preview:\n%v\n", utils.MultilinePreview(requestLookupIndexStr))
	})

	return receiptsStr + "\n" + blockRegistryStr + "\n" + requestLookupIndexStr
}

func OldReceiptsContentToStr(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
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

func NewBlocklogContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID, firstIndex, lastIndex uint32) string {
	var receiptsStr, blockRegistryStr, requestLookupIndexStr string
	GoAllAndWait(func() {
		receiptsStr = NewReceiptsContentToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New receipts preview:\n%v\n", utils.MultilinePreview(receiptsStr))
	}, func() {
		blockRegistryStr = newBlockRegistryToStr(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New block registry preview:\n%v\n", utils.MultilinePreview(blockRegistryStr))
	}, func() {
		requestLookupIndexStr = newRequestLookupIndex(contractState, firstIndex, lastIndex)
		cli.DebugLogf("New request lookup index preview:\n%v\n", utils.MultilinePreview(requestLookupIndexStr))
	})

	return receiptsStr + "\n" + blockRegistryStr + "\n" + requestLookupIndexStr
}

func NewReceiptsContentToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
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

	cli.DebugLogf("Retrieving old blocks: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
	blocks := old_collections.NewArrayReadOnly(contractState, old_blocklog.PrefixBlockRegistry)
	var blocksStr strings.Builder
	printProgress, done := NewProgressPrinter("old_blocklog", "blocks", "blocks", lastIndex-firstAvailBlockIndex+1)
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

	return blocksStr.String()
}

func newBlockRegistryToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	firstAvailBlockIndex := getFirstAvailableBlockIndex(firstIndex, lastIndex)

	cli.DebugLogf("Retrieving new blocks: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailBlockIndex, lastIndex)
	blocks := blocklog.NewStateReader(contractState).GetBlockRegistry()
	var blocksStr strings.Builder
	printProgress, done := NewProgressPrinter("new_blocklog", "blocks", "blocks", lastIndex-firstAvailBlockIndex+1)
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

	return blocksStr.String()
}

func oldRequestLookupIndex(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	cli.DebugLogf("Retrieving old request lookup index")
	lookupTable := old_collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestLookupIndex)

	// The table is too big to be stringified - so just checking stringifying requests from last 10000 blocks,
	// plus gathering general statistics

	var indexStr strings.Builder

	done1 := Go(func() {
		firstAvailBlockIndex := max(getFirstAvailableBlockIndex(firstIndex, lastIndex), 1)
		printProgress, done := NewProgressPrinter("old_blocklog", "lookup (requests)", "requests", lastIndex-firstAvailBlockIndex+1)
		defer done()

		for blockIndex := max(firstAvailBlockIndex, 1); blockIndex < lastIndex; blockIndex++ {
			printProgress()
			_, reqs := lo.Must2(old_blocklog.GetRequestsInBlock(contractState, blockIndex))

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
	})

	realCount := 0
	nonEmptryLen := 0

	done2 := Go(func() {
		printProgress, done := NewProgressPrinter("old_blocklog", "lookup (count)", "entries", lookupTable.Len())
		defer done()

		lookupTable.Iterate(func(key, value []byte) bool {
			realCount++
			if len(value) != 0 {
				nonEmptryLen++
			}

			printProgress()
			return true
		})
	})

	<-done1
	<-done2

	cli.DebugLogf("Retrieved %v request lookup index entries", realCount)
	indexStr.WriteString(fmt.Sprintf("Request lookup index virtual len: %v\n", lookupTable.Len()))
	indexStr.WriteString(fmt.Sprintf("Request lookup index real len: %v\n", realCount))
	indexStr.WriteString(fmt.Sprintf("Request lookup index non-empty len: %v\n", nonEmptryLen))

	return indexStr.String()
}

func newRequestLookupIndex(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	cli.DebugLogf("Retrieving new request lookup index")
	lookupTable := collections.NewMapReadOnly(contractState, old_blocklog.PrefixRequestLookupIndex)

	// The table is too big to be stringified - so just checking stringifying requests from last 10000 blocks,
	// plus gathering general statistics

	var indexStr strings.Builder

	done1 := Go(func() {
		firstAvailBlockIndex := max(getFirstAvailableBlockIndex(firstIndex, lastIndex), 1)
		printProgress, done := NewProgressPrinter("new_blocklog", "lookup (requests)", "requests", lastIndex-firstAvailBlockIndex+1)
		defer done()
		r := blocklog.NewStateReader(contractState)

		for blockIndex := firstAvailBlockIndex; blockIndex < lastIndex; blockIndex++ {
			printProgress()
			_, reqs := lo.Must2(r.GetRequestsInBlock(blockIndex))

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
	})

	realCount := 0
	nonEmptryLen := 0

	done2 := Go(func() {
		printProgress, done := NewProgressPrinter("new_blocklog", "lookup (entries)", "entries", lookupTable.Len())
		defer done()

		lookupTable.Iterate(func(key, value []byte) bool {
			realCount++
			if len(value) != 0 {
				nonEmptryLen++
			}

			printProgress()
			return true
		})
	})

	<-done1
	<-done2

	cli.DebugLogf("Retrieved %v request lookup index entries", realCount)
	indexStr.WriteString(fmt.Sprintf("Request lookup index virtual len: %v\n", lookupTable.Len()))
	indexStr.WriteString(fmt.Sprintf("Request lookup index real len: %v\n", realCount))
	indexStr.WriteString(fmt.Sprintf("Request lookup index non-empty len: %v\n", nonEmptryLen))

	return indexStr.String()
}
