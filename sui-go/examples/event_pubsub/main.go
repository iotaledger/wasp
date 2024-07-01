package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suisigner"

	serialization "github.com/iotaledger/wasp/sui-go/examples/event_pubsub/lib"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	api := suiclient.New(suiconn.TestnetEndpointURL)
	sender, err := suisigner.NewSignerWithMnemonic(suisigner.TestMnemonic, suisigner.KeySchemeFlagDefault)
	if err != nil {
		log.Panic(err)
	}
	err = suiclient.RequestFundsFromFaucet(sender.Address(), suiconn.TestnetFaucetURL)
	if err != nil {
		log.Panic(err)
	}

	packageID, err := sui.PackageIDFromHex("")
	if err != nil {
		log.Panic(err)
	}

	log.Println("sender: ", sender.Address())
	publisher := serialization.NewPublisher(api, sender)
	subscriber := serialization.NewSubscriber(api)

	go func() {
		for {
			publisher.PublishEvents(context.Background(), packageID)
		}
	}()

	go func() {
		for {
			subscriber.SubscribeEvent(context.Background(), packageID)
		}
	}()

	<-done
}
