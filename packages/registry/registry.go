package registry

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dks"
)

var (
	impl *Impl // A singleton.
)

// Init initializes this package.
func Init(suite dks.Suite, log *logger.Logger) {
	impl = New(suite, log)
}

// DefaultRegistry returns an initialized default registry.
func DefaultRegistry() *Impl {
	return impl
}

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	suite dks.Suite
	log   *logger.Logger
}

// New creates new instance of the registry implementation.
func New(suite dks.Suite, log *logger.Logger) *Impl {
	return &Impl{
		suite: suite,
		log:   log,
	}
}
