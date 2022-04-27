package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/testcore"
	"github.com/stretchr/testify/require"
)

const numRequests = 100000

func TestSpamOnledger(t *testing.T) {
	testutil.RunHeavy(t)
	env := setupAdvancedInccounterTest(t, 1, []int{0})

	keyPair, _, err := env.clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myClient := env.chain.SCClient(iscp.Hn(incCounterSCName), keyPair)

	for i := 0; i < numRequests; i++ {
		args := chainclient.NewPostRequestParams().WithIotas(1)
		_, err := myClient.PostRequest(inccounter.FuncIncCounter.Name, *args)
		require.NoError(t, err)
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), []int{0}, 5*time.Minute)

	res, err := env.chain.Cluster.WaspClient(0).CallView(env.chain.ChainID, blocklog.Contract.Hname(), blocklog.ViewGetEventsForBlock.Name, dict.Dict{})
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
	keyPair, myAddress, err := env.clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := iscp.NewAgentID(myAddress, 0)

	accountsClient := env.chain.SCClient(accounts.Contract.Hname(), keyPair)
	_, err = accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(1000000),
	})
	require.NoError(t, err)

	waitUntil(t, env.balanceOnChainIotaEquals(myAgentID, 1000000), util.MakeRange(0, 1), 60*time.Second, "send 1000000i")

	myClient := env.chain.SCClient(iscp.Hn(incCounterSCName), keyPair)

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
			_, err = env.chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(env.chain.ChainID, req.ID(), 30*time.Second)
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
			t.FailNow()
		}
		if n == numRequests {
			break
		}
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), []int{0}, 5*time.Minute)

	res, err := env.chain.Cluster.WaspClient(0).CallView(env.chain.ChainID, blocklog.Contract.Hname(), blocklog.ViewGetEventsForBlock.Name, dict.Dict{})
	require.NoError(t, err)
	events, err := testcore.EventsViewResultToStringArray(res)
	require.NoError(t, err)
	println(events)
}
