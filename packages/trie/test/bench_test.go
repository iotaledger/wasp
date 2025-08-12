package test

import (
	"bytes"
	"io"
	"maps"
	"math/rand/v2"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/trie"
)

func keyMaker() func() []byte {
	// small alphabet to force keys to have common prefixes
	const alphabet = "abc"
	const maxKeyLen = 10
	rnd := rand.NewChaCha8([32]byte{1, 2, 3, 4})
	return func() []byte {
		n := int(rnd.Uint64()%uint64(maxKeyLen)) + 1
		key := make([]byte, n)
		for i := range key {
			j := int(rnd.Uint64() % uint64(len(alphabet)))
			key[i] = alphabet[j]
		}
		return key
	}
}

func makeTrie(n int) (*InMemoryKVStore, []trie.Hash) {
	store := NewInMemoryKVStore()
	roots := []trie.Hash{lo.Must(trie.InitRoot(store, true))}
	makeKey := keyMaker()
	values := NewScrambledZipfian(1000, 0)

	for range n {
		tr := lo.Must(trie.NewTrieUpdatable(store, roots[len(roots)-1]))
		for range 10000 {
			key := makeKey()
			value := values.Next()
			tr.Update(key, []byte(value))
		}
		root, _, _ := tr.Commit(store)
		roots = append(roots, root)
	}
	return store, roots
}

func BenchmarkTakeSnapshot(b *testing.B) {
	// before MultiGet:
	//  10739041 ns/op   6540 kvs/op    6540 reads/op    8354404 B/op   194195 allocs/op
	// after MultiGet:
	//  12394280 ns/op   6540 kvs/op    3720 reads/op   11556460 B/op   214444 allocs/op
	// reads/op measures the amount of times the DB is called to fetch data (which is the bottleneck when using RocksDB)

	store, roots := makeTrie(1)
	r := lo.Must(trie.NewTrieReader(store, roots[len(roots)-1]))
	store.ResetStats()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := r.TakeSnapshot(io.Discard)
		require.NoError(b, err)
	}
	b.ReportMetric(float64(store.Stats.Get+store.Stats.MultiGet)/float64(b.N), "reads/op")
	b.ReportMetric(float64(store.Stats.Get+store.Stats.MultiGetKeys)/float64(b.N), "kvs/op")
}

func BenchmarkRestoreSnapshot(b *testing.B) {
	// before MultiGet:
	//   44839973 ns/op    13083 kvs/op    10263 reads/op    22344852 B/op    389676 allocs/op
	// after MultiGet:
	//   56646487 ns/op    13083 kvs/op     7443 reads/op    27590744 B/op    386339 allocs/op
	// reads/op measures the amount of times the DB is called to fetch data (which is the bottleneck when using RocksDB)

	b.StopTimer()
	buf := bytes.NewBuffer(nil)
	{
		store, roots := makeTrie(1)
		r := lo.Must(trie.NewTrieReader(store, roots[len(roots)-1]))
		err := r.TakeSnapshot(buf)
		require.NoError(b, err)
	}

	stats := InMemoryKVStoreStats{}
	for i := 0; i < b.N; i++ {
		newStore := NewInMemoryKVStore()
		newStore.Stats = stats
		b.StartTimer()
		trie.RestoreSnapshot(bytes.NewReader(buf.Bytes()), newStore, true)
		b.StopTimer()
		stats = newStore.Stats
	}

	b.ReportMetric(float64(stats.Get+stats.MultiGet)/float64(b.N), "reads/op")
	b.ReportMetric(float64(stats.Get+stats.MultiGetKeys)/float64(b.N), "kvs/op")
}

func BenchmarkPrune(b *testing.B) {
	// before MultiGet:
	//   22620182 ns/op    16476 kvs/op    12804 reads/op    15628541 B/op    314546 allocs/op
	// after MultiGet:
	//   24594662 ns/op    16476 kvs/op     9132 reads/op    19913272 B/op    307635 allocs/op
	// reads/op measures the amount of times the DB is called to fetch data (which is the bottleneck when using RocksDB)

	b.StopTimer()
	store, roots := makeTrie(3)
	penultimateRoot := roots[len(roots)-2]
	stats := InMemoryKVStoreStats{}
	for i := 0; i < b.N; i++ {
		storeClone := NewInMemoryKVStore()
		storeClone.m = maps.Clone(store.m)
		storeClone.Stats = stats
		b.StartTimer()
		_, err := trie.Prune(storeClone, penultimateRoot)
		b.StopTimer()
		require.NoError(b, err)
		stats = storeClone.Stats
	}
	b.ReportMetric(float64(stats.Get+stats.MultiGet)/float64(b.N), "reads/op")
	b.ReportMetric(float64(stats.Get+stats.MultiGetKeys)/float64(b.N), "kvs/op")
}

func BenchmarkCommit(b *testing.B) {
	// before MultiGet:
	//    7052109 ns/op   4386 kvs/op   4386 reads/op   4833882 B/op   100044 allocs/op
	// after MultiGet:
	//    9486585 ns/op   4221 kvs/op      2 reads/op   6279097 B/op   108977 allocs/op
	// reads/op measures the amount of times the DB is called to fetch data (which is the bottleneck when using RocksDB)

	b.StopTimer()

	store := NewInMemoryKVStore()
	stats := store.Stats

	makeKey := keyMaker()
	values := NewScrambledZipfian(1000, 0)

	root := lo.Must(trie.InitRoot(store, true))
	for i := 0; i < b.N; i++ {
		tr := lo.Must(trie.NewTrieUpdatable(store, root))
		for range 1000 {
			key := makeKey()
			value := values.Next()
			tr.Update(key, []byte(value))
		}
		store.Stats = stats
		b.StartTimer()
		root, _, _ = tr.Commit(store)
		b.StopTimer()
		stats = store.Stats
	}
	b.ReportMetric(float64(stats.Get+stats.MultiGet)/float64(b.N), "reads/op")
	b.ReportMetric(float64(stats.Get+stats.MultiGetKeys)/float64(b.N), "kvs/op")
}
