package wallet

import (
	"context"
	"fmt"
	"time"

	"fortio.org/safecast"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
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
		Short: "Transfer L1 tokens on L1",
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

			client := cliclients.L1Client()

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

			coinPage, err := client.GetCoins(
				context.Background(), iotagraphql.GetCoinsRequest{
					Owner: senderAddress.AsIotaAddress(),
				},
			)
			if err != nil {
				return err
			}
			allCoins := iotagraphql.Coins(coinPage.Address.Coins.Nodes)
			for cointype, balance := range tokens.Coins.Iterate() {
				pickedCoin, pickErr := iotagraphql.PickupCoinsWithCointype(
					allCoins,
					balance.BigInt(),
					iotagraphql.MustCoinTypeFromString(cointype.String()),
				)
				if pickErr != nil {
					return pickErr
				}

				coinRefs, refErr := pickedCoin.CoinRefs()
				if refErr != nil {
					return refErr
				}

				err = ptb.Pay(coinRefs, []*iotago.Address{lo.ToPtr(targetAddress.AsIotaAddress())}, []uint64{balance.Uint64()})
				if err != nil {
					return err
				}
			}

			pt := ptb.Finish()

			gasCoins, err := client.GetCoinObjsForTargetAmount(context.Background(), senderAddress.AsIotaAddress(), iotagraphql.DefaultGasPrice, iotagraphql.DefaultGasBudget)
			if err != nil {
				return fmt.Errorf("failed to find gas payment: %w", err)
			}
			gasCoins, err = iotagraphql.PickupCoinsWithFilter(
				gasCoins,
				iotagraphql.DefaultGasBudget,
				func(c iotagraphql.Coin) bool {
					addr := c.ObjectID()
					return !pt.IsInInputObjects(&addr)
				},
			)
			if err != nil {
				return fmt.Errorf("failed to find gas payment: %w", err)
			}
			if len(gasCoins) == 0 {
				return fmt.Errorf("no coin found as gas payment")
			}
			gasCoinRefs, err := gasCoins.CoinRefs()
			if err != nil {
				return err
			}
			senderIotaAddr := senderAddress.AsIotaAddress()
			tx := iotago.NewProgrammable(
				&senderIotaAddr,
				pt,
				gasCoinRefs,
				iotagraphql.DefaultGasBudget,
				iotagraphql.DefaultGasPrice,
			)
			txBytes, err := bcs.Marshal(&tx)
			if err != nil {
				return err
			}

			res, err := client.SignAndExecuteTransaction(
				context.Background(),
				txBytes,
				cryptolib.SignerToIotaSigner(myWallet),
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
