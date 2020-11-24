package registry

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/hive.go/logger"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/key"
)

var (
	impl *Impl // A singleton.
)

// Init initializes this package.
func Init(groupSuite kyber.Group, keySuite key.Suite, log *logger.Logger) {
	impl = New(groupSuite, keySuite, log)
}

// DefaultRegistry returns an initialized default registry.
func DefaultRegistry() *Impl {
	return impl
}

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	groupSuite kyber.Group
	keySuite   key.Suite
	log        *logger.Logger
}

// New creates new instance of the registry implementation.
func New(groupSuite kyber.Group, keySuite key.Suite, log *logger.Logger) *Impl {
	return &Impl{
		groupSuite: groupSuite,
		keySuite:   keySuite,
		log:        log,
	}
}
