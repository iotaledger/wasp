package m001

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

var AccountDecimals = migrations.Migration{
	Contract: accounts.Contract,
	Apply: func(accountsState kv.KVStore, log *logger.Logger) error {
		migrateBaseTokens := func(accKey []byte) {
			// converts an account base token balance from uint64 to big.Int (while changing the decimals from 6 to 18)
			key := accounts.BaseTokensKey(kv.Key(accKey))
			amountBytes := accountsState.Get(key)
			if amountBytes == nil {
				return
			}
			amount := codec.Uint64.MustDecode(amountBytes)
			amountMigrated := util.BaseTokensDecimalsToEthereumDecimals(amount, 6)
			accountsState.Set(key, codec.BigIntAbs.Encode(amountMigrated))
		}

		// iterate though all accounts,
		const keyAllAccounts = "a"
		allAccountsMap := collections.NewMapReadOnly(accountsState, keyAllAccounts)
		allAccountsMap.IterateKeys(func(accountKey []byte) bool {
			// migrate each account
			migrateBaseTokens(accountKey)
			return true
		})
		// migrate the "totals account"
		migrateBaseTokens([]byte(accounts.L2TotalsAccount))
		return nil
	},
}
