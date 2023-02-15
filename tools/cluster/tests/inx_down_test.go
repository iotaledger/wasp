package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
)

// ensures a nodes resumes normal operation after rebooting
func TestInxShutdownTest(t *testing.T) {
	t.Skip()
	env := setupNativeInccounterTest(t, 4, []int{0, 1, 2, 3})

	// restart the privtangle, this should cause INX
	l1.Stop()

	// assert wasp nodes are down
	r, err := env.Clu.MultiClient().NodeVersion()
	require.Error(t, err)
	require.Regexp(t, `connection refused`, err.Error())
	require.Nil(t, r)

	// start privatangle again
	l1.StartServers()

	// start the nodes again
	env.Clu.Start(env.dataPath)

	// assert wasp starts correctly
	_, err = env.Clu.MultiClient().NodeVersion()
	require.NoError(t, err)

	// assert everything works
	client := env.createNewClient()

	tx, err := client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)

	_, err = apiextensions.APIWaitUntilAllRequestsProcessed(env.Clu.WaspClient(0), env.Chain.ChainID, tx, 10*time.Second)
	require.NoError(t, err)

	env.expectCounter(nativeIncCounterSCHname, 1)
}
