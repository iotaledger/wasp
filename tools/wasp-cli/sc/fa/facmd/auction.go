// +build ignore

package facmd

import (
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc/fa"
	"github.com/mr-tron/base58"
)

func startAuctionCmd(args []string) {
	if len(args) != 5 {
		fa.Config.PrintUsage("start-auction <description> <color> <amount> <minumum-bid> <duration in minutes>")
		os.Exit(1)
	}

	description := args[0]

	color := decodeColor(args[1])

	amount, err := strconv.Atoi(args[2])
	log.Check(err)

	minimumBid, err := strconv.Atoi(args[3])
	log.Check(err)

	durationMinutes, err := strconv.Atoi(args[4])
	log.Check(err)

	_, err = fa.Client().StartAuction(
		description,
		color,
		int64(amount),
		int64(minimumBid),
		int64(durationMinutes),
	)
	log.Check(err)
}

func decodeColor(s string) *balance.Color {
	b, err := base58.Decode(s)
	log.Check(err)
	color, _, err := balance.ColorFromBytes(b)
	log.Check(err)
	return &color
}

func placeBidCmd(args []string) {
	if len(args) != 2 {
		fa.Config.PrintUsage("place-bid <color> <amount>")
		os.Exit(1)
	}

	color := decodeColor(args[0])

	amount, err := strconv.Atoi(args[1])
	log.Check(err)

	_, err = fa.Client().PlaceBid(color, int64(amount))
	log.Check(err)
}
