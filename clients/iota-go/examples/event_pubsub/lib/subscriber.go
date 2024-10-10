package serialization

import (
	"context"
	"fmt"
	"log"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
)

type Subscriber struct {
	client *iotaclient.Client
	// *account.Account
}

func NewSubscriber(client *iotaclient.Client) *Subscriber {
	return &Subscriber{client: client}
}

func (s *Subscriber) SubscribeEvent(ctx context.Context, packageID *iotago.PackageID) {
	resultCh := make(chan *iotajsonrpc.SuiEvent)
	err := s.client.SubscribeEvent(context.Background(), &iotajsonrpc.EventFilter{Package: packageID}, resultCh)
	if err != nil {
		log.Fatal(err)
	}

	for result := range resultCh {
		fmt.Println("result: ", result)
	}
}
