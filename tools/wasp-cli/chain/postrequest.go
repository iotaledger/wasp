package chain

import (
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

func postRequestCmd() *cobra.Command {
	var transfer []string
	var offLedger bool

	cmd := &cobra.Command{
		Use:   "post-request <name> <funcname> [params]",
		Short: "Post a request to a contract",
		Long:  "Post a request to contract <name>, function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fname := args[1]
			params := chainclient.PostRequestParams{
				Args:     util.EncodeParams(args[2:]),
				Transfer: parseAssets(transfer),
			}

			scClient := SCClient(iscp.Hn(args[0]))

			if offLedger {
				params.Nonce = uint64(time.Now().UnixNano())
				util.WithOffLedgerRequest(GetCurrentChainID(), func() (*iscp.OffLedgerRequestData, error) {
					return scClient.PostOffLedgerRequest(fname, params)
				})
			} else {
				util.WithSCTransaction(GetCurrentChainID(), func() (*iotago.Transaction, error) {
					return scClient.PostRequest(fname, params)
				})
			}
		},
	}

	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", []string{"IOTA:1"},
		"include a funds transfer as part of the transaction. Format: <color>:<amount>,<color>:amount...",
	)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}

func assetIDFromString(s string) []byte {
	ret, err := hex.DecodeString(s)
	log.Check(err)
	return ret
}

func parseAssets(args []string) *iscp.FungibleTokens {
	assets := iscp.NewEmptyAssets()
	for _, tr := range args {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatalf("colored balances syntax: <color>:<amount>,<color:amount>... -- Example: IOTA:100")
		}
		// In the past we would indicate iotas as 'IOTA:nnn'
		// Now we can simply use ':nnn', but let's keep it
		// backward compatible for now and allow both
		if parts[0] == "IOTA" {
			parts[0] = ""
		}
		assetIDBytes := assetIDFromString(parts[0])

		amount, ok := new(big.Int).SetString(parts[1], 10)
		if !ok {
			log.Fatalf("error parsing token amount")
		}

		if iscp.IsIota(assetIDBytes) {
			assets.AddIotas(amount.Uint64())
			continue
		}

		assetID, err := iscp.NativeTokenIDFromBytes(assetIDBytes)
		log.Check(err)

		assets.AddNativeTokens(assetID, amount)
	}
	return assets
}
