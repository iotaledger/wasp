package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/testutil"
	"testing"
	"time"
)

func TestWasmSend1Bet(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestWasmSend1Bet")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          3,
		"request_out":         3,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[4]

	scColor, err := PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)
	scAddress, err := address.FromBase58(sc.Address)
	check(err, t)

	ownerAddr := sc.OwnerAddress()
	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}) {
		t.Fail()
		return
	}

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: &scAddress,
		Vars: map[string]interface{}{
			"wasm":  "fairroulette",
			"fn":    "placeBet",
			"color": 3,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 100,
		},
	})
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	scAddr := sc.SCAddress()

	if !wasps.VerifyAddressBalances(scAddr, 101, map[balance.Color]int64{
		balance.ColorIOTA: 99,
		*scColor:          1,
		// +1 more pending in self sent timelocked request
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1-100, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 100,
	}) {
		t.Fail()
	}

}

func TestWasmSend5Bets(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestWasmSend5Bets")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          7,
		"request_out":         7,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[4]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress, err := address.FromBase58(sc.Address)
	check(err, t)
	ownerAddr := sc.OwnerAddress()
	check(err, t)

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}) {
		t.Fail()
		return
	}

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: &scAddress,
			Vars: map[string]interface{}{
				"wasm":  "fairroulette",
				"fn":    "placeBet",
				"color": i + 1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 100,
			},
		})
	}
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	scColor := sc.GetColor()

	if !wasps.VerifyAddressBalances(&scAddress, 501, map[balance.Color]int64{
		balance.ColorIOTA: 499, // one request sent to itself
		scColor:           1,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1-500, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 500,
	}) {
		t.Fail()
	}
}

func TestWasmSendBetsAndPlay(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestWasmSendBetsAndPlay")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          10,
		"request_out":         11,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[4]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress, err := address.FromBase58(sc.Address)
	check(err, t)
	ownerAddr := sc.OwnerAddress()
	check(err, t)

	// send 1i to the SC address. It is needed to send the request to self ("operating capital")
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: &scAddress,
		Vars: map[string]interface{}{
			"wasm": "fairroulette",
			"fn":   "nothing",
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	check(err, t)
	time.Sleep(1 * time.Second)

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
	}) {
		t.Fail()
	}
	// SetPlayPeriod must be processed first
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: &scAddress,
		Vars: map[string]interface{}{
			"wasm":       "fairroulette",
			"fn":         "playPeriod",
			"playPeriod": 10,
		},
	})
	check(err, t)

	time.Sleep(1 * time.Second)

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: &scAddress,
			Vars: map[string]interface{}{
				"wasm":  "fairroulette",
				"fn":    "placeBet",
				"color": i + 1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 100,
			},
		})
	}
	check(err, t)

	wasps.CollectMessages(30 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	scColor := sc.GetColor()

	if !wasps.VerifyAddressBalances(&scAddress, 2, map[balance.Color]int64{
		balance.ColorIOTA: 1,
		scColor:           1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
	}) {
		t.Fail()
	}
}
