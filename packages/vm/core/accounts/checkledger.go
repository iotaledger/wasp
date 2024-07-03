package accounts

import (
	"fmt"
	"math/big"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
)

func (s *StateReader) CheckLedgerConsistency() {
	t := s.GetTotalL2FungibleTokens()
	c := s.calcL2TotalFungibleTokens()
	if !t.Equals(c) {
		panic(fmt.Sprintf(
			"inconsistent on-chain account ledger\n total assets: %s\ncalc total: %s\n",
			t, c,
		))
	}
}

func (s *StateReader) calcL2TotalFungibleTokens() *isc.Assets {
	ret := isc.NewEmptyAssets()
	totalBaseTokens := big.NewInt(0)

	s.allAccountsMapR().IterateKeys(func(accountKey []byte) bool {
		// add all native tokens owned by each account
		s.nativeTokensMapR(kv.Key(accountKey)).Iterate(func(idBytes []byte, val []byte) bool {
			ret.AddNativeTokens(
				lo.Must(isc.NativeTokenIDFromBytes(idBytes)),
				lo.Must(codec.BigIntAbs.Decode(val)),
			)
			return true
		})
		// use the full decimals for each account, so no dust balance is lost in the calculation
		baseTokensFullDecimals := s.getBaseTokensFullDecimals(kv.Key(accountKey))
		totalBaseTokens = new(big.Int).Add(totalBaseTokens, baseTokensFullDecimals)
		return true
	})

	// convert from 18 decimals, remainder must be 0
	ret.BaseTokens = util.MustEthereumDecimalsToBaseTokenDecimalsExact(totalBaseTokens, parameters.Decimals)
	return ret
}
