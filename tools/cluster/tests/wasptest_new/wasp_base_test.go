package wasptest

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/assert"
)

func TestPutBootupRecord(t *testing.T) {
	clu := setup(t, "test_cluster", "TestPutBootupRecord")

	err := clu.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    0,
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	sc := &clu.SmartContractConfig[0]

	_, err = PutBootupRecord(clu, sc)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

// TODO deploy cycle with chainclient

func TestDeployChain(t *testing.T) {
	clu := setup(t, "test_cluster", "TestDeployChain")

	err := clu.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)

	sc := &clu.SmartContractConfig[0]

	err = clu.NodeClient.RequestFunds(sc.OwnerAddress())
	check(err, t)

	_, _, _, err = clu.DeployChain(sc, clu.Config.SmartContracts[0].Quorum)
	check(err, t)

	if !clu.WaitUntilExpectationsMet() {
		t.Fail()
	}

	clu.WithSCState(sc, 0, func(host string, stateIndex uint32, state dict.Dict) bool {
		fmt.Printf("%s\n", state)
		assert.EqualValues(t, 1, stateIndex)
		{
			state := codec.NewMustCodec(state)

			assert.EqualValues(t, []byte{0xFF}, state.Get(root.VarStateInitialized))

			chid, _ := state.GetChainID(root.VarChainID)
			assert.EqualValues(t, sc.ChainID(), chid)

			desc, _ := state.GetString(root.VarDescription)
			assert.EqualValues(t, []byte(sc.Description), desc)
		}
		return true
	})
}

//
//func TestActivate1Chain(t *testing.T) {
//	clu := setup(t, "test_cluster", "TestActivate1Chain")
//
//	err := clu.ListenToMessages(map[string]int{
//		"bootuprec":           2,
//		"active_committee":    1,
//		"dismissed_committee": 0,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	sc := &clu.SmartContractConfig[0]
//
//	_, err = PutBootupRecord(clu, sc)
//	check(err, t)
//
//	err = Activate1Chain(clu, sc)
//	check(err, t)
//
//	if !clu.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
//
//func TestActivateAllChains(t *testing.T) {
//	clu := setup(t, "test_cluster", "TestActivateAllSC")
//
//	err := clu.ListenToMessages(map[string]int{
//		"bootuprec":           clu.NumSmartContracts() * 2,
//		"active_committee":    clu.NumSmartContracts(),
//		"dismissed_committee": 0,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	for _, sc := range clu.SmartContractConfig {
//		_, err = PutBootupRecord(clu, &sc)
//		check(err, t)
//	}
//
//	err = ActivateAllSC(clu)
//	check(err, t)
//
//	if !clu.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
//
//func TestDeactivate1Chain(t *testing.T) {
//	clu := setup(t, "test_cluster", "TestDeactivate1SC")
//
//	err := clu.ListenToMessages(map[string]int{
//		"bootuprec":           3,
//		"active_committee":    1,
//		"dismissed_committee": 1,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	sc := &clu.SmartContractConfig[0]
//
//	_, err = PutBootupRecord(clu, sc)
//	check(err, t)
//
//	err = Activate1Chain(clu, sc)
//	check(err, t)
//
//	time.Sleep(5 * time.Second)
//
//	err = Deactivate1Chain(clu, sc)
//	check(err, t)
//
//	if !clu.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
