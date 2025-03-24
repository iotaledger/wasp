package validation

import (
	"fmt"
	"reflect"
	"strings"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func OldBlocklogContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, index uint32) string {
	return OldReceiptsContentToStr(contractState, index)
}

func OldReceiptsContentToStr(contractState old_kv.KVStoreReader, index uint32) string {
	var requestStr strings.Builder

	for blockIndex := 1; blockIndex < int(index); blockIndex++ {
		_, requests, err := old_blocklog.GetRequestsInBlock(contractState, uint32(blockIndex))
		if err != nil {
			panic(err)
		}

		for _, req := range requests {
			// 																						^ There is no concept of RequestIndexes anymore, snip it.
			str := fmt.Sprintf("Type:%s,ID:%s,BaseToken:%d\n", reflect.TypeOf(req), req.ID().OutputID().TransactionID().ToHex(), req.Assets().BaseTokens)
			requestStr.WriteString(str)
		}
	}

	return requestStr.String()
}

func NewBlocklogContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID, index uint32) string {
	return NewReceiptsContentToStr(contractState, index)
}

func NewReceiptsContentToStr(contractState kv.KVStoreReader, index uint32) string {
	var requestStr strings.Builder

	for blockIndex := 1; blockIndex < int(index); blockIndex++ {
		_, requests, err := blocklog.NewStateReader(contractState).GetRequestsInBlock(uint32(blockIndex))
		if err != nil {
			panic(err)
		}

		for _, req := range requests {
			str := fmt.Sprintf("Type:%s,ID:%s,BaseToken:%d\n", reflect.TypeOf(req), req.ID(), req.Assets().BaseTokens()/1000) // Base token conversion 9=>6
			requestStr.WriteString(str)
		}
	}

	return requestStr.String()
}
