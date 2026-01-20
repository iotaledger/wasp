package iotaclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetLatestIotaSystemState(t *testing.T) {
	client := l1starter.Instance().L1Client()
	state, err := client.GetLatestIotaSystemState(context.Background())
	require.NoError(t, err)
	require.NotNil(t, state)
}

func TestGetReferenceGasPrice(t *testing.T) {
	client := l1starter.Instance().L1Client()
	gasPrice, err := client.GetReferenceGasPrice(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, gasPrice.Int64(), int64(1000))
}
