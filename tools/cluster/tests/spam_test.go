package tests

import (
	"fmt"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/testcore"
	"github.com/stretchr/testify/require"
)

const numRequests = 100_000

// TODO this is currently broken - the indexer can't keep up and returns UTXOs that were already spent. - to revisit after updating L1 dependencies
func TestSpamOnledger(t *testing.T) {
	testutil.RunHeavy(t)
	env := setupAdvancedInccounterTest(t, 1, []int{0})

	// send requests from many different wallets to speed things up
	numAccounts := 100
	numRequestsPerAccount := numRequests / numAccounts
	errCh := make(chan error, numRequests)
	txCh := make(chan iotago.Transaction, numRequests)
	for i := 0; i < numAccounts; i++ {
		keyPair, _, err := env.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)
		go func() {
			chainClient := env.Chain.SCClient(iscp.Hn(incCounterSCName), keyPair)
			for i := 0; i < numRequestsPerAccount; i++ {
				tx, err := chainClient.PostRequest(inccounter.FuncIncCounter.Name)
				errCh <- err
				txCh <- *tx
				time.Sleep(300 * time.Millisecond) // give time for the indexer to get the new UTXOs (so we don't issue conflicting txs)
			}
		}()
	}

	// wait for all requests to be sent
	for i := 0; i < numRequests; i++ {
		err := <-errCh
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < numRequests; i++ {
		tx := <-txCh
		_, err := env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, &tx, 30*time.Second)
		require.NoError(t, err)
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), []int{0}, 5*time.Minute)

	res, err := env.Chain.Cluster.WaspClient(0).CallView(env.Chain.ChainID, blocklog.Contract.Hname(), blocklog.ViewGetEventsForBlock.Name, dict.Dict{})
	require.NoError(t, err)
	events, err := testcore.EventsViewResultToStringArray(res)
	require.NoError(t, err)
	println(events)
}

// we need to cap the limit of parallel requests, otherwise some reqs will fail due to local tcp limits: `dial tcp 127.0.0.1:9090: socket: too many open files`
const maxParallelRequests = 700

// !! WARNING !! - this test should only be run with `database.inMemory` set to `false`. Otherwise it is WAY slower, and will probably time out or take a LONG time
func TestSpamOffledger(t *testing.T) {
	testutil.RunHeavy(t)
	// single wasp node committee, to test if publishing can break state transitions
	env := setupAdvancedInccounterTest(t, 1, []int{0})

	// deposit funds for offledger requests
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	accountsClient := env.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	tx, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(100_000_000),
	})
	require.NoError(t, err)
	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 30*time.Second)
	require.NoError(t, err)

	myClient := env.Chain.SCClient(iscp.Hn(incCounterSCName), keyPair)

	maxChan := make(chan int, maxParallelRequests)
	reqSuccessChan := make(chan uint64, numRequests)
	reqErrorChan := make(chan error, 1)

	for i := 0; i < numRequests; i++ {
		maxChan <- i
		go func(nonce uint64) {
			// send the request
			req, er := myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: nonce})
			if er != nil {
				reqErrorChan <- er
				return
			}

			// wait for the request to be processed
			_, err = env.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 30*time.Second)
			if err != nil {
				reqErrorChan <- err
				return
			}
			reqSuccessChan <- nonce
			<-maxChan
		}(uint64(i + 1))
	}

	n := 0
	for {
		select {
		case <-reqSuccessChan:
			n++
		case e := <-reqErrorChan:
			// no request should fail
			fmt.Printf("ERROR sending offledger request, err: %v\n", e)
			t.Fatal()
		}
		if n == numRequests {
			break
		}
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), []int{0}, 5*time.Minute)

	res, err := env.Chain.Cluster.WaspClient(0).CallView(env.Chain.ChainID, blocklog.Contract.Hname(), blocklog.ViewGetEventsForBlock.Name, dict.Dict{})
	require.NoError(t, err)
	events, err := testcore.EventsViewResultToStringArray(res)
	require.NoError(t, err)
	println(events)
}
