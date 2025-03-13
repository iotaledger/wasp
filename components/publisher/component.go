package publisher

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/daemon"
	"github.com/iotaledger/wasp/packages/publisher"
)

func init() {
	Component = &app.Component{
		Name:     "Publisher",
		DepsFunc: func(cDeps dependencies) { deps = cDeps },
		Provide:  provide,
		Run:      run,
	}
}

var (
	Component *app.Component
	deps      dependencies
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
				Component.Logger
			),
		}
	}); err != nil {
		Component.LogPanic(err)
	}

	return nil
}

func run() error {
	err := Component.Daemon().BackgroundWorker(
		"Publisher",
		deps.Publisher.Run,
		daemon.PriorityPublisher,
	)
	if err != nil {
		panic(err)
	}
	return nil
}
