package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suisigner"

	serialization "github.com/iotaledger/wasp/sui-go/examples/event_pubsub/lib"
)

var testMnemonic = "ordinary cry margin host traffic bulb start zone mimic wage fossil eight diagram clay say remove add atom"

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testlogger.NewSimple(false)
	api, err := suiclient.NewWebsocket(ctx, suiconn.TestnetWebsocketEndpointURL, log)
	if err != nil {
		log.Panic(err)
	}
	sender, err := suisigner.NewSignerWithMnemonic(testMnemonic, suisigner.KeySchemeFlagDefault)
	if err != nil {
		log.Panic(err)
	}
	err = suiclient.RequestFundsFromFaucet(ctx, sender.Address(), suiconn.TestnetFaucetURL)
	if err != nil {
		log.Panic(err)
	}

	packageID, err := sui.PackageIDFromHex("")
	if err != nil {
		log.Panic(err)
	}

	log.Infof("sender: %s", sender.Address())
	publisher := serialization.NewPublisher(api, sender)
	subscriber := serialization.NewSubscriber(api)

	go func() {
		for {
			publisher.PublishEvents(ctx, packageID)
		}
	}()

	go func() {
		for {
			subscriber.SubscribeEvent(ctx, packageID)
		}
	}()

	<-done
}
