package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/utxodb"
)

// ensures a nodes resumes normal operation after rebooting
func TestReboot(t *testing.T) {
	env := setupNativeInccounterTest(t, 4, []int{0, 1, 2, 3})
	// env := setupNativeInccounterTest(t, 3, []int{0, 1, 2})
	client := env.createNewClient()

	// ------ TODO why does this make the test fail?
	er := env.Clu.WaspClient(0).DeactivateChain(env.Chain.ChainID)
	require.NoError(t, er)
	er = env.Clu.WaspClient(0).ActivateChain(env.Chain.ChainID)
	require.NoError(t, er)

	er = env.Clu.WaspClient(1).DeactivateChain(env.Chain.ChainID)
	require.NoError(t, er)
	er = env.Clu.WaspClient(1).ActivateChain(env.Chain.ChainID)
	require.NoError(t, er)
	//-------

	tx, err := client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 1)

	req, err := client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 2)

	// // ------ TODO why does this make the test fail?
	// er = env.Clu.WaspClient(0).DeactivateChain(env.Chain.ChainID)
	// require.NoError(t, er)
	// er = env.Clu.WaspClient(0).ActivateChain(env.Chain.ChainID)
	// require.NoError(t, er)

	// tx, err = client.PostRequest(inccounter.FuncIncCounter.Name)
	// require.NoError(t, err)
	// _, err = env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 10*time.Second)
	// require.NoError(t, err)
	// env.expectCounter(nativeIncCounterSCHname, 3)

	// reqx, err := client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	// require.NoError(t, err)
	// env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, reqx.ID(), 10*time.Second)
	// env.expectCounter(nativeIncCounterSCHname, 4)
	// //-------

	// restart the nodes
	err = env.Clu.RestartNodes(0, 1, 2, 3)
	require.NoError(t, err)

	// after rebooting, the chain should resume processing requests without issues
	tx, err = client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 3)
	// ensure offledger requests are still working
	req, err = client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 4)
}

func TestReboot2(t *testing.T) {
	env := setupNativeInccounterTest(t, 4, []int{0, 1, 2, 3})
	// env := setupNativeInccounterTest(t, 3, []int{0, 1, 2})
	client := env.createNewClient()

	tx, err := client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 1)

	req, err := client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 2)

	// ------ TODO why does this make the test fail?
	er := env.Clu.WaspClient(0).DeactivateChain(env.Chain.ChainID)
	require.NoError(t, er)
	er = env.Clu.WaspClient(0).ActivateChain(env.Chain.ChainID)
	require.NoError(t, er)

	er = env.Clu.WaspClient(1).DeactivateChain(env.Chain.ChainID)
	require.NoError(t, er)
	er = env.Clu.WaspClient(1).ActivateChain(env.Chain.ChainID)
	require.NoError(t, er)
	//-------

	// // ------ TODO why does this make the test fail?
	// er = env.Clu.WaspClient(0).DeactivateChain(env.Chain.ChainID)
	// require.NoError(t, er)
	// er = env.Clu.WaspClient(0).ActivateChain(env.Chain.ChainID)
	// require.NoError(t, er)

	// tx, err = client.PostRequest(inccounter.FuncIncCounter.Name)
	// require.NoError(t, err)
	// _, err = env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 10*time.Second)
	// require.NoError(t, err)
	// env.expectCounter(nativeIncCounterSCHname, 3)

	// reqx, err := client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	// require.NoError(t, err)
	// env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, reqx.ID(), 10*time.Second)
	// env.expectCounter(nativeIncCounterSCHname, 4)
	// //-------

	// restart the nodes
	err = env.Clu.RestartNodes(0, 1, 2, 3)
	require.NoError(t, err)

	// after rebooting, the chain should resume processing requests without issues
	tx, err = client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(env.Chain.ChainID, tx, 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 3)
	// ensure off-ledger requests are still working
	req, err = client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 4)
}

type incCounterClient struct {
	expected int64
	t        *testing.T
	env      *ChainEnv
	client   *scclient.SCClient
}

func newIncCounterClient(t *testing.T, env *ChainEnv, client *scclient.SCClient) *incCounterClient {
	return &incCounterClient{t: t, env: env, client: client}
}

func (icc *incCounterClient) MustIncOnLedger() {
	tx, err := icc.client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(icc.t, err)
	_, err = icc.env.Clu.WaspClient(0).WaitUntilAllRequestsProcessed(icc.env.Chain.ChainID, tx, 10*time.Second)
	require.NoError(icc.t, err)
	icc.expected++
	icc.env.expectCounter(nativeIncCounterSCHname, icc.expected)
}

