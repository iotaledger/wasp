package client

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
)

func StatusCmd(args []string) {
	status, err := FetchStatus()
	check(err)

	fmt.Printf("FairRoulette Smart Contract status:\n")
	fmt.Printf("  play period (s): %d\n", status.PlayPeriodSeconds)
	fmt.Printf("  next play in: %s\n", status.NextPlayIn())
	fmt.Printf("  bets for next play: %d\n", status.CurrentBetsAmount)
	dumpBets(status.CurrentBetsAmount, status.CurrentBets)
	fmt.Printf("  locked bets: %d\n", status.LockedBetsAmount)
	dumpBets(status.LockedBetsAmount, status.LockedBets)
	fmt.Printf("  last winning color: %d\n", status.LastWinningColor)
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

func dumpBets(n uint16, bets []*fairroulette.BetInfo) {
	if len(bets) < int(n) {
		fmt.Printf("    (showing first %d)\n", len(bets))
	}
	for i, bet := range bets {
		fmt.Printf("    %d: %s\n", i, bet.String())
	}
}
