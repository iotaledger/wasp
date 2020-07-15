package client

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
)

func StatusCmd(args []string) {
	players := make([]address.Address, 0)
	if len(args) == 0 {
		players = append(players, wallet.Load().Address())
	} else {
		for _, addr := range args {
			addr, err := address.FromBase58(addr)
			check(err)
			players = append(players, addr)
		}
	}

	// for arrays we fetch the length in the first query
	keys := []kv.Key{
		kv.ArraySizeKey(fairroulette.StateVarBets),       // array
		kv.ArraySizeKey(fairroulette.StateVarLockedBets), // array
		fairroulette.StateVarLastWinningColor,            // int64
		fairroulette.StateVarEntropyFromLocking,          // hash
		fairroulette.VarPlayPeriodSec,                    // int64
		fairroulette.VarNextPlayTimestamp,                // int64
	}

	playerKeys := make([]kv.Key, 0)
	for _, addr := range players {
		key := kv.DictElemKey(fairroulette.VarPlayerStats, addr.Bytes())
		playerKeys = append(playerKeys, key)
	}
	keys = append(keys, playerKeys...)

	winsPerColorKeys := kv.ArrayRangeKeys(fairroulette.VarWinsPerColor, fairroulette.NumColors, 0, fairroulette.NumColors)
	keys = append(keys, winsPerColorKeys...)

	vars, err := waspapi.QuerySCState(config.WaspApi(), config.GetSCAddress().String(), keys)
	check(err)

	codec := vars.Codec()
	nBets := codec.MustGetArray(fairroulette.StateVarBets).Len()
	nLockedBets := codec.MustGetArray(fairroulette.StateVarLockedBets).Len()
	lastwc, _, err := codec.GetInt64(fairroulette.StateVarLastWinningColor)
	check(err)
	playPeriod, _, err := codec.GetInt64(fairroulette.VarPlayPeriodSec)
	check(err)
	nextPlayTimestamp, _, err := codec.GetInt64(fairroulette.VarNextPlayTimestamp)
	check(err)

	playerStats := decodePlayerStats(vars, players, playerKeys)

	winsPerColor := decodeWinsPerColor(vars, winsPerColorKeys)

	// in a second query we fetch the array items
	betsKeys := kv.ArrayRangeKeys(fairroulette.StateVarBets, nBets, 0, 10)
	lockedBetsKeys := kv.ArrayRangeKeys(fairroulette.StateVarLockedBets, nLockedBets, 0, 10)
	vars, err = waspapi.QuerySCState(
		config.WaspApi(),
		config.GetSCAddress().String(),
		append(betsKeys, lockedBetsKeys...),
	)
	check(err)

	bets := decodeBets(vars, betsKeys)
	lockedBets := decodeBets(vars, lockedBetsKeys)

	fmt.Printf("FairRoulette Smart Contract status:\n")
	fmt.Printf("  bets: %d\n", nBets)
	dumpBets(bets)
	fmt.Printf("  locked bets: %d\n", nLockedBets)
	dumpBets(lockedBets)
	fmt.Printf("  last winning color: %d\n", lastwc)
	fmt.Printf("  play period (s): %d\n", playPeriod)
	fmt.Printf("  next play in: %s\n", formatNextPlay(nextPlayTimestamp))
	fmt.Printf("  color stats:\n")
	for color, wins := range winsPerColor {
		fmt.Printf("    color %d: %d wins\n", color, wins)
	}
	if len(playerStats) > 0 {
		fmt.Printf("  player stats:\n")
		for player, stats := range playerStats {
			fmt.Printf("    %s: %s\n", player.String()[:6], stats)
		}
	}
}

func formatNextPlay(ts int64) string {
	diff := time.Unix(0, ts).Sub(time.Now())
	// round to the second
	diff -= diff % time.Second
	if diff < 0 {
		return "unknown"
	}
	return diff.String()
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

func decodeWinsPerColor(vars kv.Map, winsPerColorKeys []kv.Key) []uint32 {
	ret := make([]uint32, 0)
	for _, key := range winsPerColorKeys {
		b, err := vars.Get(key)
		check(err)
		var n uint32
		if b != nil {
			n = util.Uint32From4Bytes(b)
		}
		ret = append(ret, n)
	}
	return ret
}

func decodePlayerStats(vars kv.Map, players []address.Address, playerKeys []kv.Key) map[address.Address]*fairroulette.PlayerStats {
	playerStats := make(map[address.Address]*fairroulette.PlayerStats)
	for i, addr := range players {
		v, err := vars.Get(playerKeys[i])
		check(err)
		ps, err := fairroulette.DecodePlayerStats(v)
		check(err)
		playerStats[addr] = ps
	}
	return playerStats
}
