package inspection

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initRequestsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "requests <AnchorID>",
		Short: "Show the owned requests of an Anchor",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			objectID, err := iotago.ObjectIDFromHex(args[0])
			log.Check(err)

			ctx := context.Background()

			obj, err := cliclients.L1Client().GetObject(ctx, iotaclient.GetObjectRequest{
				ObjectID: objectID,
				Options: &iotajsonrpc.IotaObjectDataOptions{
					ShowType: true,
				},
			})
			log.Check(err)

			if obj.Data.Type == nil {
				log.Fatalf("Failed to get Anchor type")
			}

			resource, err := iotago.NewResourceType(*obj.Data.Type)
			log.Check(err)

			packageID, err := iotago.PackageIDFromHex(resource.Address.ToHex())
			log.Check(err)

			if packageID == nil {
				log.Fatalf("Failed to get Anchors PackageID")
			}

			iscMoveClient := iscmoveclient.NewClient(cliclients.L1Client().IotaClient(), "")

			requests := make([]*iscmove.RefWithObject[iscmove.Request], 0)
			err = iscMoveClient.GetRequestsSorted(ctx, *packageID, objectID, 9999, func(err error, request *iscmove.RefWithObject[iscmove.Request]) {
				requests = append(requests, request)
			})
			log.Check(err)

			for _, request := range requests {
				fmt.Printf("Request: %s\n", request.ObjectID.String())
			}
		},
	}
}
