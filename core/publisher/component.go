package publisher

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/publisher"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:     "Publisher",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Provide:  provide,
			Run:      run,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	deps          dependencies
)

type dependencies struct {
	dig.In
	Publisher *publisher.Publisher
}

func provide(c *dig.Container) error {
	type publisherResult struct {
		dig.Out

		Publisher *publisher.Publisher
	}

	if err := c.Provide(func() publisherResult {
		return publisherResult{
			Publisher: publisher.New(
				CoreComponent.Logger(),
			),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}

func run() error {
	err := CoreComponent.Daemon().BackgroundWorker(
		"Publisher",
		deps.Publisher.Run,
		daemon.PriorityPublisher,
	)
	if err != nil {
		panic(err)
	}
	return nil
}
