package iotaclienttest

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetDynamicFields(t *testing.T) {
	ctx := context.Background()
	client := l1starter.Instance().L1Client()

	t.Run("GetObject", func(t *testing.T) {
		obj, err := client.GetObject(ctx, *iotago.MustObjectIDFromHex("0x5"))
		require.NoError(t, err)
		require.NotNil(t, obj)

		t.Logf("Object exists: %+v", obj)
	})

	t.Run("GetDynamicFields", func(t *testing.T) {
		resp, err := client.GetDynamicFields(ctx, iotagraphql.GetDynamicFieldsRequest{
			ParentObjectID: *iotago.MustObjectIDFromHex("0x5"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		t.Logf("Dynamic Fields Response: %+v", resp)
		nodes := resp.Owner.DynamicFields.Nodes
		t.Logf("Number of dynamic fields: %d", len(nodes))
		t.Logf("Has next page: %v", resp.Owner.DynamicFields.PageInfo.HasNextPage)

		// Verify we got dynamic fields
		require.NotEmpty(t, nodes, "Expected to find dynamic fields on this object")

		// Log details about the first few dynamic fields
		for i, field := range nodes {
			if i >= 3 {
				break
			}
			t.Logf("Dynamic field %d: Name=%+v, Value=%+v",
				i, field.Name, field.Value)
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
			structTag := "0x2::coin::Coin<0x2::iota::IOTA>"
			limit := int(10)
			objs, err := client.GetOwnedObjects(
				ctx, iotagraphql.GetOwnedObjectsRequest{
					Address: owner,
					Filter:  &graphqltypes.ObjectFilter{Type: lo.ToPtr(structTag)},
					Limit:   &limit,
				},
			)

			require.NoError(t, err)
			require.NotEmpty(t, objs.Address.Objects.Nodes)
		},
	)

	t.Run(
		"address owner", func(t *testing.T) {
			limit := int(9)
			objs, err := client.GetOwnedObjects(
				ctx, iotagraphql.GetOwnedObjectsRequest{
					Address: owner,
					Filter:  &graphqltypes.ObjectFilter{Owner: &owner},
					Limit:   &limit,
				},
			)
			require.NoError(t, err)
			require.NotEmpty(t, objs.Address.Objects.Nodes)
		},
	)
}
