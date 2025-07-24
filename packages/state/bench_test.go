package state_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

// run with: go test -tags rocksdb -benchmem -cpu=1 -run=' ' -bench='Bench.*' -benchtime 100x
//
// To generate mem and cpu profiles, add -cpuprofile=cpu.out -memprofile=mem.out
// Then: go tool pprof -http :8080 {cpu,mem}.out
func BenchmarkTriePruning(b *testing.B) {
	b.StopTimer()
	path := "/tmp/" + b.Name() + ".db"
	const cacheSize = database.CacheSizeDefault
	db, err := database.NewDatabase(hivedb.EngineRocksDB, path, true, cacheSize)
	require.NoError(b, err)
	b.Cleanup(func() {
		os.RemoveAll(path)
	})

	kvs := db.KVStore()
	b.Cleanup(func() {
		kvs.Close()
	})
	r := newRandomStateWithDB(b, kvs)
	trieRoots := make([]trie.Hash, 0)
	for i := 1; i <= b.N; i++ {
		b, _, _ := r.commitNewBlock(r.cs.LatestBlock(), time.Unix(int64(i), 0))
		trieRoots = append(trieRoots, b.TrieRoot())
	}
	kvs.Flush()
	rand.Shuffle(len(trieRoots), func(i, j int) {
		trieRoots[i], trieRoots[j] = trieRoots[j], trieRoots[i]
	})
	b.StartTimer()
	deletedNodes := uint(0)
	deletedValues := uint(0)
	for _, trieRoot := range trieRoots {
		stats, err := r.cs.Prune(trieRoot)
		require.NoError(b, err)
		deletedNodes += stats.DeletedNodes
		deletedValues += stats.DeletedValues
	}
	b.StopTimer()
	b.ReportMetric(float64(deletedNodes)/float64(b.N), "deleted-nodes/op")
	b.ReportMetric(float64(deletedValues)/float64(b.N), "deleted-values/op")
}
