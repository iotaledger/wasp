package client

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
	"github.com/iotaledger/wasp/tools/fairroulette/wallet"
)

const BetsSliceLength = 10

type Status struct {
	CurrentBetsAmount uint16
	CurrentBets       []*fairroulette.BetInfo

	LockedBetsAmount uint16
	LockedBets       []*fairroulette.BetInfo

	LastWinningColor int64

	PlayPeriodSeconds int64
	NextPlayTimestamp time.Time

	PlayerStats map[address.Address]*fairroulette.PlayerStats

	WinsPerColor []uint32
}

func FetchStatus(addresses []string) *Status {
	status := &Status{}

	players := make([]address.Address, 0)
	if len(addresses) == 0 {
		players = append(players, wallet.Load().Address())
	} else {
		for _, addr := range addresses {
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
	status.CurrentBetsAmount = codec.MustGetArray(fairroulette.StateVarBets).Len()
	status.LockedBetsAmount = codec.MustGetArray(fairroulette.StateVarLockedBets).Len()
	status.LastWinningColor, _, err = codec.GetInt64(fairroulette.StateVarLastWinningColor)
	check(err)
	status.PlayPeriodSeconds, _, err = codec.GetInt64(fairroulette.VarPlayPeriodSec)
	check(err)
	nextPlayTimestamp, _, err := codec.GetInt64(fairroulette.VarNextPlayTimestamp)
	status.NextPlayTimestamp = time.Unix(0, nextPlayTimestamp)
	check(err)

	status.PlayerStats = decodePlayerStats(vars, players, playerKeys)

	status.WinsPerColor = decodeWinsPerColor(vars, winsPerColorKeys)

	// in a second query we fetch the array items
	betsKeys := kv.ArrayRangeKeys(fairroulette.StateVarBets, status.CurrentBetsAmount, 0, BetsSliceLength)
	lockedBetsKeys := kv.ArrayRangeKeys(fairroulette.StateVarLockedBets, status.LockedBetsAmount, 0, 10)
	vars, err = waspapi.QuerySCState(
		config.WaspApi(),
		config.GetSCAddress().String(),
		append(betsKeys, lockedBetsKeys...),
	)
	check(err)

	status.CurrentBets = decodeBets(vars, betsKeys)
	status.LockedBets = decodeBets(vars, lockedBetsKeys)

	return status
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
