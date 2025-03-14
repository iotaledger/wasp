package parameters_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestInitL1(t *testing.T) {
	client := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL, iotaclient.WaitForEffectsDisabled)
	logger := testlogger.NewLogger(t)
	l1params0 := parameters.L1().Clone()
	err := parameters.InitL1(*client, logger)
	require.NoError(t, err)
	l1params1 := parameters.L1().Clone()
	b0, err := bcs.Marshal[parameters.L1Params](l1params0)
	require.NoError(t, err)
	b1, err := bcs.Marshal[parameters.L1Params](l1params1)
	require.NoError(t, err)
	require.NotEqual(t, b0, b1)
}
