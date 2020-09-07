package facmd

import (
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
	"github.com/iotaledger/wasp/tools/wwallet/util"
	"github.com/mr-tron/base58"
)

func startAuctionCmd(args []string) {
	if len(args) != 5 {
		fa.Config.PrintUsage("start-auction <description> <color> <amount> <minumum-bid> <duration>")
		os.Exit(1)
	}

	description := args[0]

	color := decodeColor(args[1])

	amount, err := strconv.Atoi(args[2])
	check(err)

	minimumBid, err := strconv.Atoi(args[3])
	check(err)

	durationMinutes, err := strconv.Atoi(args[4])
	check(err)

	util.WithSCRequest(fa.Config, func() (*sctransaction.Transaction, error) {
		return fa.Client().StartAuction(
			description,
			color,
			int64(amount),
			int64(minimumBid),
			int64(durationMinutes),
		)
	})
}

func decodeColor(s string) *balance.Color {
	b, err := base58.Decode(s)
	check(err)
	color, _, err := balance.ColorFromBytes(b)
	check(err)
	return &color
}

func placeBidCmd(args []string) {
	if len(args) != 2 {
		fa.Config.PrintUsage("place-bid <color> <amount>")
		os.Exit(1)
	}

	color := decodeColor(args[0])

	amount, err := strconv.Atoi(args[1])
	check(err)

	util.WithSCRequest(fa.Config, func() (*sctransaction.Transaction, error) {
		return fa.Client().PlaceBid(color, int64(amount))
	})
}
