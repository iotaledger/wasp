package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/dgravesa/go-parallel/parallel"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/pbnjay/memory"
	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/kvstore/mapdb"

	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/tools/stardust-migration/migrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
)

var calledContracts sync.Map

func addContractCall(contract old_isc.Hname, entryPoint old_isc.Hname) {
	// Get or create inner map
	innerMap, _ := calledContracts.LoadOrStore(contract, &sync.Map{})

	// Increment counter in inner map
	m := innerMap.(*sync.Map)
	count, _ := m.LoadOrStore(migrations.ContractNameFuncs[entryPoint], 0)
	m.Store(migrations.ContractNameFuncs[entryPoint], count.(int)+1)
}

func countCalls(oldChainState old_kv.KVStoreReader) {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_blocklog.Contract.Hname())
	oldRequests := old_collections.NewMapReadOnly(oldContractState, old_blocklog.PrefixRequestReceipts)

	oldRequests.Iterate(func(elemKey []byte, value []byte) bool {
		// TODO: Validate if this is fine. BlockIndex and ReqIndex is 0 here, as we don't persist these values in the db
		// So in my understanding, using 0 here is fine. If not, we need to iterate the whole request lut again and combine the tables.
		// I added a solution in commit: 96504e6165ed4056a3e8a50281215f3d7eb7c015, for now I go without.
		oldReceipt, err := old_blocklog.RequestReceiptFromBytes(value, 0, 0)
		if err != nil {
			panic(fmt.Errorf("requestReceipt migration error: %v", err))
		}
		addContractCall(oldReceipt.Request.CallTarget().Contract, oldReceipt.Request.CallTarget().EntryPoint)
		return true
	})
}

func PrintCalledContracts() {
	fmt.Println("Contract calls:")
	calledContracts.Range(func(key, value any) bool {
		contract := key.(old_isc.Hname)
		innerMap := value.(*sync.Map)

		fmt.Printf("\nContract %v:\n", contract)
		innerMap.Range(func(name, count any) bool {
			fmt.Printf("  %s: %d calls\n", name.(string), count.(int))
			return true
		})
		return true
	})
}

func initSetup(srcChainDBDir string, indexFilePath string) (old_indexedstore.IndexedStore, []old_trie.Hash) {
	srcKVS := db.ConnectOld(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))

	migrations.BuildContractNameFuncs()
	/*
		trieRoots, indexFileFound := blockindex.ReadIndexFromFile(indexFilePath)
		if !indexFileFound {
			panic("could not find index file")
		}
	*/
	return srcStore, nil
}

func dumpCoreContractCalls(srcChainDBDir string, indexFilePath string) {
	srcStore, trieRoots := initSetup(srcChainDBDir, indexFilePath)

	parallel.WithStrategy(parallel.StrategyFetchNextIndex).WithNumGoroutines(10).For(len(trieRoots), func(i int, _ int) {
		trieRootForBlock := trieRoots[uint32(i)]

		k, err := srcStore.BlockByTrieRoot(trieRootForBlock)
		if err != nil {
			panic(err)
		}

		reader := k.MutationsReader()

		countCalls(reader)

		if i > 0 && i%1000 == 0 {
			fmt.Print("")
			fmt.Print("")
			fmt.Print("")
			fmt.Printf("\n\n\nINDEX: %d\n\n", i)

			PrintCalledContracts()
		}
	})
	fmt.Println("DONE")
	PrintCalledContracts()
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocatedMB := float64(m.Alloc) / 1024 / 1024
	systemMB := float64(m.Sys) / 1024 / 1024

	fmt.Printf("Memory usage: %.2fMB (Allocated) %.2fMB (System)\n", allocatedMB, systemMB)
}

func TestMigrateBlocklog(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	srcDbPath := "/home/luke/dev/wasp_stardust_mainnet/chains/data/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"
	indexFilePath := "/home/luke/dev/wasp_stardust_mainnet/trie_db.bcs"
	srcStore, trieRoots := initSetup(srcDbPath, indexFilePath)

	destKVDB := mapdb.NewMapDB()
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVDB))

	// Coming from packages/origin/origin.go InitChain()
	originStateDraft := destStore.NewOriginStateDraft()
	originStateDraft.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.Encode(uint32(0)))
	originStateDraft.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.Encode(time.Unix(0, 0)))

	block := destStore.Commit(originStateDraft)
	destStore.SetLatest(block.TrieRoot())

	now := time.Now()

	//oldChainID := lo.Must(old_isc.ChainIDFromString("tgl1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm25gu3sa"))
	//newChainID := isctest.RandomChainID()
	for i := 0; i < len(trieRoots); i++ {
		destStateDraft := lo.Must(destStore.NewStateDraft(time.Now(), block.L1Commitment()))

		trieRootForBlock := trieRoots[uint32(i)]

		srcBlock, err := srcStore.BlockByTrieRoot(trieRootForBlock)
		if err != nil {
			panic(err)
		}

		/*
			migrations.MigrateBlocklogContract(srcBlock.MutationsReader(), destStateDraft, oldChainID, newChainID, &transaction.StateMetadata{
				L1Commitment: block.L1Commitment(),
			}, nil)
		*/
		// Handle deletions
		// Here its easy, because the key remains the same for both databases
		// For the whole migration we probably should create some OldToNewKey function handler to solve the differences.
		// I think most of all keys are just 1:1 mappings.
		for del, _ := range srcBlock.Mutations().Dels {
			destStateDraft.Del(kv.Key(del[:]))
		}

		block = destStore.Commit(destStateDraft)
		destStore.SetLatest(block.TrieRoot())

		if i > 0 && i%1000 == 0 {
			fmt.Printf("\n\n\nINDEX: %d\n\n", i)

			fmt.Printf("Time start: %s, time now: %s\n", now.String(), time.Now().String())

			printMemUsage()

			freeMemory := memory.FreeMemory()
			totalMemory := memory.TotalMemory()

			fmt.Printf("Free memory: %d, total memory: %d\n", freeMemory, totalMemory)
		}

	}

	fmt.Printf("Time start: %s, time now: %s\nDONE!", now.String(), time.Now().String())
}

func TestDumpContractCalls(t *testing.T) {
	srcDbPath := "/home/luke/dev/wasp_stardust_mainnet/chains/data/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"
	indexFilePath := "/home/luke/dev/wasp_stardust_mainnet/trie_db.bcs"

	dumpCoreContractCalls(srcDbPath, indexFilePath)
}
