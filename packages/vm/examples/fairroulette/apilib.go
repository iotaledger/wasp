package fairroulette

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

type Status struct {
	SCBalance map[balance.Color]int64

	CurrentBetsAmount uint16
	CurrentBets       []*BetInfo

	LockedBetsAmount uint16
	LockedBets       []*BetInfo

	LastWinningColor int64

	PlayPeriodSeconds int64

	FetchedAt         time.Time
	NextPlayTimestamp time.Time

	PlayerStats map[address.Address]*PlayerStats

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

func FetchStatus(nodeclient nodeclient.NodeClient, waspHost string, scAddress *address.Address) (*Status, error) {
	status := &Status{}

	balance, err := fetchSCBalance(nodeclient, scAddress)
	if err != nil {
		return nil, err
	}
	status.SCBalance = balance

	query := stateapi.NewQueryRequest(scAddress)
	query.AddArray(StateVarBets, 0, 100)
	query.AddArray(StateVarLockedBets, 0, 100)
	query.AddInt64(StateVarLastWinningColor)
	query.AddInt64(ReqVarPlayPeriodSec)
	query.AddInt64(StateVarNextPlayTimestamp)
	query.AddDictionary(StateVarPlayerStats, 100)
	query.AddArray(StateArrayWinsPerColor, 0, NumColors)

	results, err := waspapi.QuerySCState(waspHost, query)
	if err != nil {
		return nil, err
	}

	status.LastWinningColor = results[StateVarLastWinningColor].MustInt64()

	status.PlayPeriodSeconds = results[ReqVarPlayPeriodSec].MustInt64()
	if status.PlayPeriodSeconds == 0 {
		status.PlayPeriodSeconds = DefaultPlaySecondsAfterFirstBet
	}

	status.FetchedAt = time.Now().UTC()

	nextPlayTimestamp := results[StateVarNextPlayTimestamp].MustInt64()
	status.NextPlayTimestamp = time.Unix(0, nextPlayTimestamp).UTC()

	status.PlayerStats, err = decodePlayerStats(results[StateVarPlayerStats].MustDictionaryResult())
	if err != nil {
		return nil, err
	}

	status.WinsPerColor, err = decodeWinsPerColor(results[StateArrayWinsPerColor].MustArrayResult())
	if err != nil {
		return nil, err
	}

	status.CurrentBetsAmount, status.CurrentBets, err = decodeBets(results[StateVarBets].MustArrayResult())
	if err != nil {
		return nil, err
	}

	status.LockedBetsAmount, status.LockedBets, err = decodeBets(results[StateVarLockedBets].MustArrayResult())
	if err != nil {
		return nil, err
	}

	return status, nil
}

func fetchSCBalance(nodeclient nodeclient.NodeClient, scAddress *address.Address) (map[balance.Color]int64, error) {
	outs, err := nodeclient.GetAccountOutputs(scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}

func decodeBets(result *stateapi.ArrayResult) (uint16, []*BetInfo, error) {
	size := result.Len
	bets := make([]*BetInfo, 0)
	for _, b := range result.Values {
		bet, err := DecodeBetInfo(b)
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

func decodePlayerStats(result *stateapi.DictResult) (map[address.Address]*PlayerStats, error) {
	playerStats := make(map[address.Address]*PlayerStats)
	for _, e := range result.Entries {
		if len(e.Key) != address.Length {
			return nil, fmt.Errorf("not an address: %v", e.Key)
		}
		addr, _, err := address.FromBytes(e.Key)
		if err != nil {
			return nil, err
		}
		ps, err := DecodePlayerStats(e.Value)
		if err != nil {
			return nil, err
		}
		playerStats[addr] = ps
	}
	return playerStats, nil
}
