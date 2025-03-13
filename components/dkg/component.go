// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"

	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
)

func init() {
	Component = &app.Component{
		Name:    "DKG",
		Provide: provide,
	}
}

var Component *app.Component

func provide(c *dig.Container) error {
	type nodeDeps struct {
		dig.In

		NodeIdentityProvider    registry.NodeIdentityProvider
		DKShareRegistryProvider registry.DKShareRegistryProvider
		NetworkProvider         peering.NetworkProvider `name:"networkProvider"`
	}

	type nodeResult struct {
		dig.Out

		Node *dkg.Node
	}

	if err := c.Provide(func(deps nodeDeps) nodeResult {
		node, err := dkg.NewNode(
			deps.NodeIdentityProvider.NodeIdentity(),
			deps.NetworkProvider,
			deps.DKShareRegistryProvider,
			Component.Logger,
		)
		if err != nil {
			Component.LogPanic("failed to initialize the DKG node: %w", err)
		}

		return nodeResult{
			Node: node,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}
