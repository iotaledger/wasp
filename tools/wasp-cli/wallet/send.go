package wallet

import (
	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	cliwallet "github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
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
			_, targetAddress, err := iotago.ParseBech32(args[0])
			log.Check(err)

			tokens := util.ParseAssetArgs(args[1:])
			log.Check(err)

			log.Printf("\nSending \n\t%v \n\tto: %v\n\n", tokens, args[0])

			wallet := cliwallet.Load()
			senderAddress := wallet.Address()
			client := cliclients.L1Client()

			outputSet, err := client.OutputMap(senderAddress)
			log.Check(err)

			if !adjustStorageDeposit {
				// check if the resulting output needs to be adjusted for Storage Deposit
				output := transaction.MakeBasicOutput(
					targetAddress,
					senderAddress,
					tokens,
					nil,
					isc.SendOptions{},
				)
				util.SDAdjustmentPrompt(output)
			}

			tx, err := transaction.NewTransferTransaction(transaction.NewTransferTransactionParams{
				DisableAutoAdjustStorageDeposit: false,
				FungibleTokens:                  tokens,
				SendOptions:                     isc.SendOptions{},
				SenderAddress:                   senderAddress,
				SenderKeyPair:                   wallet.KeyPair,
				TargetAddress:                   targetAddress,
				UnspentOutputs:                  outputSet,
				UnspentOutputIDs:                isc.OutputSetToOutputIDs(outputSet),
			})
			log.Check(err)

			txID, err := tx.ID()
			log.Check(err)

			_, err = client.PostTxAndWaitUntilConfirmation(tx)
			log.Check(err)

			log.Printf("Transaction [%v] sent successfully.\n", txID.ToHex())
		},
	}

	cmd.Flags().BoolVarP(&adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")

	return cmd
}