func (icc *incCounterClient) MustIncOffLedger() {
	req, err := icc.client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(icc.t, err)
	_, err = icc.env.Clu.WaspClient(0).WaitUntilRequestProcessed(icc.env.Chain.ChainID, req.ID(), 10*time.Second)
	require.NoError(icc.t, err)
	icc.expected++
	icc.env.expectCounter(nativeIncCounterSCHname, icc.expected)
}

func (icc *incCounterClient) MustIncBoth(onLedgerFirst bool) {
	if onLedgerFirst {
		icc.MustIncOnLedger()
		icc.MustIncOffLedger()
	} else {
		icc.MustIncOffLedger()
		icc.MustIncOnLedger()
	}
}

// Ensures a nodes resumes normal operation after rebooting.
// In this case we have F=0 and N=3, thus any reboot violates the assumptions.
func TestRebootN3Single(t *testing.T) {
	tm := util.NewTimer()
	allNodes := []int{0, 1, 2}
	env := setupNativeInccounterTest(t, 3, allNodes)
	tm.Step("setupNativeInccounterTest")
	client := env.createNewClient()
	tm.Step("createNewClient")

	env.DepositFunds(1_000_000, client.ChainClient.KeyPair) // For Off-ledger requests to pass.
	tm.Step("DepositFunds")

	icc := newIncCounterClient(t, env, client)
	icc.MustIncBoth(true)
	tm.Step("incCounter")

	// Restart all nodes, one by one.
	for _, nodeIndex := range allNodes {
		require.NoError(t, env.Clu.RestartNodes(nodeIndex))
		icc.MustIncBoth(nodeIndex%2 == 1)
		tm.Step(fmt.Sprintf("incCounter-%v", nodeIndex))
	}
	t.Logf("Timing: %v", tm.String())
}

// Ensures a nodes resumes normal operation after rebooting.
// In this case we have F=0 and N=3, thus any reboot violates the assumptions.
// We restart 2 nodes each iteration in this scenario..
func TestRebootN3TwoNodes(t *testing.T) {
	tm := util.NewTimer()
	allNodes := []int{0, 1, 2}
	env := setupNativeInccounterTest(t, 3, allNodes)
	tm.Step("setupNativeInccounterTest")
	client := env.createNewClient()
	tm.Step("createNewClient")

	env.DepositFunds(1_000_000, client.ChainClient.KeyPair) // For Off-ledger requests to pass.
	tm.Step("DepositFunds")

	icc := newIncCounterClient(t, env, client)
	icc.MustIncBoth(true)
	tm.Step("incCounter")

	// Restart all nodes, one by one.
	for _, nodeIndex := range allNodes {
		otherTwo := lo.Filter(allNodes, func(ni int, _ int) bool { return ni != nodeIndex })
		require.NoError(t, env.Clu.RestartNodes(otherTwo...))
		icc.MustIncBoth(nodeIndex%2 == 1)
		tm.Step(fmt.Sprintf("incCounter-%v", nodeIndex))
	}
	t.Logf("Timing: %v", tm.String())
}

// Test rebooting nodes during operation.
func TestRebootDuringTasks(t *testing.T) {
	env := setupNativeInccounterTest(t, 3, []int{0, 1, 2})

	// deposit funds for offledger requests
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	env.DepositFunds(utxodb.FundsFromFaucetAmount, keyPair)
	client := env.Chain.SCClient(nativeIncCounterSCHname, keyPair)

	for i := 0; i < 10000; i++ {
		go func() {
			// ignore any error
			client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
			// require.NoError(t, err)
		}()
	}

	go func() {
		for i := 0; i < 10000; i++ {
			_, err := client.PostRequest(inccounter.FuncIncCounter.Name)
			require.NoError(t, err)
		}
	}()
	for i := 0; i < 10; i++ {
		// restart the nodes
		// TODO test rebooting only 1 node and see if the consensus breaks
		err := env.Clu.RestartNodes(0, 1, 2)
		require.NoError(t, err)
		time.Sleep(8 * time.Second)
	}
	// // after rebooting, the chain should resume processing requests without issues
	// _, err = client.PostRequest(inccounter.FuncIncCounter.Name)
	// require.NoError(t, err)
	// e.expectCounter(nativeIncCounterSCHname,2)

	ret, err := env.Clu.WaspClient(0).CallView(
		env.Chain.ChainID, nativeIncCounterSCHname, inccounter.ViewGetCounter.Name, nil,
	)
	require.NoError(t, err)
	counter, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter), 0)
	require.NoError(t, err)

	println("potato", counter)
}
