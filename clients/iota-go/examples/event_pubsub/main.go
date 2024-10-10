package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	serialization "github.com/iotaledger/wasp/clients/iota-go/examples/event_pubsub/lib"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

var testMnemonic = "ordinary cry margin host traffic bulb start zone mimic wage fossil eight diagram clay say remove add atom"

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := testlogger.NewSimple(false)
	api, err := iotaclient.NewWebsocket(ctx, iotaconn.AlphanetWebsocketEndpointURL, log)
	if err != nil {
		log.Panic(err)
	}
	sender, err := iotasigner.NewSignerWithMnemonic(testMnemonic, iotasigner.KeySchemeFlagDefault)
	if err != nil {
		log.Panic(err)
	}
	err = iotaclient.RequestFundsFromFaucet(ctx, sender.Address(), iotaconn.AlphanetFaucetURL)
	if err != nil {
		log.Panic(err)
	}

	packageID, err := iotago.PackageIDFromHex("")
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
