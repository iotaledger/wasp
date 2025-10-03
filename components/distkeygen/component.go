// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package distKeyGen implements Distributed Key Generation functionality.
package distkeygen

import (
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"

	"github.com/iotaledger/wasp/v2/packages/distkeygen"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

func init() {
	Component = &app.Component{
		Name:    "DistKeyGeneration",
		Provide: provide,
	}
}

var Component *app.Component

func provide(c *dig.Container) error {
	type nodeDeps struct {
		dig.In

		NodeIdentityProvider        registry.NodeIdentityProvider
		DistKeyPartRegistryProvider registry.DistKeyPartRegistryProvider
		NetworkProvider             peering.NetworkProvider `name:"networkProvider"`
	}

	type nodeResult struct {
		dig.Out

		Node *distkeygen.Node
	}

	if err := c.Provide(func(deps nodeDeps) nodeResult {
		node, err := distkeygen.NewNode(
			deps.NodeIdentityProvider.NodeIdentity(),
			deps.NetworkProvider,
			deps.DistKeyPartRegistryProvider,
			Component.Logger,
		)
		if err != nil {
			Component.LogPanic("failed to initialize the DistKeyGeneration node: %w", err)
		}

		return nodeResult{
			Node: node,
		}
	}); err != nil {
		Component.LogPanic(err.Error())
	}

	return nil
}
