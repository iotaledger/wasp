package tests

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
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

// func TestNanoPublisher(t *testing.T) {
// 	// single wasp node committee, to test if publishing can break state transitions
// 	env := setupAdvancedInccounterTest(t, 1, []int{0})

// 	// spawn many NANOMSG nanoClients and subscribe to everything from the node
// 	nanoClients := make([]nanoClientTest, 10)
// 	nanoURL := fmt.Sprintf("tcp://127.0.0.1:%d", env.clu.Config.NanomsgPort(0))
// 	for i := range nanoClients {
// 		nanoClients[i] = nanoClientTest{id: i, messages: []string{}}
// 		go nanoClients[i].start(t, nanoURL)
// 	}

// 	// deposit funds for offledger requests
// 	keyPair, myAddress := env.getOrCreateAddress()
// 	myAgentID := iscp.NewAgentID(myAddress, 0)

// 	accountsClient := env.chain.SCClient(accounts.Contract.Hname(), keyPair)
// 	_, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
// 		Transfer: colored.NewBalancesForIotas(10000),
// 	})
// 	require.NoError(t, err)

// 	waitUntil(t, env.balanceOnChainIotaEquals(myAgentID, 10000), util.MakeRange(0, 1), 60*time.Second, "send 100i")

// 	// send 100 requests
// 	numRequests := 100
// 	myClient := env.chain.SCClient(iscp.Hn(incCounterSCName), keyPair)

// 	for i := 0; i < numRequests; i++ {
// 		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 1)})
// 		require.NoError(t, err)
// 	}

// 	waitUntil(t, env.counterEquals(int64(numRequests)), util.MakeRange(0, 1), 60*time.Second, "requests counted")

// 	// assert all clients received the correct number of messages
// 	for _, client := range nanoClients {
// 		println(len(client.messages))
// 	}

// 	// send 100 requests

// 	for i := 0; i < numRequests; i++ {
// 		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 101)})
// 		require.NoError(t, err)
// 	}

// 	waitUntil(t, env.counterEquals(int64(numRequests*2)), util.MakeRange(0, 1), 60*time.Second, "requests counted")

// 	time.Sleep(10 * time.Second)
// 	for _, client := range nanoClients {
// 		println(len(client.messages))
// 	}
// }

const numberParam = "number"

func TestNanoPublisherFairRoulette(t *testing.T) {
	// deploy a single node chain with fairroulette contract
	clu := newCluster(t, 1)
	committee := []int{0}
	quorum := uint16(1)
	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.Base58())

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID.Base58())

	require.NoError(t, err)

	scBinary, err := ioutil.ReadFile("./wasm/fairroulette_bg.wasm")
	require.NoError(t, err)
	_, _, err = chain.DeployWasmContract("fairroulette", "fairroulette", scBinary, map[string]interface{}{})
	require.NoError(t, err)

	chEnv := &chainEnv{
		env:   &env{t: t, clu: clu},
		chain: chain,
	}
	waitUntil(t, chEnv.contractIsDeployed("fairroulette"), clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")

	// spawn many NANOMSG nanoClients and subscribe to everything from the node
	nanoClients := make([]nanoClientTest, 0)
	nanoURL := fmt.Sprintf("tcp://127.0.0.1:%d", chEnv.clu.Config.NanomsgPort(0))
	for i := range nanoClients {
		nanoClients[i] = nanoClientTest{id: i, messages: []string{}}
		go nanoClients[i].start(t, nanoURL)
	}

	// deposit funds for offledger requests
	keyPair, myAddress := chEnv.getOrCreateAddress()
	myAgentID := iscp.NewAgentID(myAddress, 0)

	accountsClient := chEnv.chain.SCClient(accounts.Contract.Hname(), keyPair)
	_, err = accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: colored.NewBalancesForIotas(1000000),
	})
	require.NoError(t, err)

	waitUntil(t, chEnv.balanceOnChainIotaEquals(myAgentID, 1000000), util.MakeRange(0, 1), 60*time.Second, "send 1000000i")

	// ----------------------------------------
	// otherWallet, _ := chEnv.getOrCreateAddress()

	// _, err = chEnv.chain.SCClient(accounts.Contract.Hname(), otherWallet).PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
	// 	Transfer: colored.NewBalancesForIotas(1000000),
	// 	Args:     requestargs.New().AddEncodeSimple(accounts.ParamAgentID, codec.EncodeAgentID(myAgentID)),
	// })
	// require.NoError(t, err)

	// waitUntil(t, chEnv.balanceOnChainIotaEquals(myAgentID, 2000000), util.MakeRange(0, 1), 60*time.Second, "send 1000000i")
	// ----------------------------------------

	// send N requests
	numRequests := 1000000
	myClient := chEnv.chain.SCClient(iscp.Hn("fairroulette"), keyPair)

	repeatNtimes := 1

	for i := 0; i < repeatNtimes; i++ {
		for i := 0; i < numRequests; i++ {
			placeBet(int64(3), myClient, t)
		}

		time.Sleep(20 * time.Second)
		totalMsgs := len(nanoClients[0].messages)
		for _, client := range nanoClients {
			require.Len(t, client.messages, totalMsgs)
		}
	}
	println("!!!!!!!!!!!!!!!!!!!!!!!!!!")
	println(errors)
}

var (
	nonce  = uint64(1)
	errors = 0
)

func placeBet(number int64, myClient *scclient.SCClient, t *testing.T) {
	args := requestargs.New().AddEncodeSimple(numberParam, codec.EncodeInt64(number))
	// nonce := uint64(time.Now().UnixNano())
	params := chainclient.PostRequestParams{Args: args, Nonce: nonce}
	nonce++
	_, err := myClient.PostOffLedgerRequest("placeBet", *params.WithIotas(1))
	if err != nil {
		errors++
	}
	// require.NoError(t, err)
}
