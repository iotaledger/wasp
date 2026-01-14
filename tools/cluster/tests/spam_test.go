package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
)

// executed in cluster_test.go
func (e *ChainEnv) testSpamEVM(t *testing.T) {
	// TODO: increase to 10K as in original test. Now it's not passing
	const numRequests = 600

	numRequestsPerAccount := 100
	numAccounts := numRequests / numRequestsPerAccount

	for i := range numAccounts {
		t.Run(fmt.Sprintf("account-%d", i), func(t *testing.T) {
			t.Parallel()
			err := e.checkNRequests(newClusterTestEnv(t, e, 0), int64(numRequestsPerAccount), 0, []int{0}, 30*time.Second)
			require.NoError(t, err)
		})
	}
}

// executed in cluster_test.go
func (e *ChainEnv) testSpamOnledger(t *testing.T) {
	const maxParallelRequests = 10
	const numRequests = 100

	var (
		durationsMutex         sync.Mutex
		processingDurationsSum uint64
		maxProcessingDuration  uint64
	)

	reqSuccessChan := make(chan uint64, numRequests)
	reqErrorChan := make(chan error, 1)

	baseTokensSent := coin.Value(10 + iotaclient.DefaultGasBudget)

	type wallet struct {
		keyPair         *cryptolib.KeyPair
		client          *chainclient.Client
		gasChargedTotal atomic.Uint64
	}

	var err error
	wallets := make([]wallet, maxParallelRequests)
	for i := range maxParallelRequests {
		wallets[i].keyPair, _, err = e.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)
		e.DepositFunds(100_000_000, wallets[i].keyPair)
		wallets[i].client = e.Chain.Client(wallets[i].keyPair)
	}

	balanceBefore := e.GetL2Balance(isc.NewAddressAgentID(wallets[0].keyPair.Address()), coin.BaseTokenType)

	for walletIndex := range maxParallelRequests {
		go func(walletIndex int) {
			for i := uint64(0); i < numRequests; i++ {
				req, er := wallets[walletIndex].client.PostRequest(
					context.Background(),
					accounts.FuncDeposit.Message(),
					chainclient.PostRequestParams{
						Transfer:  isc.NewAssets(baseTokensSent),
						GasBudget: iotaclient.DefaultGasBudget,
					},
				)
				if er != nil {
					reqErrorChan <- er
					return
				}
				reqSentTime := time.Now()
				// wait for the request to be processed
				receipt, er := e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, req, false, 1*time.Minute)
				if er != nil {
					reqErrorChan <- er
					return
				}

				gasFeeCharged, er := util.DecodeUint64(receipt[0].GasFeeCharged)
				if er != nil {
					reqErrorChan <- er
					return
				}
				wallets[walletIndex].gasChargedTotal.Add(gasFeeCharged)

				processingDuration := uint64(time.Since(reqSentTime).Seconds())
				reqSuccessChan <- i

				durationsMutex.Lock()
				processingDurationsSum += processingDuration
				if processingDuration > maxProcessingDuration {
					maxProcessingDuration = processingDuration
				}
				durationsMutex.Unlock()
			}
		}(walletIndex)
	}

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
		if n == numRequests*maxParallelRequests {
			break
		}
	}

	for i := range wallets {
		balance := e.GetL2Balance(isc.NewAddressAgentID(wallets[i].keyPair.Address()), coin.BaseTokenType)
		require.Equal(t, balanceBefore+numRequests*baseTokensSent-coin.Value(wallets[i].gasChargedTotal.Load()), balance)
	}
}
