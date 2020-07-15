package client

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
)

func StatusCmd(args []string) {
	status := FetchStatus(args)

	fmt.Printf("FairRoulette Smart Contract status:\n")
	fmt.Printf("  bets for next play: %d\n", status.CurrentBetsAmount)
	dumpBets(status.CurrentBets)
	fmt.Printf("  locked bets: %d\n", status.LockedBetsAmount)
	dumpBets(status.LockedBets)
	fmt.Printf("  last winning color: %d\n", status.LastWinningColor)
	fmt.Printf("  play period (s): %d\n", status.PlayPeriodSeconds)
	fmt.Printf("  next play in: %s\n", formatNextPlay(status.NextPlayTimestamp))
	fmt.Printf("  color stats:\n")
	for color, wins := range status.WinsPerColor {
		fmt.Printf("    color %d: %d wins\n", color, wins)
	}
	if len(status.PlayerStats) > 0 {
		fmt.Printf("  player stats:\n")
		for player, stats := range status.PlayerStats {
			fmt.Printf("    %s: %s\n", player.String()[:6], stats)
		}
	}
}

func formatNextPlay(ts time.Time) string {
	diff := ts.Sub(time.Now())
	// round to the second
	diff -= diff % time.Second
	if diff < 0 {
		return "unknown"
	}
	return diff.String()
}

func dumpBets(bets []*fairroulette.BetInfo) {
	if len(bets) > 0 {
		fmt.Printf("    (first %d):\n", BetsSliceLength)
	}
	for i, bet := range bets {
		fmt.Printf("      %d: %s\n", i, bet.String())
	}
}

