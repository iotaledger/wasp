package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	// ensure offledger requests are still working
	req, err = client.PostOffLedgerRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	_, err = env.Clu.WaspClient(0).WaitUntilRequestProcessed(env.Chain.ChainID, req.ID(), 10*time.Second)
	require.NoError(t, err)
	env.expectCounter(nativeIncCounterSCHname, 4)
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
