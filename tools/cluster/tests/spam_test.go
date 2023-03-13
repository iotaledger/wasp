package tests

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/utxodb"
)

// executed in cluster_test.go
func testSpamOnledger(t *testing.T, env *ChainEnv) {
	testutil.RunHeavy(t)
	env.deployNativeIncCounterSC(0)
	// in the privtangle setup, with 1s milestones, this test takes ~50m to process 10k requests
	const numRequests = 10_000

	// send requests from many different wallets to speed things up
	numAccounts := 1000
	numRequestsPerAccount := numRequests / numAccounts
	errCh := make(chan error, numRequests)
	txCh := make(chan iotago.Transaction, numRequests)
	for i := 0; i < numAccounts; i++ {
		keyPair, _, err := env.Clu.NewKeyPairWithFunds()
		createWalletRetries := 0
		if err != nil {
			if createWalletRetries >= 5 {
				t.Fatal("failed to create wallet, got an error 5 times, %w", err)
			}
			// wait and re-try
			createWalletRetries++
			i--
			time.Sleep(1 * time.Second)
			continue
		}
		go func() {
			chainClient := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)
			retries := 0
			for i := 0; i < numRequestsPerAccount; i++ {
				tx, err := chainClient.PostRequest(inccounter.FuncIncCounter.Name)
				if err != nil {
					if retries >= 5 {
						errCh <- fmt.Errorf("failed to issue tx, an error 5 times, %w", err)
						break
					}
					// wait and re-try the tx
					retries++
					i--
					time.Sleep(1 * time.Second)
					continue
				}
				retries = 0
				errCh <- err
				txCh <- *tx
				time.Sleep(1 * time.Second) // give time for the indexer to get the new UTXOs (so we don't issue conflicting txs)
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

	res, _, err := env.Chain.Cluster.WaspClient(0).CorecontractsApi.BlocklogGetEventsOfLatestBlock(context.Background(), env.Chain.ChainID.String()).Execute()
	require.NoError(t, err)

	println(res.Events)
}

// executed in cluster_test.go
func testSpamOffLedger(t *testing.T, env *ChainEnv) {
	testutil.RunHeavy(t)
	env.deployNativeIncCounterSC(0)

	// we need to cap the limit of parallel requests, otherwise some reqs will fail due to local tcp limits: `dial tcp 127.0.0.1:9090: socket: too many open files`
	const maxParallelRequests = 700
	const numRequests = 100_000

	// deposit funds for offledger requests
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	env.DepositFunds(utxodb.FundsFromFaucetAmount, keyPair)

	myClient := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)

	durationsMutex := sync.Mutex{}
	processingDurationsSum := uint64(0)
	maxProcessingDuration := uint64(0)

	maxChan := make(chan int, maxParallelRequests)
	reqSuccessChan := make(chan uint64, numRequests)
	reqErrorChan := make(chan error, 1)

	go func() {
		for i := 0; i < numRequests; i++ {
			maxChan <- i
			nonce := uint64(i + 1)
			go func() {
				// send the request
				req, er := myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: nonce})
				if er != nil {
					reqErrorChan <- er
					return
				}
				reqSentTime := time.Now()
				// wait for the request to be processed
				_, err = env.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), 5*time.Minute)
				if err != nil {
					reqErrorChan <- err
					return
				}
				processingDuration := uint64(time.Since(reqSentTime).Seconds())
				reqSuccessChan <- nonce
				<-maxChan

				durationsMutex.Lock()
				defer durationsMutex.Unlock()
				processingDurationsSum += processingDuration
				if processingDuration > maxProcessingDuration {
					maxProcessingDuration = processingDuration
				}
			}()
		}
	}()

	n := 0
	for {
		select {
		case <-reqSuccessChan:
			n++
		case e := <-reqErrorChan:
			// no request should fail
			fmt.Printf("ERROR sending offledger request, err: %v\n", e)
			t.Fatal(e)
		}
		if n == numRequests {
			break
		}
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), []int{0}, 5*time.Minute)

	res, _, err := env.Chain.Cluster.WaspClient(0).CorecontractsApi.BlocklogGetEventsOfLatestBlock(context.Background(), env.Chain.ChainID.String()).Execute()
	require.NoError(t, err)

	require.Regexp(t, fmt.Sprintf("counter = %d", numRequests), res.Events[len(res.Events)-1])
	avgProcessingDuration := processingDurationsSum / numRequests
	fmt.Printf("avg processing duration: %ds\n max: %ds\n", avgProcessingDuration, maxProcessingDuration)
}

// executed in cluster_test.go
func testSpamCallViewWasm(t *testing.T, env *ChainEnv) {
	testutil.RunHeavy(t)

	wallet, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	client := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), wallet)
	{
		// increment counter once
		tx, err := client.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
		_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, 30*time.Second)
		require.NoError(t, err)
	}

	const n = 200
	ch := make(chan error, n)

	for i := 0; i < n; i++ {
		go func() {
			r, err := client.CallView(context.Background(), "getCounter", nil)
			if err != nil {
				ch <- err
				return
			}

			v, err := codec.DecodeInt64(r.MustGet(inccounter.VarCounter))
			if err == nil && v != 1 {
				err = errors.New("v != 1")
			}
			ch <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-ch
		if err != nil {
			t.Error(err)
		}
	}
}
