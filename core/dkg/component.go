// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/core/peering"
	"github.com/iotaledger/wasp/core/registry"
	dkg_pkg "github.com/iotaledger/wasp/packages/dkg"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:      "DKG",
			Configure: configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
	defaultNode   *dkg_pkg.Node
)

func configure() error {
	reg := registry.DefaultRegistry()
	peeringProvider := peering.DefaultNetworkProvider()
	var err error
	defaultNode, err = dkg_pkg.NewNode(
		reg.GetNodeIdentity(),
		peeringProvider,
		reg,
		CoreComponent.Logger().Desugar().WithOptions(zap.IncreaseLevel(logger.LevelWarn)).Sugar(),
	)
	if err != nil {
		panic(xerrors.Errorf("failed to initialize the DKG node: %w", err))
	}

	return err
}

// DefaultNode returns the default instance of the DKG Node Provider.
// It should be used to access all the DKG Node functions (not the DKG Initiator's).
func DefaultNode() *dkg_pkg.Node {
	return defaultNode
}
