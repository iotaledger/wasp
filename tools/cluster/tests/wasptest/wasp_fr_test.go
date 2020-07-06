package wasptest

import (
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"testing"
	"time"
)

func TestSend1Bet(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1Bet")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               2,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[0]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := []*waspapi.RequestBlockJson{
		{Address: sc.Address,
			RequestCode: vmconst.RequestCodeNOP,
		},
	}
	err = SendRequestsNTimes(wasps, sc.OwnerIndexUtxodb, 1, reqs, 0*time.Second)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}
