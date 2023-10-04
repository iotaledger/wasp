package tests

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
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
	numAccounts := numRequests / 10
	numRequestsPerAccount := numRequests / numAccounts
	errCh := make(chan error, numRequests)
	txCh := make(chan iotago.Transaction, numRequests)
	for i := 0; i < numAccounts; i++ {
		createWalletRetries := 0

		var keyPair *cryptolib.KeyPair
		for {
			var err error
			keyPair, _, err = env.Clu.NewKeyPairWithFunds()
			if err != nil {
				if createWalletRetries >= 5 {
					t.Fatal("failed to create wallet, got an error 5 times, %w", err)
				}
				// wait and re-try
				createWalletRetries++
				time.Sleep(200 * time.Millisecond)
				continue
			}

			break
		}
		go func() {
			chainClient := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)
			for i := 0; i < numRequestsPerAccount; i++ {
				retries := 0
				for {
					tx, err := chainClient.PostRequest(inccounter.FuncIncCounter.Name)
					if err != nil {
						if retries >= 5 {
							errCh <- fmt.Errorf("failed to issue tx, an error 5 times, %w", err)
							break
						}
						if err.Error() == "no valid inputs found to create transaction" ||
							err.Error() == "block was not included in the ledger. IsTransaction: true, LedgerInclusionState: conflicting, ConflictReason: 1" {
							// wait and retry the tx
							retries++
							time.Sleep(200 * time.Millisecond)
							continue
						}
						errCh <- err // fail if the error is something else
						return
					}
					errCh <- err
					txCh <- *tx
					break
				}
				time.Sleep(200 * time.Millisecond) // give time for the indexer to get the new UTXOs (so we don't issue conflicting txs)
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
		_, err := env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, &tx, false, 30*time.Second)
		require.NoError(t, err)
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), []int{0}, 30*time.Second)

	res, _, err := env.Chain.Cluster.WaspClient(0).CorecontractsApi.BlocklogGetEventsOfLatestBlock(context.Background(), env.Chain.ChainID.String()).Execute()
	require.NoError(t, err)

	eventBytes, err := iotago.DecodeHex(res.Events[len(res.Events)-1].Payload)
	require.NoError(t, err)
	lastEventCounterValue := codec.MustDecodeInt64(eventBytes)
	require.EqualValues(t, lastEventCounterValue, numRequests)
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

	maxChan := make(chan uint64, maxParallelRequests)
	reqSuccessChan := make(chan uint64, numRequests)
	reqErrorChan := make(chan error, 1)

	go func() {
		for i := uint64(0); i < numRequests; i++ {
			maxChan <- i
			go func(nonce uint64) {
				// send the request
				req, er := myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: nonce})
				if er != nil {
					reqErrorChan <- er
					return
				}
				reqSentTime := time.Now()
				// wait for the request to be processed
				_, err = env.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, req.ID(), false, 1*time.Minute)
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
			}(i)
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

	eventBytes, err := iotago.DecodeHex(res.Events[len(res.Events)-1].Payload)
	require.NoError(t, err)
	lastEventCounterValue := codec.MustDecodeInt64(eventBytes)
	require.EqualValues(t, lastEventCounterValue, numRequests)
	avgProcessingDuration := processingDurationsSum / numRequests
	fmt.Printf("avg processing duration: %ds\n max: %ds\n", avgProcessingDuration, maxProcessingDuration)
}

// executed in cluster_test.go
func testSpamCallViewWasm(t *testing.T, env *ChainEnv) {
	testutil.RunHeavy(t)
	env.deployNativeIncCounterSC(0)

	wallet, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	client := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), wallet)
	{
		// increment counter once
		tx, err := client.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
		_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, tx, false, 30*time.Second)
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

			v, err := codec.DecodeInt64(r.Get(inccounter.VarCounter))
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

// executed in cluster_test.go
func testSpamEVM(t *testing.T, env *ChainEnv) {
	testutil.RunHeavy(t)
	env.deployNativeIncCounterSC(0)

	const numRequests = 1_000

	// deposit funds for EVM
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	evmPvtKey, evmAddr := solo.NewEthereumAccount()
	evmAgentID := isc.NewEthereumAddressAgentID(env.Chain.ChainID, evmAddr)
	env.TransferFundsTo(isc.NewAssetsBaseTokens(utxodb.FundsFromFaucetAmount-1*isc.Million), nil, keyPair, evmAgentID)

	// deploy solidity inccounter
	storageContractAddr, storageContractABI := env.DeploySolidityContract(evmPvtKey, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(42))

	initialBlockIndex, err := env.Chain.BlockIndex()
	require.NoError(t, err)

	jsonRPCClient := env.EVMJSONRPClient(0) // send request to node #0
	nonce := env.GetNonceEVM(evmAddr)
	transactions := make([]*types.Transaction, numRequests)
	for i := uint64(0); i < numRequests; i++ {
		// send tx to change the stored value
		callArguments, err2 := storageContractABI.Pack("store", uint32(i))
		require.NoError(t, err2)
		tx, err2 := types.SignTx(
			types.NewTransaction(nonce+i, storageContractAddr, big.NewInt(0), 100000, env.GetGasPriceEVM(), callArguments),
			EVMSigner(),
			evmPvtKey,
		)
		require.NoError(t, err2)
		err2 = jsonRPCClient.SendTransaction(context.Background(), tx)
		require.NoError(t, err2)
		transactions[i] = tx
	}

	// await txs confirmed
	for _, tx := range transactions {
		_, err2 := env.Clu.MultiClient().WaitUntilEVMRequestProcessedSuccessfully(env.Chain.ChainID, tx.Hash(), false, 5*time.Second)
		require.NoError(t, err2)
	}

	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{storageContractAddr},
		FromBlock: big.NewInt(int64(initialBlockIndex + 1)),
		ToBlock:   big.NewInt(int64(initialBlockIndex + numRequests)),
	}

	logs, err := jsonRPCClient.FilterLogs(context.Background(), filterQuery)
	require.NoError(t, err)

	for i, l := range logs {
		t.Logf("log %d is from block %d with tx index %d", i, l.BlockNumber, l.TxIndex)
	}

	t.Logf("len of logs must be %d, is actually %d", numRequests, len(logs))
	require.Len(t, logs, numRequests)
}
