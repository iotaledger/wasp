// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	suite tcrypto.Suite
	log   *logger.Logger
}

// New creates new instance of the registry implementation.
func NewRegistry(suite tcrypto.Suite, log *logger.Logger) *Impl {
	return &Impl{
		suite: suite,
		log:   log,
	}
}
