package wasptest

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
)

func TestLogsc(t *testing.T) {
	// setup
	wasps := setup(t, "logsc_cluster", "TestLogsc")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               3,
		"logsc-addlog":        1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]
	err = putScData(sc, wasps)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequests(wasps, sc, 1, logsc.RequestCodeAddLog, map[string]string{
		"message": "hi",
	}, 0)
	check(err, t)

	wasps.CollectMessages(30 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}
