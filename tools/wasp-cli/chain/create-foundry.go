package chain

import (
	"encoding/hex"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/spf13/cobra"
)

func initCreateFoundryCmd() *cobra.Command {
	var maxSupply, mintedTokens, meltedTokens int64

	return buildPostRequestCmd(
		"create-foundry",
		"Call accounts core contract foundryCreateNew to create a new foundry",
		accounts.Contract.Name,
		accounts.FuncFoundryCreateNew.Name,
		func(cmd *cobra.Command) {
			cmd.Flags().Int64Var(&maxSupply, "max-supply", 1000000, "Maximum token supply")
			cmd.Flags().Int64Var(&mintedTokens, "minted-tokens", 0, "Minted tokens")
			cmd.Flags().Int64Var(&meltedTokens, "melted-tokens", 0, "Melted tokens")
		},
		func(cmd *cobra.Command) []string {
			tokenScheme := &iotago.SimpleTokenScheme{
				MaximumSupply: big.NewInt(maxSupply),
				MintedTokens:  big.NewInt(mintedTokens),
				MeltedTokens:  big.NewInt(meltedTokens),
			}

			tokenSchemeBytes := codec.EncodeTokenScheme(tokenScheme)

			return []string{"string", "t", "bytes", "0x" + hex.EncodeToString(tokenSchemeBytes)}
		},
	)
}
