package solo

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/stretchr/testify/require"
)

func (ch *Chain) RawKVStore() kv.KVStore {
	return ch.VirtualStateAccess().KVStore()
}

func (ch *Chain) GetRaw(k []byte) []byte {
	ret, err := ch.RawKVStore().Get(kv.Key(k))
	require.NoError(ch.Env.T, err)
	return ret
}

func (ch *Chain) HasRaw(k []byte) bool {
	ret, err := ch.RawKVStore().Has(kv.Key(k))
	require.NoError(ch.Env.T, err)
	return ret
}
