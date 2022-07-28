package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/stretchr/testify/require"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
)

type nanoClientTest struct {
	id       int
	messages []string
}

func (c *nanoClientTest) start(t *testing.T, url string) {
	sock, err := sub.NewSocket()
	require.NoError(t, err)

	err = sock.Dial(url)
	require.NoError(t, err)
	// Empty byte array effectively subscribes to everything
	err = sock.SetOption(mangos.OptionSubscribe, []byte(""))
	require.NoError(t, err)

	for {
		msg, err := sock.Recv()
		require.NoError(t, err)
		c.messages = append(c.messages, string(msg))
	}
}

func TestNanoPublisher(t *testing.T) {
	// single wasp node committee, to test if publishing can break state transitions
	env := setupAdvancedInccounterTest(t, 1, []int{0})

	// spawn many NANOMSG nanoClients and subscribe to everything from the node
	nanoClients := make([]nanoClientTest, 10)
	nanoURL := fmt.Sprintf("tcp://127.0.0.1:%d", env.Clu.Config.NanomsgPort(0))
	for i := range nanoClients {
		nanoClients[i] = nanoClientTest{id: i, messages: []string{}}
		go nanoClients[i].start(t, nanoURL)
	}

	// deposit funds for offledger requests
	keyPair, _, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	accountsClient := env.Chain.SCClient(accounts.Contract.Hname(), keyPair)
	reqTx, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: isc.NewFungibleBaseTokens(1_000_000),
	})
	require.NoError(t, err)

	_, err = env.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, reqTx, 30*time.Second)
	require.NoError(t, err)

	// send 100 requests
	numRequests := 100
	myClient := env.Chain.SCClient(isc.Hn(nativeIncCounterSCName), keyPair)

	reqIDs := make([]isc.RequestID, numRequests)
	for i := 0; i < numRequests; i++ {
		req, err := myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 1)})
		reqIDs[i] = req.ID()
		require.NoError(t, err)
	}

	for _, reqID := range reqIDs {
		_, err = env.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(env.Chain.ChainID, reqID, 30*time.Second)
		require.NoError(t, err)
	}

	waitUntil(t, env.counterEquals(int64(numRequests)), util.MakeRange(0, 1), 60*time.Second, "requests counted")

	// assert all clients received the correct number of messages
	for _, client := range nanoClients {
		println(len(client.messages))
	}

	// send 100 requests

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 101)})
		require.NoError(t, err)
	}

	waitUntil(t, env.counterEquals(int64(numRequests*2)), util.MakeRange(0, 1), 60*time.Second, "requests counted")

	time.Sleep(10 * time.Second)
	for _, client := range nanoClients {
		println(len(client.messages))
	}
}
