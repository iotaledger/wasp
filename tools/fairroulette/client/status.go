package client

import (
	"fmt"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func StatusCmd(args []string) {
	state, err := waspapi.QuerySCState(config.WaspApi(), config.GetSCAddress().String(), []kv.Key{
		kv.ArraySizeKey(fairroulette.StateVarBets),       // array
		kv.ArraySizeKey(fairroulette.StateVarLockedBets), // array
		fairroulette.StateVarLastWinningColor,            // int64
		fairroulette.StateVarEntropyFromLocking,          // hash
		fairroulette.VarPlayPeriodSec,                    // int64
	})
	check(err)

	codec := state.Codec()

	fmt.Printf("FairRoulette Smart Contract status:\n")

	nBets := codec.GetArray(fairroulette.StateVarBets).Len()
	fmt.Printf("  bets: %d\n", nBets)

	nLockedBets := codec.GetArray(fairroulette.StateVarLockedBets).Len()
	fmt.Printf("  locked bets: %d\n", nLockedBets)

	lastwc, _, err := codec.GetInt64(fairroulette.StateVarLastWinningColor)
	check(err)
	fmt.Printf("  last winning color: %d\n", lastwc)

	playPeriod, _, err := codec.GetInt64(fairroulette.VarPlayPeriodSec)
	check(err)
	fmt.Printf("  play period (s): %d\n", playPeriod)
}
