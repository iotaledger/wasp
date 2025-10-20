package wallet

import (
	"context"
	"fmt"
	"time"

	"fortio.org/safecast"

	"github.com/spf13/cobra"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"
)

func initSendFundsCmd() *cobra.Command { //nolint:funlen
	cmd := &cobra.Command{
		Use:   "send-funds <target-address> <token-id1>|<amount1> <token-id2>|<amount2> ...",
		Short: "Transfer L1 tokens",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetAddress, err := cryptolib.NewAddressFromHexString(args[0])
			if err != nil {
				return err
			}

			tokens, err := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))
			if err != nil {
				return err
			}

			log.Printf("\nSending \n\t%v \n\tto: %v\n\n", tokens, args[0])

			myWallet := wallet.Load()
			senderAddress := myWallet.Address()
			util.TryManageCoinsAmount(cmd.Context())
			time.Sleep(3 * time.Second)

			client := cliclients.L1ClientBinding()

			balances, err := client.GetAllBalances(context.Background(), senderAddress.AsIotaAddress())
			if err != nil {
				return err
			}
			for _, balance := range balances {
				requestedAmt := tokens.Coins.Get(coin.MustTypeFromString(balance.CoinType.String()))
				var coinValue uint64
				coinValue, err = safecast.Convert[uint64](balance.TotalBalance.Int64())
				if err != nil {
					return err
				}
				if coin.Value(coinValue) < requestedAmt {
					return fmt.Errorf("not enough balance")
				}
			}

			ptb := iotago.NewProgrammableTransactionBuilder()

			coinPage, err := client.GetAllCoins(
				context.Background(), iotaclient.GetAllCoinsRequest{
					Owner: senderAddress.AsIotaAddress(),
				},
			)
			if err != nil {
				return err
			}
			for cointype, balance := range tokens.Coins.Iterate() {
				var pickedCoin *iotajsonrpc.PickedCoins
				pickedCoin, err = iotajsonrpc.PickupCoinsWithCointype(
					coinPage,
					balance.BigInt(),
					iotajsonrpc.MustCoinTypeFromString(cointype.String()),
				)
				if err != nil {
					return err
				}

				err = ptb.Pay(pickedCoin.CoinRefs(), []*iotago.Address{targetAddress.AsIotaAddress()}, []uint64{balance.Uint64()})
				if err != nil {
					return err
				}
			}

			pt := ptb.Finish()

			gasPayments, err := client.FindCoinsForGasPayment(context.TODO(), senderAddress.AsIotaAddress(), pt, iotaclient.DefaultGasPrice, iotaclient.DefaultGasBudget)
			if err != nil {
				return fmt.Errorf("failed to find gas payment: %s", err)
			}
			if len(gasPayments) == 0 {
				return fmt.Errorf("no coin found as gas payment")
			}
			tx := iotago.NewProgrammable(
				senderAddress.AsIotaAddress(),
				pt,
				gasPayments,
				iotaclient.DefaultGasBudget,
				iotaclient.DefaultGasPrice,
			)
			txBytes, err := bcs.Marshal(&tx)
			if err != nil {
				return err
			}

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
			if err != nil {
				return err
			}
			fmt.Printf("%v", res)
			return nil
		},
	}

	return cmd
}
