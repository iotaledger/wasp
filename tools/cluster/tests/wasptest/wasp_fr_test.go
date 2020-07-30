package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
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
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          3,
		"request_out":         3,
		"state":               -1,
	})
	check(err, t)

	scColors, err := PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[3]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)
	scAddress, err := address.FromBase58(sc.Address)
	check(err, t)

	ownerAddr := sc.OwnerAddress()

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: fairroulette.RequestPlaceBet,
		Vars: map[string]interface{}{
			fairroulette.ReqVarColor: 3,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1001,
		},
	})
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	scColor := *scColors[sc.Address]

	scAddr, err := address.FromBase58(sc.Address)
	assert.NoError(t, err)

	if !wasps.VerifyAddressBalances(scAddr, 1002, map[balance.Color]int64{
		balance.ColorIOTA: 1000,
		scColor:           1,
		// +1 more pending timelocked request
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, 1000000000-1-1000-1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1 - 1000 - 1,
	}) {
		t.Fail()
	}

}

func TestSend5Bets(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend5Bets")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          7,
		"request_out":         7,
		"state":               -1,
	})
	check(err, t)

	scColors, err := PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[3]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress, err := address.FromBase58(sc.Address)
	check(err, t)
	ownerAddr := sc.OwnerAddress()
	check(err, t)

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddress,
			RequestCode: fairroulette.RequestPlaceBet,
			Vars: map[string]interface{}{
				fairroulette.ReqVarColor: i,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 1000,
			},
		})
	}
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	scColor := *scColors[sc.Address]

	if !wasps.VerifyAddressBalances(scAddress, 5001, map[balance.Color]int64{
		balance.ColorIOTA: 4999, // one request sent to itself
		scColor:           1,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(ownerAddr, 1000000000-1-5000, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1 - 5000,
	}) {
		t.Fail()
	}
}

func TestSendBetsAndPlay(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSendBetsAndPlay")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          9,
		"request_out":         10,
		"state":               -1,
	})
	check(err, t)

	scColors, err := PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[3]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress, err := address.FromBase58(sc.Address)
	check(err, t)
	ownerAddr := sc.OwnerAddress()
	check(err, t)

	// SetPlayPeriod must be processed first
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: fairroulette.RequestSetPlayPeriod,
		Vars: map[string]interface{}{
			fairroulette.ReqVarPlayPeriodSec: 10,
		},
	})
	check(err, t)

	time.Sleep(1 * time.Second)

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddress,
			RequestCode: fairroulette.RequestPlaceBet,
			Vars: map[string]interface{}{
				fairroulette.ReqVarColor: i,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 1000,
			},
		})
	}
	check(err, t)

	wasps.CollectMessages(30 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	scColor := *scColors[sc.Address]
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		scColor: 1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, 1000000000-1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1,
	}) {
		t.Fail()
	}
}
