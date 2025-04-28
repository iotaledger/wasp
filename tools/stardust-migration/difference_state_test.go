package main

import (
	"fmt"
	_ "net/http/pprof"
	"testing"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/cryptolib"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/migrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

func TestDifferences(t *testing.T) {
	srcDbPath := "/home/luke/dev/wasp_stardust_mainnet/chains/data/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"
	indexFilePath := "/home/luke/dev/wasp_stardust_mainnet/trie_db.bcs"
	srcStore, _ := initSetup(srcDbPath, indexFilePath)

	destKVDB := mapdb.NewMapDB()
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVDB))

	oldState := lo.Must(srcStore.StateByIndex(1))
	oldBlockLog := oldstate.GetContactStateReader(oldState, old_blocklog.Contract.Hname())

	fmt.Println("OLD STATE")
	oldBlockLog.Iterate("", func(key old_kv.Key, value []byte) bool {
		fmt.Println(key)
		return true
	})

	destStateDraft := destStore.NewOriginStateDraft()
	migrations.MigrateBlocklogContract(oldState, destStateDraft, old_isc.ChainID{}, &transaction.StateMetadata{}, cryptolib.NewRandomAddress(), 10000, false)
	newContractState := newstate.GetContactState(destStateDraft, blocklog.Contract.Hname())

	fmt.Println("NEW STATE")
	newContractState.Iterate("", func(key kv.Key, value []byte) bool {
		fmt.Println(key)
		return true
	})
}
