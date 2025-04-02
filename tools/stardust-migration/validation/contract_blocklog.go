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
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

const blockRetentionPeriod = 10000

func OldBlocklogContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, firstIndex, lastIndex uint32) string {
	receiptsStr := OldReceiptsContentToStr(contractState, firstIndex, lastIndex)
	cli.DebugLogf("Old receipts preview:\n%v\n", utils.MultilinePreview(receiptsStr))

	blockRegistryStr := oldBlockRegistryToStr(contractState, firstIndex, lastIndex)
	cli.DebugLogf("Old block registry preview:\n%v\n", utils.MultilinePreview(blockRegistryStr))

	return receiptsStr + "\n" + blockRegistryStr
}

func OldReceiptsContentToStr(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)

	cli.DebugLogf("Retrieving old receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)
	var requestStr strings.Builder
	reqCount := 0
	printProgress, done := cli.NewProgressPrinter("receipts", lastIndex-firstAvailableBlockIndex)
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
		if !strings.Contains(err.Error(), "request not found") {
			panic(err)
		}

		requestStr.WriteString(fmt.Sprintf("Last pruned block (%v): unavailable\n", firstUnavailableBlockIndex))
	}

	return requestStr.String()
}

func NewBlocklogContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID, firstIndex, lastIndex uint32) string {
	receiptsStr := NewReceiptsContentToStr(contractState, firstIndex, lastIndex)
	cli.DebugLogf("New receipts preview:\n%v\n", utils.MultilinePreview(receiptsStr))

	blockRegistryStr := newBlockRegistryToStr(contractState, firstIndex, lastIndex)
	cli.DebugLogf("New block registry preview:\n%v\n", utils.MultilinePreview(blockRegistryStr))

	return receiptsStr + "\n" + blockRegistryStr
}

func NewReceiptsContentToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)

	cli.DebugLogf("Retrieving new receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)
	var requestStr strings.Builder
	reqCount := 0
	printProgress, done := cli.NewProgressPrinter("receipts", lastIndex-firstAvailableBlockIndex)
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
	printProgress, done := cli.NewProgressPrinter("blocks", lastIndex-firstAvailBlockIndex)
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
	printProgress, done := cli.NewProgressPrinter("blocks", lastIndex-firstAvailBlockIndex)
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
