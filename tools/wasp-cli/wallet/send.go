package wallet

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

func sendFundsCmd() *cobra.Command {
	var adjustDustDeposit bool

	cmd := &cobra.Command{
		Use:   "send-funds <target-address> <token-id>:<amount> <token-id2>:<amount> ...",
		Short: "Transfer L1 tokens",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			_, targetAddress, err := iotago.ParseBech32(args[0])
			log.Check(err)

			tokens := util.ParseFungibleTokens(args[1:])
			log.Check(err)

			log.Printf("\nSending \n\t%v \n\tto: %v\n\n", tokens, args[0])

			wallet := Load()
			senderAddress := wallet.Address()
			client := config.L1Client()

			outputSet, err := client.OutputMap(senderAddress)
			log.Check(err)

			if !adjustDustDeposit {
				// check if the resulting output needs to be adjusted for Storage Deposit
				output := transaction.MakeBasicOutput(
					targetAddress,
					senderAddress,
					tokens,
					nil,
					isc.SendOptions{},
					true,
				)
				util.SDAdjustmentPrompt(output)
			}

			tx, err := transaction.NewTransferTransaction(transaction.NewTransferTransactionParams{
				DisableAutoAdjustDustDeposit: false,
				FungibleTokens:               tokens,
				SendOptions:                  isc.SendOptions{},
				SenderAddress:                senderAddress,
				SenderKeyPair:                wallet.KeyPair,
				TargetAddress:                targetAddress,
				UnspentOutputs:               outputSet,
				UnspentOutputIDs:             isc.OutputSetToOutputIDs(outputSet),
			})
			log.Check(err)

			txID, err := tx.ID()
			log.Check(err)

			err = client.PostTx(tx)
			log.Check(err)

			log.Printf("Transaction [%v] sent successfully.\n", txID.ToHex())
		},
	}

	cmd.Flags().BoolVarP(&adjustDustDeposit, "adjust-storage-deposit", "a", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")

	return cmd
}
