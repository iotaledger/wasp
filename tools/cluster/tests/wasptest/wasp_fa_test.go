package wasptest

import (
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"testing"
	"time"
)

const scNumFairAuction = 5

// sending 5 NOP requests with 1 sec sleep between each
func TestFairAuctionSetOwnerMargin(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

	err = PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := []*waspapi.RequestBlockJson{
		{Address: sc.Address,
			RequestCode: fairauction.RequestSetOwnerMargin,
			Vars: map[string]interface{}{
				fairauction.VarReqOwnerMargin: 100,
			},
		},
	}
	err = SendRequestsNTimes(wasps, sc.OwnerIndexUtxodb, 1, reqs, 0)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}
