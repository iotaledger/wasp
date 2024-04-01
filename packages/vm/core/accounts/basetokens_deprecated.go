package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

// deprecated on v1.0.1-rc.16

func getBaseTokensDEPRECATED(state kv.KVStoreReader, accountKey kv.Key) uint64 {
	return codec.MustDecodeUint64(state.Get(BaseTokensKey(accountKey)), 0)
}

func getBaseTokensFullDecimalsDEPRECATED(state kv.KVStoreReader, accountKey kv.Key) *big.Int {
	amount := codec.MustDecodeUint64(state.Get(BaseTokensKey(accountKey)), 0)
	return util.BaseTokensDecimalsToEthereumDecimals(amount, parameters.L1().BaseToken.Decimals)
}

func setBaseTokensFullDecimalsDEPRECATED(state kv.KVStore, accountKey kv.Key, amount *big.Int) {
	baseTokens, _ := util.EthereumDecimalsToBaseTokenDecimals(amount, parameters.L1().BaseToken.Decimals)
	state.Set(BaseTokensKey(accountKey), codec.EncodeUint64(baseTokens))
}
