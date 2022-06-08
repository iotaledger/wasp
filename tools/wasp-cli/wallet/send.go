package wallet

import (
	"github.com/spf13/cobra"
)

var sendFundsCmd = &cobra.Command{
	Use:   "send-funds <target-address> <token-id> <amount>",
	Short: "Transfer L1 tokens",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		panic("TODO implement")
		// wallet := Load()
		// sourceAddress := wallet.Address()
		// _, targetAddress, err := iotago.ParseBech32(args[0])
		// log.Check(err)

		// tokenID := decodeTokenID(args[1])

		// amount, err := strconv.Atoi(args[2])
		// log.Check(err)

		// outs, err := config.GoshimmerClient().GetConfirmedOutputs(sourceAddress)
		// log.Check(err)

		// tx := util.WithTransaction(func() (*ledgerstate.Transaction, error) {
		// 	txb := utxoutil.NewBuilder(outs...)
		// 	bals := colored.ToL1Map(colored.NewBalancesForColor(tokenID, uint64(amount)))
		// 	err := txb.AddSigLockedColoredOutput(targetAddress, bals)
		// 	log.Check(err)
		// 	err = txb.AddRemainderOutputIfNeeded(sourceAddress, nil, true)
		// 	log.Check(err)
		// 	return txb.BuildWithED25519(wallet.KeyPair())
		// })

		// log.Printf("Transaction %s posted successfully.\n", tx.ID())
	},
}
