package wasptest

import (
	"fmt"
	"strings"
	"testing"

	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/logsc"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
)

// FIXME move log tests to the same keys.json file with other tests

func startLogSC(t *testing.T, expectations map[string]int) (*cluster.Cluster, *cluster.SmartContractFinalConfig) {
	clu := setup(t, "logsc_cluster", "TestLogsc")

	err := clu.ListenToMessages(expectations)
	check(err, t)

	sc := &clu.SmartContractConfig[0]

	_, err = PutBootupRecord(clu, sc)
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
		"state":               -1,
		"logsc-addlog":        -1,
	})

	reqs := []*waspapi.RequestBlockJson{{
		Address:     sc.Address,
		RequestCode: logsc.RequestCodeAddLog,
		Vars: map[string]interface{}{
			"message": "message 0",
		},
	}}
	err := SendRequestsNTimes(clu, sc.OwnerSigScheme(), 1, reqs)
	check(err, t)

	clu.WaitUntilExpectationsMet()

	if !clu.Report() {
		t.Fail()
	}

	if !clu.VerifySCState(sc, 2, map[kv.Key][]byte{
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
		"state":               -1,
		"logsc-addlog":        -1,
	})

	reqs := MakeRequests(5, func(i int) *waspapi.RequestBlockJson {
		return &waspapi.RequestBlockJson{
			Address:     sc.Address,
			RequestCode: logsc.RequestCodeAddLog,
			Vars: map[string]interface{}{
				"message": fmt.Sprintf("message %d", i),
			},
		}
	})
	err := SendRequestsNTimes(clu, sc.OwnerSigScheme(), 1, reqs)
	check(err, t)

	clu.WaitUntilExpectationsMet()

	if !clu.Report() {
		t.Fail()
	}

	clu.WithSCState(sc, func(host string, stateIndex uint32, state kv.Map) bool {
		{
			state := state.ToGoMap()
			assert.EqualValues(t, 8, len(state)) // 5 log items + log length + program_hash + owner address
			assert.Equal(t, util.Uint64To8Bytes(uint64(5)), state["log"])
			foundValues := make(map[string]bool)
			for i := 0; i < 5; i++ {
				key := kv.Key(fmt.Sprintf("log:%d", i))
				value := string(state[key])
				assert.NotNil(t, state[key])
				foundValues[value] = true
				assert.True(t, strings.HasPrefix(value, "message "))
			}
			assert.EqualValues(t, 5, len(foundValues))
		}
		return true
	})
}
