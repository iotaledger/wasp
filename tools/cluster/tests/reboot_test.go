package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/native/inccounter"
)

// ensures a nodes resumes normal operation after rebooting
func TestReboot(t *testing.T) {
	e := setupAdvancedInccounterTest(t, 3, []int{0, 1, 2})
	client := e.createNewClient()

	_, err := client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	e.counterEquals(1)

	// restart the nodes
	err = e.Clu.RestartNodes(0, 1, 2)
	require.NoError(t, err)

	// after rebooting, the chain should resume processing requests without issues
	_, err = client.PostRequest(inccounter.FuncIncCounter.Name)
	require.NoError(t, err)
	e.counterEquals(2)
}
