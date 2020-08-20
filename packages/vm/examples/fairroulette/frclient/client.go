package frclient

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

type FairRouletteClient struct {
	nodeClient nodeclient.NodeClient
	waspHost   string
	scAddress  *address.Address
	sigScheme  signaturescheme.SignatureScheme
}

func NewClient(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address, sigScheme signaturescheme.SignatureScheme) *FairRouletteClient {
	return &FairRouletteClient{nodeClient, waspHost, scAddress, sigScheme}
}

type Status struct {
	SCBalance map[balance.Color]int64
	FetchedAt time.Time

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
	diff := s.NextPlayTimestamp.Sub(s.FetchedAt)
	// round to the second
	diff -= diff % time.Second
	if diff < 0 {
		return "unknown"
	}
	return diff.String()
}

func (frc *FairRouletteClient) FetchStatus() (*Status, error) {
	status := &Status{
		FetchedAt: time.Now().UTC(),
	}

	balance, err := frc.fetchSCBalance()
	if err != nil {
		return nil, err
	}
	status.SCBalance = balance

	query := stateapi.NewQueryRequest(frc.scAddress)
	query.AddArray(fairroulette.StateVarBets, 0, 100)
	query.AddArray(fairroulette.StateVarLockedBets, 0, 100)
	query.AddInt64(fairroulette.StateVarLastWinningColor)
	query.AddInt64(fairroulette.ReqVarPlayPeriodSec)
	query.AddInt64(fairroulette.StateVarNextPlayTimestamp)
	query.AddDictionary(fairroulette.StateVarPlayerStats, 100)
	query.AddArray(fairroulette.StateArrayWinsPerColor, 0, fairroulette.NumColors)

	results, err := waspapi.QuerySCState(frc.waspHost, query)
	if err != nil {
		return nil, err
	}

	lastWinningColor, _, err := results[fairroulette.StateVarLastWinningColor].MustInt64()
	if err != nil {
		return nil, err
	}
	status.LastWinningColor = lastWinningColor

	playPeriodSeconds, _, err := results[fairroulette.ReqVarPlayPeriodSec].MustInt64()
	if err != nil {
		return nil, err
	}
	status.PlayPeriodSeconds = playPeriodSeconds
	if status.PlayPeriodSeconds == 0 {
		status.PlayPeriodSeconds = fairroulette.DefaultPlaySecondsAfterFirstBet
	}

	nextPlayTimestamp, _, err := results[fairroulette.StateVarNextPlayTimestamp].MustInt64()
	if err != nil {
		return nil, err
	}
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

func (frc *FairRouletteClient) fetchSCBalance() (map[balance.Color]int64, error) {
	outs, err := frc.nodeClient.GetAccountOutputs(frc.scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
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

func (frc *FairRouletteClient) postRequest(code sctransaction.RequestCode, amountIotas int64, vars map[string]interface{}) (*sctransaction.Transaction, error) {
	tx, err := waspapi.CreateRequestTransaction(
		frc.nodeClient,
		frc.sigScheme,
		[]*waspapi.RequestBlockJson{{
			Address:     frc.scAddress.String(),
			RequestCode: code,
			AmountIotas: amountIotas,
			Vars:        vars,
		}},
	)
	if err != nil {
		return nil, err
	}
	return tx, frc.nodeClient.PostTransaction(tx.Transaction)
}

func (frc *FairRouletteClient) Bet(color int, amount int) (*sctransaction.Transaction, error) {
	return frc.postRequest(fairroulette.RequestPlaceBet, int64(amount), map[string]interface{}{
		fairroulette.ReqVarColor: int64(color),
	})
}

func (frc *FairRouletteClient) SetPeriod(seconds int) (*sctransaction.Transaction, error) {
	return frc.postRequest(fairroulette.RequestSetPlayPeriod, 0, map[string]interface{}{
		fairroulette.ReqVarPlayPeriodSec: int64(seconds),
	})
}
