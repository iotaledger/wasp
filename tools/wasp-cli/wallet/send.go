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

var sendFundsCmd = &cobra.Command{
	Use:   "send-funds <target-address> <color> <amount>",
	Short: "Transfer tokens",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		wallet := Load()
		sourceAddress := wallet.Address()

		targetAddress, err := ledgerstate.AddressFromBase58EncodedString(args[0])
		log.Check(err)

		color := decodeColor(args[1])

		amount, err := strconv.Atoi(args[2])
		log.Check(err)

		outs, err := config.GoshimmerClient().GetConfirmedOutputs(sourceAddress)
		log.Check(err)

		tx := util.WithTransaction(func() (*ledgerstate.Transaction, error) {
			txb := utxoutil.NewBuilder(outs...)
			err := txb.AddSigLockedColoredOutput(targetAddress, map[ledgerstate.Color]uint64{color: uint64(amount)})
			log.Check(err)
			err = txb.AddRemainderOutputIfNeeded(sourceAddress, nil, true)
			log.Check(err)
			return txb.BuildWithED25519(wallet.KeyPair())
		})

		log.Printf("Transaction %s posted successfully.\n", tx.ID())
	},
}

func decodeColor(s string) ledgerstate.Color {
	color, err := ledgerstate.ColorFromBase58EncodedString(s)
	log.Check(err)
	return color
}
