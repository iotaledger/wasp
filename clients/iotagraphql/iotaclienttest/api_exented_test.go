package iotaclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
)

func TestGetDynamicFields(t *testing.T) {
	ctx := context.Background()
	client := iotagraphql.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)

	// Test object that contains dynamic fields on testnet
	testObjectID := iotago.MustObjectIDFromHex("0xabe5833dcc82909869439112ff1fe5090bcb7cc0f22f6a5bf9241e3a864f7e3c")

	t.Run("GetObject", func(t *testing.T) {
		obj, err := client.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: testObjectID,
			Options: &iotajsonrpc.IotaObjectDataOptions{
				ShowContent: true,
				ShowType:    true,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, obj)
		require.NotNil(t, obj.Data)

		t.Logf("Object exists: %+v", obj.Data)
		if obj.Data.Content != nil {
			t.Logf("Object content: %+v", obj.Data.Content)
		}
	})

	t.Run("GetDynamicFields", func(t *testing.T) {
		resp, err := client.GetDynamicFields(ctx, iotaclient.GetDynamicFieldsRequest{
			ParentObjectID: testObjectID,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		t.Logf("Dynamic Fields Response: %+v", resp)
		t.Logf("Number of dynamic fields: %d", len(resp.Data))
		t.Logf("Has next page: %v", resp.HasNextPage)

		// Verify we got dynamic fields
		require.NotEmpty(t, resp.Data, "Expected to find dynamic fields on this object")

		// Log details about the first few dynamic fields
		for i, field := range resp.Data {
			if i >= 3 {
				break
			}
			t.Logf("Dynamic field %d: Name=%+v, Type=%+v, ObjectType=%s",
				i, field.Name, field.Type, field.ObjectType)
		}
	})
}

func TestGetOwnedObjects(t *testing.T) {
	ctx := context.Background()
	// Use the dynamically mapped port from the local test instance
	client := iotagraphql.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	t.Run(
		"struct tag", func(t *testing.T) {
			structTag, err := iotago.StructTagFromString("0x2::coin::Coin<0x2::iota::IOTA>")
			require.NoError(t, err)
			query := iotajsonrpc.IotaObjectResponseQuery{
				Filter: &iotajsonrpc.IotaObjectDataFilter{
					StructType: structTag,
				},
				Options: &iotajsonrpc.IotaObjectDataOptions{
					ShowType:    true,
					ShowContent: true,
				},
			}
			limit := int(10)
			objs, err := client.GetOwnedObjects(
				ctx, iotaclient.GetOwnedObjectsRequest{
					Address: owner,
					Query:   &query,
					Limit:   &limit,
				},
			)

			require.NoError(t, err)
			require.NotEmpty(t, objs.Data)
		},
	)

	t.Run(
		"move module", func(t *testing.T) {
			query := iotajsonrpc.IotaObjectResponseQuery{
				Filter: &iotajsonrpc.IotaObjectDataFilter{
					AddressOwner: owner,
				},
				Options: &iotajsonrpc.IotaObjectDataOptions{
					ShowType:    true,
					ShowContent: true,
				},
			}
			limit := int(9)
			objs, err := client.GetOwnedObjects(
				ctx, iotaclient.GetOwnedObjectsRequest{
					Address: owner,
					Query:   &query,
					Limit:   &limit,
				},
			)
			require.NoError(t, err)
			require.NotEmpty(t, objs.Data)
		},
	)
}
