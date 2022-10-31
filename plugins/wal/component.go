package wal

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/wal"
)

func init() {
	Plugin = &app.Plugin{
		Component: &app.Component{
			Name:    "WAL",
			Params:  params,
			Provide: provide,
		},
		IsEnabled: func() bool {
			return ParamsWAL.Enabled
		},
	}
}

var Plugin *app.Plugin

func provide(c *dig.Container) error {
	type walResult struct {
		dig.Out

		WAL *wal.WAL
	}

	if err := c.Provide(func() walResult {
		return walResult{
			WAL: wal.New(Plugin.Logger(), ParamsWAL.Directory),
		}
	}); err != nil {
		Plugin.LogPanic(err)
	}

	return nil
}
