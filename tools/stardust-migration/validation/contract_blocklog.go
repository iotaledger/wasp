package validation

import (
	"fmt"
	"reflect"
	"strings"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

const blockRetentionPeriod = 10000

func OldBlocklogContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, firstIndex, lastIndex uint32) string {
	return OldReceiptsContentToStr(contractState, firstIndex, lastIndex)
}

func OldReceiptsContentToStr(contractState old_kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var requestStr strings.Builder
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)
	cli.Logf("Retrieving old receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)

	for blockIndex := firstAvailableBlockIndex; blockIndex < lastIndex; blockIndex++ {
		_, requests, err := old_blocklog.GetRequestsInBlock(contractState, blockIndex)
		if err != nil {
			panic(err)
		}

		for _, req := range requests {
			// 																						^ There is no concept of RequestIndexes anymore, snip it.
			str := fmt.Sprintf("Type:%s,ID:%s,BaseToken:%d\n", reflect.TypeOf(req), req.ID().OutputID().TransactionID().ToHex(), req.Assets().BaseTokens)
			requestStr.WriteString(str)
		}
	}

	if firstAvailableBlockIndex > 0 {
		firstUnavailableBlockIndex := firstAvailableBlockIndex - 1
		_, _, err := old_blocklog.GetRequestsInBlock(contractState, firstUnavailableBlockIndex)
		requestStr.WriteString(fmt.Sprintf("Last pruned block (%v): %v\n",
			firstUnavailableBlockIndex, lo.Ternary(err == nil, "AVAILABLE", "unavailable")))
	}

	return requestStr.String()
}

func NewBlocklogContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID, firstIndex, lastIndex uint32) string {
	return NewReceiptsContentToStr(contractState, firstIndex, lastIndex)
}

func NewReceiptsContentToStr(contractState kv.KVStoreReader, firstIndex, lastIndex uint32) string {
	var requestStr strings.Builder
	var firstAvailableBlockIndex = getFirstAvailableBlockIndex(firstIndex, lastIndex)
	cli.Logf("Retrieving new receipts content: %v-%v => %v-%v", firstIndex, lastIndex, firstAvailableBlockIndex, lastIndex)

	for blockIndex := firstAvailableBlockIndex; blockIndex < lastIndex; blockIndex++ {
		_, requests, err := blocklog.NewStateReader(contractState).GetRequestsInBlock(blockIndex)
		if err != nil {
			panic(err)
		}

		for _, req := range requests {
			str := fmt.Sprintf("Type:%s,ID:%s,BaseToken:%d\n", reflect.TypeOf(req), req.ID(), req.Assets().BaseTokens()/1000) // Base token conversion 9=>6
			requestStr.WriteString(str)
		}
	}

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
