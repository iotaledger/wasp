// Excluded for now as we right now don't support minting new coins
//go:build exclude

package chain

import (
	"encoding/hex"
	"math/big"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func initCreateNativeTokenCmd() *cobra.Command {
	var maxSupply, mintedTokens, meltedTokens int64
	var tokenName, tokenSymbol string
	var tokenDecimals uint8

	return buildPostRequestCmd(
		"create-native-token",
		"Calls accounts core contract nativeTokenCreate to create a new native token",
		accounts.Contract.Name,
		accounts.FuncNativeTokenCreate.Name,
		func(cmd *cobra.Command) {
			cmd.Flags().Int64Var(&maxSupply, "max-supply", 1000000, "Maximum token supply")
			cmd.Flags().Int64Var(&mintedTokens, "minted-tokens", 0, "Minted tokens")
			cmd.Flags().Int64Var(&meltedTokens, "melted-tokens", 0, "Melted tokens")
			cmd.Flags().StringVar(&tokenName, "token-name", "", "Token name")
			cmd.Flags().StringVar(&tokenSymbol, "token-symbol", "", "Token symbol")
			cmd.Flags().Uint8Var(&tokenDecimals, "token-decimals", uint8(8), "Token decimals")
		},
		func(cmd *cobra.Command) []string {
			tokenScheme := &iotago.SimpleTokenScheme{
				MaximumSupply: big.NewInt(maxSupply),
				MintedTokens:  big.NewInt(mintedTokens),
				MeltedTokens:  big.NewInt(meltedTokens),
			}

			tokenSchemeBytes := codec.Encode[TokenScheme](tokenScheme)

			return []string{
				"string", accounts.ParamTokenScheme, "bytes", "0x" + hex.EncodeToString(tokenSchemeBytes),
				"string", accounts.ParamTokenName, "bytes", "0x" + hex.EncodeToString(codec.Encode[string](tokenName)),
				"string", accounts.ParamTokenTickerSymbol, "bytes", "0x" + hex.EncodeToString(codec.Encode[string](tokenSymbol)),
				"string", accounts.ParamTokenDecimals, "bytes", "0x" + hex.EncodeToString(codec.Encode[uint8](tokenDecimals)),
			}
		},
	)
}
