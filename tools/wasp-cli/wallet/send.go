package wallet

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func initSendFundsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-funds <target-address> <token-id1>|<amount1> <token-id2>|<amount2> ...",
		Short: "Transfer L1 tokens",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			targetAddress, err := cryptolib.NewAddressFromHexString(args[0])
			log.Check(err)

			tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))
			log.Check(err)

			log.Printf("\nSending \n\t%v \n\tto: %v\n\n", tokens, args[0])

			myWallet := wallet.Load()
			senderAddress := myWallet.Address()
			util.TryMergeAllCoins(cmd.Context())

			client := cliclients.L1Client()

			balances, err := client.GetAllBalances(context.Background(), senderAddress.AsIotaAddress())
			log.Check(err)
			for _, balance := range balances {
				requestedAmt := tokens.Coins[coin.MustTypeFromString(balance.CoinType.String())]
				if coin.Value(balance.TotalBalance.Int64()) < requestedAmt {
					panic("not enough balance")
				}
			}

			ptb := iotago.NewProgrammableTransactionBuilder()

			coinPage, err := client.GetAllCoins(
				context.Background(), iotaclient.GetAllCoinsRequest{
					Owner: senderAddress.AsIotaAddress(),
				},
			)
			log.Check(err)
			for cointype, balance := range tokens.Coins {
				if cointype.MatchesStringType(coin.BaseTokenType.String()) {
					argSplitCoins := ptb.Command(iotago.Command{SplitCoins: &iotago.ProgrammableSplitCoins{
						Coin:    iotago.GetArgumentGasCoin(),
						Amounts: []iotago.Argument{ptb.MustForceSeparatePure(balance.Uint64())},
					}})
					ptb.Command(iotago.Command{TransferObjects: &iotago.ProgrammableTransferObjects{
						Objects: []iotago.Argument{
							iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *argSplitCoins.Result, Result: uint16(0)}},
						},
						Address: ptb.MustPure(targetAddress.AsIotaAddress()),
					}})
				} else {
					pickedCoin, err := iotajsonrpc.PickupCoinsWithCointype(
						coinPage,
						balance.BigInt(),
						iotajsonrpc.MustCoinTypeFromString(cointype.String()),
					)
					log.Check(err)

					err = ptb.Pay(pickedCoin.CoinRefs(), []*iotago.Address{targetAddress.AsIotaAddress()}, []uint64{balance.Uint64()})
					log.Check(err)
				}
			}

			pt := ptb.Finish()

			gasPayments, err := client.FindCoinsForGasPayment(context.TODO(), senderAddress.AsIotaAddress(), pt, iotaclient.DefaultGasPrice, iotaclient.DefaultGasBudget)
			if err != nil {
				panic(fmt.Sprintf("failed to find gas payment: %s", err))
			}
			if len(gasPayments) == 0 {
				panic("no coin found as gas payment")
			}
			tx := iotago.NewProgrammable(
				senderAddress.AsIotaAddress(),
				pt,
				gasPayments,
				iotaclient.DefaultGasBudget,
				iotaclient.DefaultGasPrice,
			)
			txBytes, err := bcs.Marshal(&tx)
			log.Check(err)

			res, err := client.SignAndExecuteTransaction(
				context.Background(),
				&iotaclient.SignAndExecuteTransactionRequest{
					Signer:      cryptolib.SignerToIotaSigner(myWallet),
					TxDataBytes: txBytes,
					Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
						ShowEffects:       true,
						ShowObjectChanges: true,
					},
				},
			)

			log.Check(err)
			fmt.Printf("%v", res)
		},
	}

	return cmd
}
