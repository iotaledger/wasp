package wasptest

import (
	"fmt"
	"testing"
	"time"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/tools/cluster"
)

func startLogSC(t *testing.T, expectations map[string]int) (*cluster.Cluster, *cluster.SmartContractFinalConfig) {
	clu := setup(t, "logsc_cluster", "TestLogsc")

	err := clu.ListenToMessages(expectations)
	check(err, t)

	sc := &clu.SmartContractConfig[0]
	err = putScData(sc, clu)
	check(err, t)

	err = Activate1SC(clu, sc)
	check(err, t)

	err = CreateOrigin1SC(clu, sc)
	check(err, t)

	return clu, sc
}

func TestLogsc1(t *testing.T) {
	clu, sc := startLogSC(t, map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               3,
		"logsc-addlog":        1,
	})

	err := SendRequests(clu, sc, makeLogRequests(sc, 1))
	check(err, t)

	clu.CollectMessages(30 * time.Second)

	if !clu.Report() {
		t.Fail()
	}

	if !clu.VerifySCState(sc, 2, map[string][]byte{
		"log":   util.Uint64To8Bytes(uint64(1)),
		"log:0": []byte("message 0"),
	}) {
		t.Fail()
	}
}

func TestLogsc5(t *testing.T) {
	clu, sc := startLogSC(t, map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               3,
		"logsc-addlog":        5,
	})

	err := SendRequests(clu, sc, makeLogRequests(sc, 5))
	check(err, t)

	clu.CollectMessages(30 * time.Second)

	if !clu.Report() {
		t.Fail()
	}

	if !clu.VerifySCState(sc, 2, map[string][]byte{
		"log":   util.Uint64To8Bytes(uint64(5)),
		"log:0": []byte("message 0"),
		"log:1": []byte("message 1"),
		"log:2": []byte("message 2"),
		"log:3": []byte("message 3"),
		"log:4": []byte("message 4"),
	}) {
		t.Fail()
	}
}

func makeLogRequests(sc *cluster.SmartContractFinalConfig, n int) []*waspapi.RequestBlockJson {
	reqs := make([]*waspapi.RequestBlockJson, 0)
	for i := 0; i < n; i++ {
		reqs = append(reqs, &waspapi.RequestBlockJson{
			Address:     sc.Address,
			RequestCode: uint16(logsc.RequestCodeAddLog),
			Vars:        map[string]string{"message": fmt.Sprintf("message %d", i)},
		})
	}

	return reqs
}
