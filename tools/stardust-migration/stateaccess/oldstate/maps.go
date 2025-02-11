package oldstate

import (
	"github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/nnikolash/wasp-types-exported/packages/kv"
	"github.com/nnikolash/wasp-types-exported/packages/kv/codec"
)

func IterateAccountMaps(chainState kv.KVStoreReader, prefix kv.Key, sizeOfMapKeys int, f func(*isc.AgentID, kv.Key, []byte)) uint32 {
	count := uint32(0)
	chainState.Iterate(prefix, func(k kv.Key, v []byte) bool {
		agentIDBytes := k[:len(k)-sizeOfMapKeys-1]
		agentID := codec.MustDecodeAgentID([]byte(agentIDBytes))
		mapKey := k[len(k)-sizeOfMapKeys:]
		f(&agentID, mapKey, v)
		count++
		return true
	})
	return count
}
