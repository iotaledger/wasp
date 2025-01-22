package tests

import (
	"context"
	"testing"
	"time"

	"github.com/iotaledger/wasp/clients/chainclient"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/contracts/inccounter"
)

// ensures a nodes resumes normal operation after rebooting
func TestInxShutdownTest(t *testing.T) {
	t.Skip("Cluster tests currently disabled")
	
	dataPath := "test-inx-down"
	env := setupNativeInccounterTest(t, 4, []int{0, 1, 2, 3}, dataPath)

	// restart the privtangle, this will cause an INX disconnection on wasp
	//l1.Stop()

	// assert wasp nodes are down
	_, err := env.Clu.MultiClient().NodeVersion()
	require.Error(t, err)
	require.Regexp(t, `connection refused`, err.Error())

	// start privatangle again
	//l1.StartExistingServers()

	// start the nodes again
	err = env.Clu.Start()
	require.NoError(t, err)

	// assert wasp starts correctly
	_, err = env.Clu.MultiClient().NodeVersion()
	require.NoError(t, err)

	// assert requests are processed
	client := env.createNewClient()

	tx, err := client.PostRequest(context.Background(), inccounter.FuncIncCounter.Message(nil), chainclient.PostRequestParams{})
	require.NoError(t, err)

	_, err = apiextensions.APIWaitUntilAllRequestsProcessed(context.Background(), env.Clu.WaspClient(0), env.Chain.ChainID, tx, true, 10*time.Second)
	require.NoError(t, err)

	env.expectCounter(1)
}
