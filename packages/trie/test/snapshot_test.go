package test

import (
	"io"
	"math/rand/v2"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/trie"
)

type storeWithReadStats struct {
	trie.KVReader
	stats struct {
		Reads    uint
		KeysRead uint
	}
}

var _ trie.KVReader = &storeWithReadStats{}

func (store *storeWithReadStats) Get(key []byte) []byte {
	store.stats.Reads++
	store.stats.KeysRead++
	return store.KVReader.Get(key)
}

func (store *storeWithReadStats) MultiGet(keys [][]byte) [][]byte {
	store.stats.Reads++
	store.stats.KeysRead += uint(len(keys))
	return store.KVReader.MultiGet(keys)
}

func (store *storeWithReadStats) Has(key []byte) bool {
	store.stats.Reads++
	store.stats.KeysRead++
	return store.KVReader.Has(key)
}

func BenchmarkTakeSnapshot(b *testing.B) {
	// before MultiGet:
	//  10739041 ns/op   6540 kvs/op    6540 reads/op    8354404 B/op   194195 allocs/op
	// after MultiGet:
	//  12394280 ns/op   6540 kvs/op    3720 reads/op   11556460 B/op   214444 allocs/op
	// reads/op measures the amount of times the DB is called to fetch data

	store := NewInMemoryKVStore()
	root := trie.MustInitRoot(store)

	// compose trie
	{
		// small alphabet to force keys to have common prefixes
		const alphabet = "abc"
		const maxKeyLen = 10
		rnd := rand.NewChaCha8([32]byte{1, 2, 3, 4})
		makeKey := func() []byte {
			n := int(rnd.Uint64()%uint64(maxKeyLen)) + 1
			key := make([]byte, n)
			for i := range key {
				j := int(rnd.Uint64() % uint64(len(alphabet)))
				key[i] = alphabet[j]
			}
			return key
		}
		values := NewScrambledZipfian(1000)

		tr := lo.Must(trie.NewTrieUpdatable(store, root))

		for range 10000 {
			key := makeKey()
			value := values.Next()
			tr.Update([]byte(key), []byte(value))
		}

		root, _ = tr.Commit(store)
	}

	storeWithReadStats := &storeWithReadStats{
		KVReader: store,
	}
	r := lo.Must(trie.NewTrieReader(storeWithReadStats, root))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := r.TakeSnapshot(io.Discard)
		require.NoError(b, err)
	}

	b.ReportMetric(float64(storeWithReadStats.stats.Reads)/float64(b.N), "reads/op")
	b.ReportMetric(float64(storeWithReadStats.stats.KeysRead)/float64(b.N), "kvs/op")
}
