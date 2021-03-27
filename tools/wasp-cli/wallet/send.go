package wallet

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"os"
	"strconv"

	"github.com/iotaledger/wasp/packages/txutil/vtxbuilder"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	clientutil "github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func sendFundsCmd(args []string) {
	if len(args) < 3 {
		log.Usage("%s send-funds <target-address> <color> <amount>\n", os.Args[0])
	}

	wallet := Load()
	sourceAddress := wallet.Address()

	targetAddress, err := ledgerstate.AddressFromBase58EncodedString(args[0])
	log.Check(err)

	color := decodeColor(args[1])

	amount, err := strconv.Atoi(args[2])
	log.Check(err)

	bals, err := config.GoshimmerClient().GetConfirmedOutputs(sourceAddress)
	log.Check(err)

	vtxb, err := vtxbuilder.NewFromOutputBalances(bals)
	log.Check(err)

	log.Check(vtxb.MoveTokensToAddress(targetAddress, color, uint64(amount)))

	tx := vtxb.Build(false)
	tx.Sign(wallet.SignatureScheme())

	clientutil.PostTransaction(tx)
}

func decodeColor(s string) ledgerstate.Color {
	color, err := ledgerstate.ColorFromBase58EncodedString(s)
	log.Check(err)
	return color
}
