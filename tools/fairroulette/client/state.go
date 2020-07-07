package main

import (
	"fmt"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
)

func dumpState(scAddress string) {
	state, err := waspapi.QuerySCState(config.waspApi, scAddress, []kv.Key{
		kv.ArraySizeKey(fairroulette.StateVarBets),       // array
		kv.ArraySizeKey(fairroulette.StateVarLockedBets), // array
		fairroulette.StateVarLastWinningColor,            // int64
		fairroulette.StateVarEntropyFromLocking,          // hash
		fairroulette.VarPlayPeriodSec,                    // int64
	})
	check(err)

	codec := state.Codec()

	fmt.Printf("FairRoulette Smart Contract State:\n")

	nBets := codec.GetArray(fairroulette.StateVarBets).Len()
	fmt.Printf("bets: %d\n", nBets)

	nLockedBets := codec.GetArray(fairroulette.StateVarLockedBets).Len()
	fmt.Printf("locked bets: %d\n", nLockedBets)

	lastwc, _, err := codec.GetInt64(fairroulette.StateVarLastWinningColor)
	check(err)
	fmt.Printf("last winning color: %d\n", lastwc)

	entropy, _, err := codec.GetHashValue(fairroulette.StateVarEntropyFromLocking)
	check(err)
	fmt.Printf("entropy: %s\n", entropy)

	playPeriod, _, err := codec.GetInt64(fairroulette.VarPlayPeriodSec)
	check(err)
	fmt.Printf("play period (s): %d\n", playPeriod)
}
