// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	chainStateStoreProvider := func(chain isc.ChainID) (kvstore.KVStore, error) {
		return mapdb.NewMapDB(), nil
	}

	_ = New(logger, nil, coreprocessors.NewConfigWithCoreContracts(), 10, time.Second, false, nil, chainStateStoreProvider, false, "")
}
