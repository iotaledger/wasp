package inspection

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
)

func initRequestsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "requests <AnchorID>",
		Short: "Show the owned requests of an Anchor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			objectID, err := iotago.ObjectIDFromHex(args[0])
			if err != nil {
				return err
			}

			ctx := context.Background()

			obj, err := cliclients.L1Client().GetObject(ctx, iotagraphql.GetObjectRequest{
				ObjectID: objectID,
				Options: &iotagraphql.IotaObjectDataOptions{
					ShowType: true,
				},
			})
			if err != nil {
				return err
			}

			if obj.Data.Type == nil {
				return fmt.Errorf("failed to get Anchor type")
			}

			resource, err := iotago.NewResourceType(*obj.Data.Type)
			if err != nil {
				return err
			}

			packageID, err := iotago.PackageIDFromHex(resource.Address.ToHex())
			if err != nil {
				return err
			}

			if packageID == nil {
				return fmt.Errorf("failed to get Anchors PackageID")
			}

			iscMoveClient := iscmoveclient.NewClient(cliclients.L1Client().GetIotaClient().(*iotaclient.Client), "")

			requests := make([]*iscmove.RefWithObject[iscmove.Request], 0)
			err = iscMoveClient.GetRequestsSorted(ctx, *packageID, objectID, 9999, func(err error, request *iscmove.RefWithObject[iscmove.Request]) {
				requests = append(requests, request)
			})
			if err != nil {
				return err
			}

			for _, request := range requests {
				fmt.Printf("Request: %s\n", request.ObjectID.String())
			}
			return nil
		},
	}
}
