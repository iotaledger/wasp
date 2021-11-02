// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/dbtest"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func TestMemoryDB(t *testing.T) {
	dbtest.TestDatabaseSuite(t, func() ethdb.KeyValueStore {
		return NewKVAdapter(dict.New())
	})
}
