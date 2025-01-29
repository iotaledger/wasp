package main

import (
	"bytes"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/nnikolash/wasp-types-exported/packages/kv"
	"github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	"github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"
)

func GetAnchorOutput(chainState kv.KVStoreReader) *iotago.AliasOutput {
	contractState := getContactStateReader(chainState, blocklog.Contract.Hname())

	registry := collections.NewArrayReadOnly(contractState, blocklog.PrefixBlockRegistry)
	if registry.Len() == 0 {
		panic("Block registry is empty")
	}

	blockInfoBytes := registry.GetAt(registry.Len() - 1)

	var blockInfo blocklog.BlockInfo
	lo.Must0(blockInfo.Read(bytes.NewReader(blockInfoBytes)))

	return blockInfo.PreviousAliasOutput.GetAliasOutput()
}
