package client

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
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

func (s *Status) NextPlayIn() string {
	diff := s.NextPlayTimestamp.Sub(time.Now())
	// round to the second
	diff -= diff % time.Second
	if diff < 0 {
		return "unknown"
	}
	return diff.String()
}

func FetchStatus(addresses []string) (*Status, error) {
	status := &Status{}

	players := make([]address.Address, 0)
	for _, addr := range addresses {
		addr, err := address.FromBase58(addr)
		if err != nil {
			return nil, err
		}
		players = append(players, addr)
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
	if err != nil {
		return nil, err
	}

	codec := vars.MustCodec()
	status.CurrentBetsAmount = codec.GetArray(fairroulette.StateVarBets).Len()
	status.LockedBetsAmount = codec.GetArray(fairroulette.StateVarLockedBets).Len()
	status.LastWinningColor, _ = codec.GetInt64(fairroulette.StateVarLastWinningColor)
	status.PlayPeriodSeconds, _ = codec.GetInt64(fairroulette.VarPlayPeriodSec)
	nextPlayTimestamp, _ := codec.GetInt64(fairroulette.VarNextPlayTimestamp)
	status.NextPlayTimestamp = time.Unix(0, nextPlayTimestamp)
	if err != nil {
		return nil, err
	}

	status.PlayerStats, err = decodePlayerStats(vars, players, playerKeys)
	if err != nil {
		return nil, err
	}

	status.WinsPerColor, err = decodeWinsPerColor(vars, winsPerColorKeys)
	if err != nil {
		return nil, err
	}

	// in a second query we fetch the array items
	betsKeys := kv.ArrayRangeKeys(fairroulette.StateVarBets, status.CurrentBetsAmount, 0, BetsSliceLength)
	lockedBetsKeys := kv.ArrayRangeKeys(fairroulette.StateVarLockedBets, status.LockedBetsAmount, 0, 10)
	vars, err = waspapi.QuerySCState(
		config.WaspApi(),
		config.GetSCAddress().String(),
		append(betsKeys, lockedBetsKeys...),
	)
	if err != nil {
		return nil, err
	}

	status.CurrentBets, err = decodeBets(vars, betsKeys)
	if err != nil {
		return nil, err
	}
	status.LockedBets, err = decodeBets(vars, lockedBetsKeys)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func decodeBets(state kv.Map, keys []kv.Key) ([]*fairroulette.BetInfo, error) {
	bets := make([]*fairroulette.BetInfo, 0)
	for _, k := range keys {
		b, err := state.Get(k)
		if err != nil {
			return nil, err
		}
		bet, err := fairroulette.DecodeBetInfo(b)
		if err != nil {
			return nil, err
		}
		bets = append(bets, bet)
	}
	return bets, nil
}

func decodeWinsPerColor(vars kv.Map, winsPerColorKeys []kv.Key) ([]uint32, error) {
	ret := make([]uint32, 0)
	for _, key := range winsPerColorKeys {
		b, err := vars.Get(key)
		if err != nil {
			return nil, err
		}
		var n uint32
		if b != nil {
			n = util.Uint32From4Bytes(b)
		}
		ret = append(ret, n)
	}
	return ret, nil
}

func decodePlayerStats(vars kv.Map, players []address.Address, playerKeys []kv.Key) (map[address.Address]*fairroulette.PlayerStats, error) {
	playerStats := make(map[address.Address]*fairroulette.PlayerStats)
	for i, addr := range players {
		v, err := vars.Get(playerKeys[i])
		if err != nil {
			return nil, err
		}
		ps, err := fairroulette.DecodePlayerStats(v)
		if err != nil {
			return nil, err
		}
		playerStats[addr] = ps
	}
	return playerStats, nil
}
