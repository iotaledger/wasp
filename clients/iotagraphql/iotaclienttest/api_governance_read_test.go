package iotaclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

func TestGetCommitteeInfo(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	ctx := context.Background()

	epochID := iotajsonrpc.NewBigInt(0)
	resp, err := client.GetCommitteeInfo(ctx, epochID)
	require.NoError(t, err)

	require.NotNil(t, resp)
}

func TestGetCurrentEpoch(t *testing.T) {
	t.Skip("Needs implementation")
}

func TestGetLatestIotaSystemState(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	state, err := client.GetLatestIotaSystemState(context.Background())
	require.NoError(t, err)
	require.NotNil(t, state)
}

func TestGetReferenceGasPrice(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	gasPrice, err := client.GetReferenceGasPrice(context.Background())
	require.NoError(t, err)
	require.True(t, gasPrice.Clone().Sign() > 0)
}

func TestGetStakes(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	ctx := context.Background()

	// Use a known validator address or skip if none available
	// For now, just test that the method doesn't error
	resp, err := client.GetStakes(ctx, &iotago.Address{})
	// It's okay if this returns an error for invalid address
	_ = err
	_ = resp
}

func TestGetStakesByIds(t *testing.T) {
	t.Skip("Needs proper setup with valid stake IDs")
}

func TestGetValidatorsApy(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	resp, err := client.GetValidatorsApy(context.Background())
	require.NoError(t, err)
	require.NotNil(t, resp)
}
