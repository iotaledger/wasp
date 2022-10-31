// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "Registry",
			Params:  params,
			Provide: provide,
		},
	}
}

var CoreComponent *app.CoreComponent

func provide(c *dig.Container) error {
	type registryConfigResult struct {
		dig.Out

		RegistryConfig *registry.Config
	}

	if err := c.Provide(func() registryConfigResult {
		return registryConfigResult{
			RegistryConfig: &registry.Config{
				UseText:  ParamsRegistry.UseText,
				Filename: ParamsRegistry.FileName,
			},
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type registryDeps struct {
		dig.In

		DatabaseManager *dbmanager.DBManager
	}

	type registryResult struct {
		dig.Out

		DefaultRegistry registry.Registry
	}

	if err := c.Provide(func(deps registryDeps) registryResult {
		if ParamsRegistry.UseText {
			return registryResult{
				DefaultRegistry: registry.NewTextRegistry(CoreComponent.Logger(), ParamsRegistry.FileName),
			}
		}

		return registryResult{
			DefaultRegistry: registry.NewRegistry(CoreComponent.Logger(), deps.DatabaseManager.GetRegistryKVStore()),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}
