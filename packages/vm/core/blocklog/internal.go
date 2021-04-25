package blocklog

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func SaveNextBlockInfo(state kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(state, BlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	return registry.MustLen() - 1
}
