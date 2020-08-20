package wallet

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder/vtxbuilder"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/util"
	"github.com/mr-tron/base58"
)

func sendFundsCmd(args []string) {
	if len(args) < 3 {
		fmt.Printf("Usage: %s wallet send-funds <target-address> <color> <amount>\n", os.Args[0])
		os.Exit(1)
	}

	wallet := Load()
	sourceAddress := wallet.Address()

	targetAddress, err := address.FromBase58(args[1])
	check(err)

	color := decodeColor(args[2])

	amount, err := strconv.Atoi(args[3])
	check(err)

	bals, err := config.GoshimmerClient().GetAccountOutputs(&sourceAddress)
	check(err)

	vtxb, err := vtxbuilder.NewFromOutputBalances(bals)
	check(err)

	check(vtxb.MoveToAddress(targetAddress, *color, int64(amount)))

	tx := vtxb.Build(false)
	tx.Sign(wallet.SignatureScheme())

	util.PostTransaction(tx)
}

func decodeColor(s string) *balance.Color {
	b, err := base58.Decode(s)
	check(err)
	color, _, err := balance.ColorFromBytes(b)
	check(err)
	return &color
}
