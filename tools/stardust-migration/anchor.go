package main

import (
	"bytes"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func GetAnchorOutput(chainState kv.KVStoreReader) *iotago.AliasOutput {
	contractState := getContactStateReader(chainState, blocklog.Contract.Hname())

	registry := collections.NewArrayReadOnly(contractState, PrefixBlockRegistry)
	if registry.Len() == 0 {
		panic("Block registry is empty")
	}

	blockInfoBytes := registry.GetAt(registry.Len() - 1)

	var blockInfo blocklog.BlockInfo
	must(blockInfo.Read(bytes.NewReader(blockInfoBytes)))

	return blockInfo.PreviousAliasOutput.GetAliasOutput()
}
