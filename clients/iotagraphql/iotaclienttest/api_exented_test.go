package iotaclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetDynamicFields(t *testing.T) {
	ctx := context.Background()
	client := l1starter.Instance().L1Client()

	t.Run("GetObject", func(t *testing.T) {
		obj, err := client.GetObject(ctx, iotaclient.GetObjectRequest{
			ObjectID: iotago.MustObjectIDFromHex("0x5"),
			Options: &iotagraphql.IotaObjectDataOptions{
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
			ParentObjectID: iotago.MustObjectIDFromHex("0x5"),
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
	client := l1starter.Instance().L1Client()
	owner := l1starter.ISCPackageOwner.Address()

	t.Run(
		"struct tag", func(t *testing.T) {
			structTag, err := iotago.StructTagFromString("0x2::coin::Coin<0x2::iota::IOTA>")
			require.NoError(t, err)
			query := iotagraphql.IotaObjectResponseQuery{
				Filter: &iotagraphql.IotaObjectDataFilter{
					StructType: structTag,
				},
				Options: &iotagraphql.IotaObjectDataOptions{
					ShowType:    true,
					ShowContent: true,
				},
			}
			limit := int(10)
			objs, err := client.GetOwnedObjects(
				ctx, iotagraphql.GetOwnedObjectsRequest{
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
			query := iotagraphql.IotaObjectResponseQuery{
				Filter: &iotagraphql.IotaObjectDataFilter{
					AddressOwner: owner,
				},
				Options: &iotagraphql.IotaObjectDataOptions{
					ShowType:    true,
					ShowContent: true,
				},
			}
			limit := int(9)
			objs, err := client.GetOwnedObjects(
				ctx, iotagraphql.GetOwnedObjectsRequest{
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
