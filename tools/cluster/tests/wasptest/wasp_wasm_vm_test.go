package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

// sending 5 NOP requests with 1 sec sleep between each
func TestWasmVMSend5Requests1Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestWasmVMSend5Requests1Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2, // wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[4]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()
	ownerAddress := sc.OwnerAddress()

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: wasmpoc.RequestNop,
		})
		check(err, t)
		time.Sleep(1 * time.Second)
	}

	wasps.CollectMessages(10 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddress, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  []byte(wasmpoc.ProgramHash),
	}) {
		t.Fail()
	}
}

func TestWasmSend1ReqIncSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestWasmSend1ReqIncSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1, // we dont know if it will go in same batch with init request or separate
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[4]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestInc,
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 2, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(1)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
	}) {
		t.Fail()
	}
}

func TestWasmSend1ReqIncRepeatSuccessTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncRepeatTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          4,
		"request_out":         5,
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

	scAddress := sc.SCAddress()

	// send 1i to the SC address. It is needed to send the request to self
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestNop,
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	check(err, t)
	time.Sleep(1 * time.Second)

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestIncRepeat1,
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 2, map[balance.Color]int64{
		balance.ColorIOTA: 1,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(2)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestWasmChainIncTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestChainIncTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          chainOfRequestsLength + 3,
		"request_out":         chainOfRequestsLength + 4,
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

	scAddress := sc.SCAddress()

	// send 5i to the SC address. It is needed to send 5 requests to self at once
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestNop,
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5,
		},
	})
	check(err, t)
	time.Sleep(1 * time.Second)

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestIncRepeatMany,
		Vars: map[string]interface{}{
			inccounter.ArgNumRepeats: chainOfRequestsLength,
		},
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 6, map[balance.Color]int64{
		balance.ColorIOTA: 5,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(chainOfRequestsLength + 1)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
		inccounter.VarNumRepeats:    util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}

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
		SCAddress:   &scAddress,
		RequestCode: wasmpoc.RequestPlaceBet,
		Vars: map[string]interface{}{
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

	scAddress := sc.SCAddress()
	ownerAddr := sc.OwnerAddress()

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}) {
		t.Fail()
		return
	}

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: wasmpoc.RequestPlaceBet,
			Vars: map[string]interface{}{
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

	if !wasps.VerifyAddressBalances(scAddress, 501, map[balance.Color]int64{
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

	scAddress := sc.SCAddress()
	ownerAddr := sc.OwnerAddress()

	// send 1i to the SC address. It is needed to send the request to self ("operating capital")
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestNop,
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
		SCAddress:   scAddress,
		RequestCode: wasmpoc.RequestPlayPeriod,
		Vars: map[string]interface{}{
			"playPeriod": 10,
		},
	})
	check(err, t)

	time.Sleep(1 * time.Second)

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: wasmpoc.RequestPlaceBet,
			Vars: map[string]interface{}{
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
	if !wasps.VerifyAddressBalances(scAddress, 2, map[balance.Color]int64{
		scColor:           1,
		balance.ColorIOTA: 1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
	}) {
		t.Fail()
	}
}
