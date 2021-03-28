package wallet

import (
	"os"
	"strconv"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func mintCmd(args []string) {
	if len(args) < 1 {
		log.Usage("%s mint <amount>\n", os.Args[0])
	}

	wallet := Load()

	amount, err := strconv.Atoi(args[0])
	log.Check(err)

	tx := util.WithTransaction(func() (*transaction.Transaction, error) {
		return vtxbuilder.NewColoredTokensTransaction(config.GoshimmerClient(), wallet.SignatureScheme(), uint64(amount))
	})

	log.Printf("Minted %d tokens of color %s\n", amount, tx.ID())
}
