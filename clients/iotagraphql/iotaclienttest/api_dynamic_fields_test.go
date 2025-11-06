package iotaclienttest

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
)

func TestGetDynamicFields(t *testing.T) {
	// t.Skip("Requires a valid object with dynamic fields")

	ctx := context.Background()
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	// Use a test address as parent object ID
	// In a real test, this should be an actual object that has dynamic fields
	parentID := iotago.MustAddressFromHex("0x0a78f4a0195ef76c4c15f5d8d4014725be09021cd2f9817b5e501f14a42e62b8")

	resp, err := client.GetDynamicFields(ctx, iotaclient.GetDynamicFieldsRequest{
		ParentObjectID: parentID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Log the response for inspection
	t.Logf("Dynamic Fields Response: %+v", resp)
	t.Logf("Number of dynamic fields: %d", len(resp.Data))

	// Check pagination info
	t.Logf("Has next page: %v", resp.HasNextPage)
}

func TestGetDynamicFieldsPagination(t *testing.T) {
	t.Skip("Requires a valid object with dynamic fields")

	ctx := context.Background()
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	// First, get an object that we know has dynamic fields
	ownerPtr := iotago.MustAddressFromHex(testcommon.TestAddress)
	require.NoError(t, iotaclient.RequestFundsFromFaucet(ctx, ownerPtr, iotaconn.AlphanetFaucetURL))
	owner := ownerPtr

	// Test pagination by requesting a small number of fields at a time
	resp, err := client.GetDynamicFields(ctx, iotaclient.GetDynamicFieldsRequest{
		ParentObjectID: owner,
		Limit:          lo.ToPtr(uint(2)),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// If there are dynamic fields and pagination indicates more
	if len(resp.Data) > 0 && resp.HasNextPage {
		// Get the next page
		resp2, err := client.GetDynamicFields(ctx, iotaclient.GetDynamicFieldsRequest{
			ParentObjectID: owner,
			Limit:          lo.ToPtr(uint(2)),
			Cursor:         resp.NextCursor,
		})
		require.NoError(t, err)
		require.NotNil(t, resp2)

		t.Logf("First page: %d fields, Second page: %d fields",
			len(resp.Data),
			len(resp2.Data))
	}
}
