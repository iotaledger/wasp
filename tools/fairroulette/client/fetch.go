package client

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

const BetsSliceLength = 10

type Status struct {
	SCBalance int64

	CurrentBetsAmount uint16
	CurrentBets       []*fairroulette.BetInfo

	LockedBetsAmount uint16
	LockedBets       []*fairroulette.BetInfo

	LastWinningColor int64

	PlayPeriodSeconds int64

	FetchedAt         time.Time
	NextPlayTimestamp time.Time

	PlayerStats map[address.Address]*fairroulette.PlayerStats

	WinsPerColor []uint32
}

func (s *Status) NextPlayIn() string {
	diff := s.NextPlayTimestamp.Sub(s.FetchedAt)
	// round to the second
	diff -= diff % time.Second
	if diff < 0 {
		return "unknown"
	}
	return diff.String()
}

func FetchStatus() (*Status, error) {
	status := &Status{}

	balance, err := fetchSCBalance()
	if err != nil {
		return nil, err
	}
	status.SCBalance = balance

	address := config.GetSCAddress()

	query := stateapi.NewQueryRequest(&address)
	query.AddArray(fairroulette.StateVarBets, 0, 100)
	query.AddArray(fairroulette.StateVarLockedBets, 0, 100)
	query.AddInt64(fairroulette.StateVarLastWinningColor)
	query.AddInt64(fairroulette.ReqVarPlayPeriodSec)
	query.AddInt64(fairroulette.StateVarNextPlayTimestamp)
	query.AddDictionary(fairroulette.StateVarPlayerStats, 100)
	query.AddArray(fairroulette.StateArrayWinsPerColor, 0, fairroulette.NumColors)

	results, err := waspapi.QuerySCState(config.WaspApi(), query)
	if err != nil {
		return nil, err
	}

	status.LastWinningColor = results[fairroulette.StateVarLastWinningColor].MustInt64()

	status.PlayPeriodSeconds = results[fairroulette.ReqVarPlayPeriodSec].MustInt64()
	if status.PlayPeriodSeconds == 0 {
		status.PlayPeriodSeconds = fairroulette.DefaultPlaySecondsAfterFirstBet
	}

	status.FetchedAt = time.Now().UTC()

	nextPlayTimestamp := results[fairroulette.StateVarNextPlayTimestamp].MustInt64()
	status.NextPlayTimestamp = time.Unix(0, nextPlayTimestamp).UTC()

	status.PlayerStats, err = decodePlayerStats(results[fairroulette.StateVarPlayerStats].MustDictionaryResult())
	if err != nil {
		return nil, err
	}

	status.WinsPerColor, err = decodeWinsPerColor(results[fairroulette.StateArrayWinsPerColor].MustArrayResult())
	if err != nil {
		return nil, err
	}

	status.CurrentBetsAmount, status.CurrentBets, err = decodeBets(results[fairroulette.StateVarBets].MustArrayResult())
	if err != nil {
		return nil, err
	}

	status.LockedBetsAmount, status.LockedBets, err = decodeBets(results[fairroulette.StateVarLockedBets].MustArrayResult())
	if err != nil {
		return nil, err
	}

	return status, nil
}

func fetchSCBalance() (int64, error) {
	scAddress := config.GetSCAddress()
	outs, err := nodeapi.GetAccountOutputs(config.GoshimmerApi(), &scAddress)
	if err != nil {
		return 0, err
	}
	byColor, _ := util.OutputBalancesByColor(outs)
	b, _ := byColor[balance.ColorIOTA]
	return b, nil
}

func decodeBets(result *stateapi.ArrayResult) (uint16, []*fairroulette.BetInfo, error) {
	size := result.Len
	bets := make([]*fairroulette.BetInfo, 0)
	for _, b := range result.Values {
		bet, err := fairroulette.DecodeBetInfo(b)
		if err != nil {
			return 0, nil, err
		}
		bets = append(bets, bet)
	}
	return size, bets, nil
}

func decodeWinsPerColor(result *stateapi.ArrayResult) ([]uint32, error) {
	ret := make([]uint32, 0)
	for _, b := range result.Values {
		var n uint32
		if b != nil {
			n = util.Uint32From4Bytes(b)
		}
		ret = append(ret, n)
	}
	return ret, nil
}

func decodePlayerStats(result *stateapi.DictResult) (map[address.Address]*fairroulette.PlayerStats, error) {
	playerStats := make(map[address.Address]*fairroulette.PlayerStats)
	for _, e := range result.Entries {
		if len(e.Key) != address.Length {
			return nil, fmt.Errorf("not an address: %v", e.Key)
		}
		addr, _, err := address.FromBytes(e.Key)
		if err != nil {
			return nil, err
		}
		ps, err := fairroulette.DecodePlayerStats(e.Value)
		if err != nil {
			return nil, err
		}
		playerStats[addr] = ps
	}
	return playerStats, nil
}
