package fairauction

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/tools/wasp-client/config/fa"
	"github.com/iotaledger/wasp/tools/wasp-client/util"
	"github.com/mr-tron/base58"
)

func startAuctionCmd(args []string) {
	if len(args) != 5 {
		fa.Config.PrintUsage("start-auction <description> <color> <amount> <minumum-bid> <duraion>")
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

	util.WithTransaction(func() (*transaction.Transaction, error) {
		tx, err := fa.Client().StartAuction(
			description,
			color,
			int64(amount),
			int64(minimumBid),
			int64(durationMinutes),
		)
		if err != nil {
			return nil, fmt.Errorf("StartAuction failed: %v", err)
		}
		return tx.Transaction, nil
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

	util.WithTransaction(func() (*transaction.Transaction, error) {
		tx, err := fa.Client().PlaceBid(color, int64(amount))
		if err != nil {
			return nil, err
		}
		return tx.Transaction, nil
	})
}
