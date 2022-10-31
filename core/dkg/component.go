// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import (
	"go.uber.org/dig"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "DKG",
			Provide: provide,
		},
	}
}

var CoreComponent *app.CoreComponent

func provide(c *dig.Container) error {
	type nodeDeps struct {
		dig.In

		DefaultRegistry        registry.Registry
		DefaultNetworkProvider peering.NetworkProvider `name:"defaultNetworkProvider"`
	}

	type nodeResult struct {
		dig.Out

		DefaultNode *dkg.Node `name:"defaultNode"`
	}

	if err := c.Provide(func(deps nodeDeps) nodeResult {
		defaultNode, err := dkg.NewNode(
			deps.DefaultRegistry.GetNodeIdentity(),
			deps.DefaultNetworkProvider,
			deps.DefaultRegistry,
			CoreComponent.Logger().Desugar().WithOptions(zap.IncreaseLevel(logger.LevelWarn)).Sugar(),
		)
		if err != nil {
			CoreComponent.LogPanic(xerrors.Errorf("failed to initialize the DKG node: %w", err))
		}

		return nodeResult{
			DefaultNode: defaultNode,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}
