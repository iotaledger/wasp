package wallet

import (
	"github.com/spf13/cobra"
)

var mintCmd = &cobra.Command{
	Use:   "mint <amount>",
	Short: "Mint some colored tokens",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		panic("TODO implement")
		// amount, err := strconv.Atoi(args[0])
		// log.Check(err)

		// wallet := Load()
		// address := wallet.Address()

		// outs, err := config.GoshimmerClient().GetConfirmedOutputs(address)
		// log.Check(err)

		// txb := utxoutil.NewBuilder(outs...)
		// log.Check(txb.AddSigLockedIOTAOutput(address, uint64(amount), uint64(amount)))
		// log.Check(txb.AddRemainderOutputIfNeeded(address, nil, true))
		// tx, err := txb.BuildWithED25519(wallet.KeyPair())
		// log.Check(err)

		// util.PostTransaction(tx)

		// minted := utxoutil.GetMintedAmounts(tx)
		// if len(minted) == 0 {
		// 	panic("transaction does not contain minted tokens")
		// }
		// for color := range minted {
		// 	log.Printf("Minted %d tokens of color %s\n", amount, color.Base58())
		// }
		// log.Printf("Transaction ID: %s\n", tx.ID().Base58())
	},
}
