package fairauction

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/tools/wwallet/sc/fa"
)

func statusCmd(args []string) {
	status, err := fa.Client().FetchStatus()
	check(err)

	fmt.Printf("FairAuction Smart Contract status:\n")
	dumpBalance(status.SCBalance)
	fmt.Printf("  Owner margin: %d promilles\n", status.OwnerMarginPromille)
	dumpAuctions(status.Auctions)
}

func dumpBalance(bal map[balance.Color]int64) {
	fmt.Printf("  SC balance:\n")
	for color, amount := range bal {
		fmt.Printf("    %s: %d\n", color, amount)
	}
}

func dumpAuctions(auctions map[balance.Color]*fairauction.AuctionInfo) {
	fmt.Printf("  Auctions:\n")
	for color, auction := range auctions {
		fmt.Printf("  - color: %s\n", color)
		fmt.Printf("    owner: %s\n", auction.AuctionOwner)
		fmt.Printf("    description: %s\n", auction.Description)
		fmt.Printf("    started at: %s\n", time.Unix(0, auction.WhenStarted).UTC())
		fmt.Printf("    duration: %d minutes\n", auction.DurationMinutes)
		fmt.Printf("    deposit: %d\n", auction.TotalDeposit)
		fmt.Printf("    tokens for sale: %d\n", auction.NumTokens)
		fmt.Printf("    minimum bid: %d\n", auction.MinimumBid)
		fmt.Printf("    owner margin: %d promilles\n", auction.OwnerMargin)
		fmt.Printf("    bids:\n")
		for _, bid := range auction.Bids {
			fmt.Printf("    - bidder: %s\n", bid.Bidder)
			fmt.Printf("      amount: %d IOTAs\n", bid.Total)
			fmt.Printf("      when: %s\n", time.Unix(0, bid.When).UTC())
		}
	}
}
