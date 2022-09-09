// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
)

func TestBasic(t *testing.T) {
	logger := testlogger.NewLogger(t)
	getOrCreateKVStore := func(chain *isc.ChainID) kvstore.KVStore {
		return mapdb.NewMapDB()
	}

	_ = New(logger, coreprocessors.Config(), 10, time.Second, false, nil, getOrCreateKVStore)
}
