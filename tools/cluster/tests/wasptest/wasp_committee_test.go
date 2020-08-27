package wasptest

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

func TestKillNode(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestSend1ReqIncSimple")

	sc := &wasps.SmartContractConfig[2]

	_, err := PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	wasps.StopNode(3)

	scAddress := sc.SCAddress()

	_, err = waspapi.RunAndWaitForRequestProcessedMulti(wasps.ActiveApiHosts(), &scAddress, 0, 30*time.Second, func() (*sctransaction.Transaction, error) {
		return waspapi.CreateSimpleRequest(wasps.NodeClient, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddress,
			RequestCode: inccounter.RequestInc,
		})
	})
	check(err, t)

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
