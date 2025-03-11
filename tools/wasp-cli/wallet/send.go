package wallet

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func initSendFundsCmd() *cobra.Command {
	var adjustStorageDeposit bool

	cmd := &cobra.Command{
		Use:   "send-funds <target-address> <token-id>:<amount> <token-id2>:<amount> ...",
		Short: "Transfer L1 tokens",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			panic("FIXME")
			targetAddress, err := cryptolib.NewAddressFromHexString(args[0])
			log.Check(err)

			tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))
			log.Check(err)

			log.Printf("\nSending \n\t%v \n\tto: %v\n\n", tokens, args[0])

			myWallet := wallet.Load()
			senderAddress := myWallet.Address()

			client := cliclients.L1Client()

			tx, err := client.TransferIota(context.Background(), iotaclient.TransferIotaRequest{
				Signer:    senderAddress.AsIotaAddress(),
				Amount:    iotajsonrpc.NewBigInt(1337),
				Recipient: targetAddress.AsIotaAddress(),
			})
			log.Check(err)

			res, err := client.SignAndExecuteTransaction(
				context.Background(),
				&iotaclient.SignAndExecuteTransactionRequest{
					Signer:      cryptolib.SignerToIotaSigner(myWallet),
					TxDataBytes: tx.TxBytes,
					Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
						ShowEffects:       true,
						ShowObjectChanges: true,
					},
				},
			)

			log.Check(err)
			fmt.Printf("%v", res)

			/*
				tx, err := transaction.NewTransferTransaction(transaction.NewTransferTransactionParams{
					DisableAutoAdjustStorageDeposit: false,
					FungibleTokens:                  tokens,
					SendOptions:                     isc.SendOptions{},
					SenderAddress:                   senderAddress,
					SenderKeyPair:                   myWallet,
					TargetAddress:                   targetAddress,
					UnspentOutputs:                  outputSet,
					UnspentOutputIDs:                isc.OutputSetToOutputIDs(outputSet),
				})
				log.Check(err)

				txID, err := tx.ID()
				log.Check(err)

				log.Printf("Transaction [%v] sent successfully.\n", txID.ToHex())*/
		},
	}

	cmd.Flags().BoolVarP(&adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")

	return cmd
}
