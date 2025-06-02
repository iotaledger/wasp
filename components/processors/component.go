// Package processors provides data processing functionality for the application.
package processors

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"

	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

func init() {
	Component = &app.Component{
		Name:    "Processors",
		Provide: provide,
	}
}

var Component *app.Component

func provide(c *dig.Container) error {
	type processorsConfigResult struct {
		dig.Out

		ProcessorsConfig *processors.Config
	}

	if err := c.Provide(func() processorsConfigResult {
		return processorsConfigResult{
			ProcessorsConfig: coreprocessors.NewConfig(),
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}
