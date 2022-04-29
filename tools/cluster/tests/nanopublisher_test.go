package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
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
	nanoURL := fmt.Sprintf("tcp://127.0.0.1:%d", env.clu.Config.NanomsgPort(0))
	for i := range nanoClients {
		nanoClients[i] = nanoClientTest{id: i, messages: []string{}}
		go nanoClients[i].start(t, nanoURL)
	}

	// deposit funds for offledger requests
	keyPair, myAddress, err := env.clu.NewKeyPairWithFunds()
	require.NoError(t, err)
	myAgentID := iscp.NewAgentID(myAddress, 0)

	accountsClient := env.chain.SCClient(accounts.Contract.Hname(), keyPair)
	_, err = accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: iscp.NewTokensIotas(10000),
	})
	require.NoError(t, err)

	waitUntil(t, env.balanceOnChainIotaEquals(myAgentID, 10000), util.MakeRange(0, 1), 60*time.Second, "send 100i")

	// send 100 requests
	numRequests := 100
	myClient := env.chain.SCClient(iscp.Hn(incCounterSCName), keyPair)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 1)})
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
