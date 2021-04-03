package wallet

import (
	"strconv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

var mintCmd = &cobra.Command{
	Use:   "mint <amount>",
	Short: "Mint some colored tokens",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		amount, err := strconv.Atoi(args[0])
		log.Check(err)

		wallet := Load()
		address := wallet.Address()

		outs, err := config.GoshimmerClient().GetConfirmedOutputs(address)
		log.Check(err)

		tx := util.WithTransaction(func() (*ledgerstate.Transaction, error) {
			txb := utxoutil.NewBuilder(outs...)
			log.Check(txb.AddSigLockedColoredOutput(address, nil, uint64(amount)))
			log.Check(txb.AddReminderOutputIfNeeded(address, nil, true))
			return txb.BuildWithED25519(wallet.KeyPair())
		})

		log.Printf("Minted %d tokens of color %s\n", amount, tx.ID())
	},
}
