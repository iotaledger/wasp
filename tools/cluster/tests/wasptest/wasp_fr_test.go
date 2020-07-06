package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSend1Bet(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1Bet")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           5,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          3,
		"request_out":         3,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[3]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := []*waspapi.RequestBlockJson{
		{
			Address:     sc.Address,
			RequestCode: fairroulette.RequestPlaceBet,
			AmountIotas: 10001,
			Vars: map[string]interface{}{
				fairroulette.ReqVarColor: 3,
			},
		},
	}
	err = SendRequestsNTimes(wasps, sc.OwnerIndexUtxodb, 1, reqs, 0*time.Second)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	tmp, err := valuetransaction.IDFromBase58(sc.Color)
	assert.NoError(t, err)
	scColor := (balance.Color)(tmp)

	if !wasps.VerifySCAccountBalances(sc, map[balance.Color]int64{
		balance.ColorIOTA: 10001,
		scColor:           1,
	}) {
		t.Fail()
	}
}

func TestSend5Bets(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1Bet")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           5,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          7,
		"request_out":         7,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[3]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := MakeRequests(5, func(i int) *waspapi.RequestBlockJson {
		return &waspapi.RequestBlockJson{
			Address:     sc.Address,
			RequestCode: fairroulette.RequestPlaceBet,
			AmountIotas: 10001,
			Vars: map[string]interface{}{
				fairroulette.ReqVarColor: i,
			},
		}
	})
	for _, req := range reqs {
		err = SendRequestsNTimes(wasps, sc.OwnerIndexUtxodb, 1,
			[]*waspapi.RequestBlockJson{req}, 0*time.Second)
	}
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	tmp, err := valuetransaction.IDFromBase58(sc.Color)
	assert.NoError(t, err)
	scColor := (balance.Color)(tmp)

	if !wasps.VerifySCAccountBalances(sc, map[balance.Color]int64{
		balance.ColorIOTA: 50005,
		scColor:           1,
	}) {
		t.Fail()
	}
}
