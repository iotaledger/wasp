package m001

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

var AccountDecimals = migrations.Migration{
	Contract: accounts.Contract,
	Apply: func(state kv.KVStore, log *logger.Logger) error {
		migrateBaseTokens := func(accKey []byte) {
			// converts an account base token balance from uint64 to big.Int (while changing the decimals from 6 to 18)
			key := accounts.BaseTokensKey(kv.Key(accKey))
			amount := codec.MustDecodeUint64(state.Get(key))
			amountMigrated := util.MustBaseTokensDecimalsToEthereumDecimalsExact(amount, 6)
			state.Set(key, codec.EncodeBigIntAbs(amountMigrated))
		}

		// iterate though all accounts,
		allAccountsMap := accounts.AllAccountsMapR(state)
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
