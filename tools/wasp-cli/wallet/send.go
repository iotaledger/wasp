package wallet

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

var sendFundsCmd = &cobra.Command{
	Use:   "send-funds <target-address> <token-id>:<amount> <token-id2>:<amount> ...",
	Short: "Transfer L1 tokens",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		_, targetAddress, err := iotago.ParseBech32(args[0])
		log.Check(err)

		tokens := util.ParseFungibleTokens(args[1:])
		log.Check(err)

		log.Printf("Sending %v to %v (%v)\n", tokens, args[0], targetAddress)

		wallet := Load()
		senderAddress := wallet.Address()
		client := config.L1Client()

		outputSet, err := client.OutputMap(senderAddress)
		log.Check(err)

		tx, err := transaction.NewTransferTransaction(transaction.NewTransferTransactionParams{
			DisableAutoAdjustDustDeposit: !adjustDustDeposit,
			FungibleTokens:               tokens,
			SendOptions:                  iscp.SendOptions{},
			SenderAddress:                senderAddress,
			SenderKeyPair:                wallet.KeyPair,
			TargetAddress:                targetAddress,
			UnspentOutputs:               outputSet,
			UnspentOutputIDs:             iscp.OutputSetToOutputIDs(outputSet),
		})

		log.Check(err)

		txID, err := tx.ID()

		log.Check(err)

		err = client.PostTx(tx)

		log.Check(err)

		log.Printf("Transaction [%v] successfully sent", txID)
	},
}
