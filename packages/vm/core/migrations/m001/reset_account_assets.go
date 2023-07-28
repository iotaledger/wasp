package m001

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

// for testnet -- delete when deploying ShimmerEVM
var ResetAccountAssets = migrations.Migration{
	Contract: accounts.Contract,
	Apply: func(accountsState kv.KVStore, log *logger.Logger) error {
		accounts.NativeTokenOutputMap(accountsState).Erase()
		erasePrefix(accountsState, accounts.PrefixNativeTokens)

		accounts.AllFoundriesMap(accountsState).Erase()
		erasePrefix(accountsState, accounts.PrefixFoundries)

		erasePrefix(accountsState, accounts.PrefixNFTs)
		erasePrefix(accountsState, accounts.PrefixNFTsByCollection)
		accounts.NFTOutputMap(accountsState).Erase()
		accounts.NFTDataMap(accountsState).Erase()
		return nil
	},
}

func erasePrefix(state kv.KVStore, prefix kv.Key) {
	var keys []kv.Key
	state.IterateKeys(prefix, func(k kv.Key) bool {
		keys = append(keys, k)
		return true
	})
	for _, k := range keys {
		state.Del(k)
	}
}
