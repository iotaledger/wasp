package client

import (
	"fmt"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func StatusCmd(args []string) {
	// for arrays we fetch the length in the first query
	state, err := waspapi.QuerySCState(config.WaspApi(), config.GetSCAddress().String(), []kv.Key{
		kv.ArraySizeKey(fairroulette.StateVarBets),       // array
		kv.ArraySizeKey(fairroulette.StateVarLockedBets), // array
		fairroulette.StateVarLastWinningColor,            // int64
		fairroulette.StateVarEntropyFromLocking,          // hash
		fairroulette.VarPlayPeriodSec,                    // int64
	})
	check(err)

	codec := state.Codec()
	nBets := codec.GetArray(fairroulette.StateVarBets).Len()
	nLockedBets := codec.GetArray(fairroulette.StateVarLockedBets).Len()
	lastwc, _, err := codec.GetInt64(fairroulette.StateVarLastWinningColor)
	check(err)
	playPeriod, _, err := codec.GetInt64(fairroulette.VarPlayPeriodSec)
	check(err)

	// in a second query we fetch the items
	betsKeys := kv.ArrayRangeKeys(fairroulette.StateVarBets, nBets, 0, 10)
	lockedBetsKeys := kv.ArrayRangeKeys(fairroulette.StateVarLockedBets, nLockedBets, 0, 10)
	state, err = waspapi.QuerySCState(config.WaspApi(), config.GetSCAddress().String(), append(betsKeys, lockedBetsKeys...))
	check(err)

	bets := decodeBets(state, betsKeys)
	lockedBets := decodeBets(state, lockedBetsKeys)

	fmt.Printf("FairRoulette Smart Contract status:\n")
	fmt.Printf("  bets: %d\n", nBets)
	dumpBets(bets)
	fmt.Printf("  locked bets: %d\n", nLockedBets)
	dumpBets(lockedBets)
	fmt.Printf("  last winning color: %d\n", lastwc)
	fmt.Printf("  play period (s): %d\n", playPeriod)
}

func decodeBets(state kv.Map, keys []kv.Key) []*fairroulette.BetInfo {
	bets := make([]*fairroulette.BetInfo, 0)
	for _, k := range keys {
		b, err := state.Get(k)
		check(err)
		bet, err := fairroulette.DecodeBetInfo(b)
		check(err)
		bets = append(bets, bet)
	}
	return bets
}

func dumpBets(bets []*fairroulette.BetInfo) {
	if len(bets) > 0 {
		fmt.Printf("    first 10:\n")
	}
	for i, bet := range bets {
		fmt.Printf("      %d: %s\n", i, bet.String())
	}
}
