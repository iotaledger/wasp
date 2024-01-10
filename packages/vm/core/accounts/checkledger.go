package accounts

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
)

// only used in internal tests and solo
func CheckLedger(state kv.KVStoreReader, checkpoint string) {
	t := GetTotalL2FungibleTokens(state)
	c := calcL2TotalFungibleTokens(state)
	if !t.Equals(c) {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n total assets: %s\ncalc total: %s\n",
			checkpoint, t, c))
	}

	totalAccNFTs := GetTotalL2NFTs(state)
	if len(lo.FindDuplicates(totalAccNFTs)) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n duplicate NFTs\n", checkpoint))
	}
	calculatedNFTs := calcL2TotalNFTs(state)
	if len(lo.FindDuplicates(calculatedNFTs)) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n duplicate NFTs\n", checkpoint))
	}
	left, right := lo.Difference(calculatedNFTs, totalAccNFTs)
	if len(left)+len(right) != 0 {
		panic(fmt.Sprintf("inconsistent on-chain account ledger @ checkpoint '%s'\n NFTs don't match\n", checkpoint))
	}
}
