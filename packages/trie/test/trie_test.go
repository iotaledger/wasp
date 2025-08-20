package test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/Yiling-J/theine-go"
	"github.com/dgraph-io/ristretto"
	"github.com/dgryski/go-clockpro"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pingcap/go-ycsb/pkg/generator"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/util"
)

func TestBasic(t *testing.T) {
	store := NewInMemoryKVStore()

	var roots []trie.Hash
	{
		roots = append(roots, lo.Must(trie.InitRoot(store, true)))
	}
	{
		root0 := roots[0]
		state, err := trie.NewTrieReader(store, root0)
		require.NoError(t, err)
		require.EqualValues(t, []byte(nil), state.Get([]byte("a")))
	}

	fmt.Printf("--- DebugDump %d\n", len(roots))
	trie.DebugDump(store, roots, os.Stdout)

	{
		tr, err := trie.NewTrieUpdatable(store, roots[0])
		require.NoError(t, err)
		tr.Update([]byte("a"), []byte("a"))
		tr.Update([]byte("b"), []byte("b"))
		root1, _, stats := tr.Commit(store)
		roots = append(roots, root1)
		// the trie at root1 has 4 nodes (see the output of DebugDump below)
		require.EqualValues(t, 4, stats.CreatedNodes)
		require.EqualValues(t, 0, stats.CreatedValues)
	}

	fmt.Printf("--- DebugDump %d\n", len(roots))
	trie.DebugDump(store, roots, os.Stdout)

	{
		tr, err := trie.NewTrieUpdatable(store, roots[1])
		require.NoError(t, err)
		tr.Update([]byte("b"), []byte("bb"))
		root2, _, _ := tr.Commit(store)
		roots = append(roots, root2)
		require.NoError(t, err)

		require.EqualValues(t, []byte("a"), tr.Get([]byte("a")))
		require.EqualValues(t, []byte("bb"), tr.Get([]byte("b")))
	}

	fmt.Printf("--- DebugDump %d\n", len(roots))
	trie.DebugDump(store, roots, os.Stdout)

	{
		tr, err := trie.NewTrieUpdatable(store, roots[2])
		require.NoError(t, err)
		tr.Update([]byte("a"), nil)
		tr.Update([]byte("cccddd"), []byte("c"))
		tr.Update([]byte("ccceee"), bytes.Repeat([]byte("c"), 70))
		root3, _, _ := tr.Commit(store)
		roots = append(roots, root3)
		require.NoError(t, err)

		require.Nil(t, tr.Get([]byte("a")))
	}

	fmt.Printf("--- DebugDump %d\n", len(roots))
	trie.DebugDump(store, roots, os.Stdout)

	state, err := trie.NewTrieReader(store, roots[3])
	require.NoError(t, err)
	require.Nil(t, state.Get([]byte("a")))

	{
		tr, err := trie.NewTrieUpdatable(store, roots[3])
		require.NoError(t, err)
		tr.Update([]byte("b"), nil)
		tr.Update([]byte("cccddd"), nil)
		tr.Update([]byte("ccceee"), nil)
		root4, _, _ := tr.Commit(store)
		roots = append(roots, root4)
		require.NoError(t, err)

		require.Nil(t, tr.Get([]byte("a")))
		require.Nil(t, tr.Get([]byte("b")))
		require.Nil(t, tr.Get([]byte("cccddd")))
		require.Nil(t, tr.Get([]byte("ccceee")))
	}
	// trie is now empty, so hash4 should be equal to hash0
	require.Equal(t, roots[0], roots[4])
}

func TestBasic2(t *testing.T) {
	store := NewInMemoryKVStore()

	root0 := lo.Must(trie.InitRoot(store, true))

	var root1 trie.Hash
	{
		tr, err := trie.NewTrieUpdatable(store, root0)
		require.NoError(t, err)
		tr.Update([]byte{0x00}, []byte{0})
		tr.Update([]byte{0x01}, []byte{0})
		tr.Update([]byte{0x10}, []byte{0})
		root1, _, _ = tr.Commit(store)
	}

	tr, err := trie.NewTrieReader(store, root1)
	require.NoError(t, err)
	require.True(t, tr.Has([]byte{0x00}))
	require.True(t, tr.Has([]byte{0x01}))
	require.True(t, tr.Has([]byte{0x10}))

	trie.DebugDump(store, []trie.Hash{root0, root1}, io.Discard)
}

func TestBasic3(t *testing.T) {
	store := NewInMemoryKVStore()

	root0 := lo.Must(trie.InitRoot(store, true))

	var root1 trie.Hash
	{
		tr, err := trie.NewTrieUpdatable(store, root0)
		require.NoError(t, err)
		tr.Update([]byte{0x30}, []byte{1})
		tr.Update([]byte{0x31}, []byte{1})
		tr.Update([]byte{0xb0}, []byte{1})
		tr.Update([]byte{0xb2}, []byte{1})
		root1, _, _ = tr.Commit(store)
	}

	tr, err := trie.NewTrieReader(store, root1)
	require.NoError(t, err)
	require.Equal(t, []byte{1}, tr.Get([]byte{0x30}))
	require.Equal(t, []byte{1}, tr.Get([]byte{0x31}))
	trie.DebugDump(store, []trie.Hash{root0, root1}, io.Discard)
	require.Equal(t, []byte{1}, tr.Get([]byte{0xb0}))
	require.Equal(t, []byte{1}, tr.Get([]byte{0xb2}))

	trie.DebugDump(store, []trie.Hash{root0, root1}, io.Discard)
}

func TestKeyTooLong(t *testing.T) {
	store := NewInMemoryKVStore()

	tooLongKey := make([]byte, trie.KeyMaxLength+1)
	root0 := lo.Must(trie.InitRoot(store, true))
	{
		tr, err := trie.NewTrieUpdatable(store, root0)
		require.NoError(t, err)
		require.Panics(t, func() {
			tr.Update(tooLongKey, []byte{0})
		})
	}
}

func TestCreateTrie(t *testing.T) {
	t.Run("ok init-"+"", func(t *testing.T) {
		rootC1 := lo.Must(trie.InitRoot(NewInMemoryKVStore(), true))
		require.NotNil(t, rootC1)

		rootC2 := lo.Must(trie.InitRoot(NewInMemoryKVStore(), true))
		require.NotNil(t, rootC2)

		require.Equal(t, rootC1, rootC2)
	})
	t.Run("update 1"+"", func(t *testing.T) {
		store := NewInMemoryKVStore()
		const (
			key   = "key"
			value = "value"
		)

		rootInitial := lo.Must(trie.InitRoot(store, true))
		require.NotNil(t, rootInitial)

		tr, err := trie.NewTrieUpdatable(store, rootInitial)
		require.NoError(t, err)

		require.Empty(t, tr.GetStr(""))

		tr.UpdateStr(key, value)
		rootCnext, _, _ := tr.Commit(store)
		trie.DebugDump(store, []trie.Hash{rootInitial, rootCnext}, io.Discard)
		t.Logf("initial root commitment: %s", rootInitial)
		t.Logf("next root commitment: %s", rootCnext)

		trInit, err := trie.NewTrieReader(store, rootInitial)
		require.NoError(t, err)
		require.Empty(t, trInit.GetStr(""))

		v := tr.GetStr(key)
		require.EqualValues(t, value, v)

		require.True(t, tr.HasStr(key))
	})
	t.Run("update 2 long value"+"", func(t *testing.T) {
		store := NewInMemoryKVStore()
		const (
			key   = "key"
			value = "value"
		)

		rootInitial := lo.Must(trie.InitRoot(store, true))
		require.NotNil(t, rootInitial)

		tr, err := trie.NewTrieUpdatable(store, rootInitial)
		require.NoError(t, err)

		require.Empty(t, tr.GetStr(""))

		tr.UpdateStr(key, strings.Repeat(value, 500))
		rootCnext, _, stats := tr.Commit(store)
		require.NotZero(t, stats.CreatedValues)
		t.Logf("initial root commitment: %s", rootInitial)
		t.Logf("next root commitment: %s", rootCnext)
		trie.DebugDump(store, []trie.Hash{rootInitial, rootCnext}, io.Discard)

		require.Equal(t, rootCnext, tr.Root())

		require.Empty(t, tr.GetStr(""))

		v := tr.GetStr(key)
		require.EqualValues(t, strings.Repeat(value, 500), v)

		require.True(t, tr.HasStr(key))
	})
}

func TestBaseUpdate(t *testing.T) {
	runTest := func(data []string) {
		t.Run("update many", func(t *testing.T) {
			store := NewInMemoryKVStore()
			rootInitial := lo.Must(trie.InitRoot(store, true))
			require.NotNil(t, rootInitial)

			tr, err := trie.NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			// data = data[:2]
			for _, key := range data {
				value := strings.Repeat(key, 5)
				tr.UpdateStr(key, value)
			}
			rootNext, _, _ := tr.Commit(store)
			t.Logf("after commit: %s", rootNext)

			trie.DebugDump(store, []trie.Hash{rootInitial, rootNext}, io.Discard)

			err = tr.SetRoot(rootNext)
			require.NoError(t, err)

			for _, key := range data {
				v := tr.GetStr(key)
				require.EqualValues(t, strings.Repeat(key, 5), v)
			}
		})
	}
	data := []string{"ab", "acd", "a", "dba", "abc", "abd", "abcdafgh", "aaaaaaaaaaaaaaaa", "klmnt"}

	runTest([]string{"a", "ab"})
	runTest([]string{"ab", "acb"})
	runTest([]string{"abc", "a"})
	runTest(data)
}

var traceScenarios = false

func runUpdateScenario(trieUpdatable *trie.TrieUpdatable, store trie.KVStore, scenario []string) (
	checklist map[string]string,
	committedRoots []trie.Hash,
) {
	checklist = make(map[string]string)
	uncommitted := false
	var ret trie.Hash
	for _, cmd := range scenario {
		if cmd == "" {
			continue
		}
		if cmd == "*" {
			ret, _, _ = trieUpdatable.Commit(store)
			committedRoots = append(committedRoots, ret)
			if traceScenarios {
				fmt.Printf("+++ commit. Root: '%s'\n", ret)
			}
			uncommitted = false
			continue
		}
		var key, value []byte
		before, after, found := strings.Cut(cmd, "/")
		if found {
			if before == "" {
				panic("key must not be empty")
			}
			key = []byte(before)
			if after != "" {
				value = []byte(after)
			}
		} else {
			key = []byte(cmd)
			value = []byte(cmd)
		}
		trieUpdatable.Update(key, value)
		checklist[string(key)] = string(value)
		uncommitted = true
		if traceScenarios {
			if len(value) > 0 {
				fmt.Printf("SET '%s' -> '%s'\n", string(key), string(value))
			} else {
				fmt.Printf("DEL '%s'\n", string(key))
			}
		}
	}
	if uncommitted {
		ret, _, _ = trieUpdatable.Commit(store)
		committedRoots = append(committedRoots, ret)
		if traceScenarios {
			fmt.Printf("+++ commit. Root: '%s'\n", ret)
		}
	}
	if traceScenarios {
		fmt.Printf("+++ return root: '%s'\n", ret)
	}
	return checklist, committedRoots
}

func checkResult(t *testing.T, trie *trie.TrieUpdatable, checklist map[string]string) {
	for key, expectedValue := range checklist {
		v := trie.GetStr(key)
		if traceScenarios {
			if v != "" {
				fmt.Printf("FOUND '%s': '%s' (expected '%s')\n", key, v, expectedValue)
			} else {
				fmt.Printf("NOT FOUND '%s' (expected '%s')\n", key, func() string {
					if expectedValue != "" {
						return "FOUND"
					}
					return "NOT FOUND"
				}())
			}
		}
		require.EqualValues(t, expectedValue, v)
	}
}

func TestBaseScenarios(t *testing.T) {
	tf := func(data []string) func(t *testing.T) {
		return func(t *testing.T) {
			store := NewInMemoryKVStore()
			rootInitial := lo.Must(trie.InitRoot(store, true))
			require.NotNil(t, rootInitial)

			tr, err := trie.NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			checklist, newRoots := runUpdateScenario(tr, store, data)
			checkResult(t, tr, checklist)

			trie.DebugDump(store, append([]trie.Hash{rootInitial}, newRoots...), io.Discard)
		}
	}
	data1 := []string{"ab", "acd", "-a", "-ab", "abc", "abd", "abcdafgh", "-acd", "aaaaaaaaaaaaaaaa", "klmnt"}

	t.Run("1-1", tf([]string{"a", "a/"}))
	t.Run("1-2", tf([]string{"a", "*", "a/"}))
	t.Run("1-3", tf([]string{"a", "b", "*", "b/", "a/"}))
	t.Run("1-4", tf([]string{"a", "b", "*", "a/", "b/"}))
	t.Run("1-5", tf([]string{"a", "b", "*", "a/", "b/bb", "c"}))
	t.Run("1-6", tf([]string{"a", "b", "*", "a/", "b/bb", "c"}))
	t.Run("1-7", tf([]string{"a", "b", "*", "a/", "b", "c"}))
	t.Run("1-8", tf([]string{"acb/", "*", "acb/bca", "acb/123"}))
	t.Run("1-9", tf([]string{"abc", "a", "abc/", "a/"}))
	t.Run("1-10", tf([]string{"abc", "a", "a/", "abc/", "klmn"}))

	t.Run("8", tf(data1))

	t.Run("12", tf([]string{"a", "ab", "-a"}))

	data2 := []string{"a", "ab", "abc", "abcd", "abcde", "-abd", "-a"}
	t.Run("14", tf(data2))

	data3 := []string{"a", "ab", "abc", "abcd", "abcde", "-abcde", "-abcd", "-abc", "-ab", "-a"}
	t.Run("14", tf(data3))

	data4 := genRnd3()
	name := "update-many-"
	t.Run(name+"1", tf(data4))
}

func TestDeletionLoop(t *testing.T) {
	runTest := func(initScenario, scenario []string) {
		store := NewInMemoryKVStore()
		beginRoot := lo.Must(trie.InitRoot(store, true))
		roots := []trie.Hash{beginRoot}
		tr, err := trie.NewTrieUpdatable(store, beginRoot)
		require.NoError(t, err)
		t.Logf("TestDeletionLoop: model: '%s', init='%s', scenario='%s'", "", initScenario, scenario)
		_, roots1 := runUpdateScenario(tr, store, initScenario)
		roots = append(roots, roots1...)
		trie.DebugDump(store, roots, io.Discard)
		_, roots2 := runUpdateScenario(tr, store, scenario)
		require.Equal(t, roots[len(roots)-1], roots2[len(roots2)-1])
		roots = append(roots, roots2...)
		trie.DebugDump(store, roots, io.Discard)
	}
	runAll := func(init, sc []string) {
		runTest(init, sc)
	}
	runAll([]string{"a", "b"}, []string{"1", "*", "1/"})
	runAll([]string{"a", "ab", "abc"}, []string{"ac", "*", "ac/"})
	runAll([]string{"a", "ab", "abc"}, []string{"ac", "ac/"})
	runAll([]string{}, []string{"a", "a/"})
	runAll([]string{"a", "ab", "abc"}, []string{"a/", "a"})
	runAll([]string{}, []string{"a", "a/"})
	runAll([]string{"a"}, []string{"a/", "a"})
	runAll([]string{"a"}, []string{"b", "b/"})
	runAll([]string{"a"}, []string{"b", "*", "b/"})
	runAll([]string{"a", "bc"}, []string{"1", "*", "2", "*", "3", "1/", "2/", "3/"})
}

func TestDeterminism(t *testing.T) {
	tf := func(scenario1, scenario2 []string) func(t *testing.T) {
		return func(t *testing.T) {
			var root1, root2 trie.Hash
			{
				store1 := NewInMemoryKVStore()
				initRoot1 := lo.Must(trie.InitRoot(store1, true))
				roots1 := []trie.Hash{initRoot1}

				tr1, err := trie.NewTrieUpdatable(store1, initRoot1)
				require.NoError(t, err)

				checklist1, newRoots1 := runUpdateScenario(tr1, store1, scenario1)
				roots1 = append(roots1, newRoots1...)
				trie.DebugDump(store1, roots1, io.Discard)
				checkResult(t, tr1, checklist1)
				root1 = roots1[len(roots1)-1]
			}
			{
				store2 := NewInMemoryKVStore()
				initRoot2 := lo.Must(trie.InitRoot(store2, true))
				roots2 := []trie.Hash{initRoot2}

				tr2, err := trie.NewTrieUpdatable(store2, initRoot2)
				require.NoError(t, err)

				checklist2, newRoots2 := runUpdateScenario(tr2, store2, scenario2)
				roots2 = append(roots2, newRoots2...)
				trie.DebugDump(store2, roots2, io.Discard)
				checkResult(t, tr2, checklist2)
				root2 = roots2[len(roots2)-1]
			}
			require.Equal(t, root1, root2)
		}
	}
	{
		s1 := []string{"a", "ab"}
		s2 := []string{"ab", "a"}
		name := "order-simple-"
		t.Run(name+"1", tf(s1, s2))
	}
	{
		s1 := genRnd3()
		s2 := reverse(s1)
		name := "order-reverse-many-"
		t.Run(name+"1", tf(s1, s2))
	}
	{
		s1 := []string{"a", "ab"}
		s2 := []string{"a", "*", "ab"}
		name := "commit-simple-"
		t.Run(name+"1", tf(s1, s2))
	}
}

func TestRefcounts(t *testing.T) {
	store := NewInMemoryKVStore()
	root0 := lo.Must(trie.InitRoot(store, true))

	var root1 trie.Hash
	{
		tr := lo.Must(trie.NewTrieUpdatable(store, root0))
		tr.Update([]byte("key1"), bytes.Repeat([]byte{'x'}, 100))
		root1, _, _ = tr.Commit(store)
	}

	enabled, refcounts := trie.NewRefcounts(store)
	require.True(t, enabled)
	checkNode := func(h string, n uint32) {
		require.Equal(t, n, refcounts.GetNode(lo.Must(trie.HashFromBytes(lo.Must(hex.DecodeString(h))))))
	}
	checkValue := func(v string, n uint32) {
		require.Equal(t, n, refcounts.GetValue(lo.Must(hex.DecodeString(v))))
	}

	trie.DebugDump(store, []trie.Hash{root0, root1}, io.Discard)
	// [trie store]
	//  [] c:534f98b3ad630819d284287b647283a1d5dbcf90 ext:[] term:<nil>
	//  [] c:e71db7c574e3b92e39ae790f08b1a12321a75586 ext:[] term:<nil>
	//      [6] c:9d1a773be81d04ec2ccd0f82511fe4325c74b825 ext:[11 6 5 7 9 3 1] term:ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0
	//          [v: ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0 -> "xxxxxxxxxxxxxxxxx..."]
	//
	// all nodes and values have refcount = 1
	checkNode("e71db7c574e3b92e39ae790f08b1a12321a75586", 1)
	checkNode("534f98b3ad630819d284287b647283a1d5dbcf90", 1)
	checkNode("9d1a773be81d04ec2ccd0f82511fe4325c74b825", 1)
	checkValue("ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0", 1)

	var root2 trie.Hash
	{
		tr := lo.Must(trie.NewTrieUpdatable(store, root1))
		tr.Update([]byte("yyyy"), bytes.Repeat([]byte{'y'}, 100))
		root2, _, _ = tr.Commit(store)
	}

	_ = root2
	trie.DebugDump(store, []trie.Hash{root0, root1, root2}, io.Discard)
	// [trie store]
	//  [] c:534f98b3ad630819d284287b647283a1d5dbcf90 ext:[] term:<nil>
	//  [] c:e71db7c574e3b92e39ae790f08b1a12321a75586 ext:[] term:<nil>
	//      [6] c:9d1a773be81d04ec2ccd0f82511fe4325c74b825 ext:[11 6 5 7 9 3 1] term:ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0
	//          [v: ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0 -> "xxxxxxxxxxxxxxxxx..."]
	//  [] c:72c1b1df73c0765419c296f6a2765f685dd5bfe9 ext:[] term:<nil>
	//      [6] c:9d1a773be81d04ec2ccd0f82511fe4325c74b825 ext:[11 6 5 7 9 3 1] term:ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0
	//          [v: ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0 -> "xxxxxxxxxxxxxxxxx..."]
	//      [7] c:17a2f463569bbc3e2bd810c2d327819a6ff95e17 ext:[9 7 9 7 9 7 9] term:4c234c80dfe3d0069436a290ad85582b40835179
	//          [v: 4c234c80dfe3d0069436a290ad85582b40835179 -> "yyyyyyyyyyyyyyyyy..."]
	//
	// note that node 9d1a773be81d04ec2ccd0f82511fe4325c74b825 should have refcount = 2
	checkNode("9d1a773be81d04ec2ccd0f82511fe4325c74b825", 2)
	checkNode("17a2f463569bbc3e2bd810c2d327819a6ff95e17", 1)
	checkValue("ddeb47bbcdfd1d2b4355c1a66f22d302ecb05bb0", 1)
}

func TestTrieDAGEdgeCase(t *testing.T) {
	store := NewInMemoryKVStore()
	root0 := lo.Must(trie.InitRoot(store, true))

	// Compose a degenerate trie that is actually a DAG
	var root1 trie.Hash
	{
		tr := lo.Must(trie.NewTrieUpdatable(store, root0))
		tr.Update([]byte("a"), bytes.Repeat([]byte{'y'}, 100))
		tr.Update([]byte("axc"), bytes.Repeat([]byte{'y'}, 100))
		tr.Update([]byte("ayc"), bytes.Repeat([]byte{'y'}, 100))
		tr.Update([]byte("A"), bytes.Repeat([]byte{'y'}, 100))
		tr.Update([]byte("Axc"), bytes.Repeat([]byte{'y'}, 100))
		tr.Update([]byte("Ayc"), bytes.Repeat([]byte{'y'}, 100))
		root1, _, _ = tr.Commit(store)
	}

	enabled, refcounts := trie.NewRefcounts(store)
	require.True(t, enabled)
	checkNode := func(h string, n uint32) {
		require.Equal(t, n, refcounts.GetNode(lo.Must(trie.HashFromBytes(lo.Must(hex.DecodeString(h))))))
	}
	checkValue := func(v string, n uint32) {
		require.Equal(t, n, refcounts.GetValue(lo.Must(hex.DecodeString(v))))
	}

	trie.DebugDump(store, []trie.Hash{root0, root1}, io.Discard)
	// [trie store]
	//  [] c:534f98b3ad630819d284287b647283a1d5dbcf90 ext:[] term:<nil> (seen: 1)
	//  [] c:21d8e5ebf834af2b24e5bba418dd59929b2e6017 ext:[] term:<nil> (seen: 1)
	//      [4] c:465d25b13ec4bc17e35a30ed9a459c6a1a1ed04b ext:[1] term:4c234c80dfe3d0069436a290ad85582b40835179 (seen: 1)
	//          [v: 4c234c80dfe3d0069436a290ad85582b40835179 -> "yyyyyyyyyyyyyyyyy..."] (seen: 1)
	//          [7] c:c5b0cdba802bc5300c4c475096b678276d480b68 ext:[] term:<nil> (seen: 1)
	//              [8] c:23a7faaaa299574c0553d3b10592033ae839a76f ext:[6 3] term:4c234c80dfe3d0069436a290ad85582b40835179 (seen: 1)
	//                  [v: 4c234c80dfe3d0069436a290ad85582b40835179 -> "yyyyyyyyyyyyyyyyy..."] (seen: 2)
	//              [9] c:23a7faaaa299574c0553d3b10592033ae839a76f ext:[6 3] term:4c234c80dfe3d0069436a290ad85582b40835179 (seen: 2)
	//      [6] c:465d25b13ec4bc17e35a30ed9a459c6a1a1ed04b ext:[1] term:4c234c80dfe3d0069436a290ad85582b40835179 (seen: 2)
	// [node refcounts]
	//    21d8e5ebf834af2b24e5bba418dd59929b2e6017: 1
	//    465d25b13ec4bc17e35a30ed9a459c6a1a1ed04b: 2
	//    c5b0cdba802bc5300c4c475096b678276d480b68: 1
	//    23a7faaaa299574c0553d3b10592033ae839a76f: 2
	//    534f98b3ad630819d284287b647283a1d5dbcf90: 1
	// [value refcounts]
	//    4c234c80dfe3d0069436a290ad85582b40835179: 2

	checkNode("534f98b3ad630819d284287b647283a1d5dbcf90", 1)
	checkNode("21d8e5ebf834af2b24e5bba418dd59929b2e6017", 1)
	checkNode("c5b0cdba802bc5300c4c475096b678276d480b68", 1)
	checkNode("23a7faaaa299574c0553d3b10592033ae839a76f", 2)
	checkNode("465d25b13ec4bc17e35a30ed9a459c6a1a1ed04b", 2)
	checkValue("4c234c80dfe3d0069436a290ad85582b40835179", 2)

	// snapshot / restore all roots to a different store,
	// then check that the resulting store is identical
	{
		store2 := NewInMemoryKVStore()
		var roots2 []trie.Hash
		for _, root := range []trie.Hash{root0, root1} {
			buf := bytes.NewBuffer(nil)
			lo.Must(trie.NewTrieReader(store, root)).TakeSnapshot(buf)
			err := trie.RestoreSnapshot(buf, store2, true)
			require.NoError(t, err)
			roots2 = append(roots2, root)
			trie.DebugDump(store2, roots2, io.Discard)
		}
		require.Equal(t, store.m, store2.m)
	}

	// prune all roots one by one, then check that the resulting store is empty
	{
		roots := []trie.Hash{root0, root1}
		for len(roots) > 0 {
			root := roots[0]
			roots = roots[1:]
			_, err := trie.Prune(store, root)
			require.NoError(t, err)
			trie.DebugDump(store, roots, io.Discard)
		}
		trie.DeleteRefcountsFlag(store)
		require.Empty(t, store.m)
	}
}

func TestIterate(t *testing.T) {
	iterTest := func(scenario []string) func(t *testing.T) {
		return func(t *testing.T) {
			store := NewInMemoryKVStore()
			rootInitial := lo.Must(trie.InitRoot(store, true))
			require.NotNil(t, rootInitial)

			tr, err := trie.NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			checklist, roots := runUpdateScenario(tr, store, scenario)
			checkResult(t, tr, checklist)
			trie.DebugDump(store, append([]trie.Hash{rootInitial}, roots...), io.Discard)

			trr, err := trie.NewTrieReader(store, roots[len(roots)-1])
			require.NoError(t, err)
			var iteratedKeys1 [][]byte
			trr.Iterate(func(k []byte, v []byte) bool {
				if traceScenarios {
					fmt.Printf("---- iter --- '%s': '%s'\n", string(k), string(v))
				}
				require.NotEmpty(t, k)
				vCheck := checklist[string(k)]
				require.True(t, len(v) > 0)
				require.EqualValues(t, []byte(vCheck), v)
				iteratedKeys1 = append(iteratedKeys1, k)
				return true
			})

			// assert that iteration order is deterministic
			var iteratedKeys2 [][]byte
			trr.IterateKeys(func(k []byte) bool {
				iteratedKeys2 = append(iteratedKeys2, k)
				return true
			})
			require.EqualValues(t, iteratedKeys1, iteratedKeys2)
		}
	}
	{
		name := "iterate-one-"
		scenario := []string{"a"}
		t.Run(name+"1", iterTest(scenario))
	}
	{
		name := "iterate-"
		scenario := []string{"a", "b", "c", "*", "a/"}
		t.Run(name+"1", iterTest(scenario))
	}
	{
		name := "iterate-big-"
		scenario := genRnd3()
		t.Run(name+"1", iterTest(scenario))
	}
}

func TestIteratePrefix(t *testing.T) {
	iterTest := func(scenario []string, prefix string) func(t *testing.T) {
		return func(t *testing.T) {
			store := NewInMemoryKVStore()
			rootInitial := lo.Must(trie.InitRoot(store, true))
			require.NotNil(t, rootInitial)

			tr, err := trie.NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			_, roots := runUpdateScenario(tr, store, scenario)
			trie.DebugDump(store, append([]trie.Hash{rootInitial}, roots...), io.Discard)

			trr, err := trie.NewTrieReader(store, roots[len(roots)-1])
			require.NoError(t, err)

			countIter := 0
			trr.Iterator([]byte(prefix)).Iterate(func(k []byte, v []byte) bool {
				if traceScenarios {
					fmt.Printf("---- iter --- '%s': '%s'\n", string(k), string(v))
				}
				countIter++
				require.True(t, strings.HasPrefix(string(k), prefix))
				return true
			})
			countOrig := 0
			for _, s := range scenario {
				if strings.HasPrefix(s, prefix) {
					countOrig++
				}
			}
			require.EqualValues(t, countOrig, countIter)
		}
	}
	{
		name := "iterate-ab"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111"}
		prefix := "ab"
		t.Run(name+"1", iterTest(scenario, prefix))
	}
	{
		name := "iterate-a"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111", "baba", "ababa"}
		prefix := "a"
		t.Run(name+"1", iterTest(scenario, prefix))
	}
	{
		name := "iterate-empty"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111", "baba", "ababa"}
		prefix := ""
		t.Run(name+"1", iterTest(scenario, prefix))
	}
	{
		name := "iterate-none"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111", "baba", "ababa"}
		prefix := "---"
		t.Run(name+"1", iterTest(scenario, prefix))
	}
}

func TestDeletePrefix(t *testing.T) {
	iterTest := func(scenario []string, prefix string) func(t *testing.T) {
		return func(t *testing.T) {
			store := NewInMemoryKVStore()
			rootInitial := lo.Must(trie.InitRoot(store, true))
			require.NotNil(t, rootInitial)
			roots := []trie.Hash{rootInitial}

			tr, err := trie.NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			_, newRoots := runUpdateScenario(tr, store, scenario)
			roots = append(roots, newRoots...)
			trie.DebugDump(store, roots, io.Discard)

			tr, err = trie.NewTrieUpdatable(store, roots[len(roots)-1])
			require.NoError(t, err)

			deleted := tr.DeletePrefix([]byte(prefix))
			newRoot, _, _ := tr.Commit(store)
			roots = append(roots, newRoot)
			trie.DebugDump(store, roots, io.Discard)

			tr.Iterator([]byte(prefix)).Iterate(func(k []byte, v []byte) bool {
				if traceScenarios {
					fmt.Printf("---- iter --- '%s': '%s'\n", string(k), string(v))
				}
				require.NotEmpty(t, k)
				if deleted && prefix != "" {
					require.False(t, strings.HasPrefix(string(k), prefix))
				}
				return true
			})
		}
	}
	{
		name := "delete-ab"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111"}
		prefix := "ab"
		t.Run(name+"1", iterTest(scenario, prefix))
	}
	{
		name := "delete-a"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111", "baba", "ababa"}
		prefix := "a"
		t.Run(name+"1", iterTest(scenario, prefix))
	}
	{
		name := "delete-root"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111", "baba", "ababa"}
		prefix := ""
		t.Run(name+"1", iterTest(scenario, prefix))
	}
	{
		name := "delete-none"
		scenario := []string{"a", "ab", "c", "cd", "abcd", "klmn", "aaa", "abra", "111", "baba", "ababa"}
		prefix := "---"
		t.Run(name+"1", iterTest(scenario, prefix))
	}
}

const letters = "abcdefghijklmnop"

func genRnd3() []string {
	ret := make([]string, 0, len(letters)*len(letters)*len(letters))

	rnd := util.NewPseudoRand(1)
	for i := range letters {
		for j := range letters {
			for k := range letters {
				s := string([]byte{letters[i], letters[j], letters[k]})
				s = s + s + s + s
				r1 := rnd.Intn(len(s))
				r2 := rnd.Intn(len(s))
				if r2 < r1 {
					r1, r2 = r2, r1
				}
				ret = append(ret, s[r1:r2])
			}
		}
	}
	return ret
}

func reverse(orig []string) []string {
	ret := make([]string, 0, len(orig))
	for i := len(orig) - 1; i >= 0; i-- {
		ret = append(ret, orig[i])
	}
	return ret
}

type NewGeneratorFunc = func(int) Generator

type Generator interface {
	Next() string
}

//------------------------------------------------------------------------------

type ScrambledZipfian struct {
	r *rand.Rand
	z *generator.ScrambledZipfian
}

func NewScrambledZipfian(maximum int, seed int64) Generator {
	return &ScrambledZipfian{
		r: rand.New(rand.NewSource(seed)),
		z: generator.NewScrambledZipfian(0, int64(maximum), generator.ZipfianConstant),
	}
}

func (g *ScrambledZipfian) Next() string {
	return strconv.FormatUint(uint64(g.z.Next(g.r)), 10)
}

type CacheBenchmark struct {
	Init  func(size int)
	Close func()
	Get   func(key []byte) bool
	Set   func(key []byte, value []byte) bool
}

const (
	itemCost = 40 * 4
)

var (
	lruCache       *lru.Cache[string, []byte]
	ristrettoCache *ristretto.Cache
	fastcacheCache *fastcache.Cache
	theineCache    *theine.Cache[string, []byte]
	clockproCache  *clockpro.Cache

	cacheSize  = []int{1e4, 1e6}
	multiplier = []int{10, 100, 1000}
	caches     = map[string]CacheBenchmark{
		"golang-lru": {
			Init: func(size int) {
				lruCache, _ = lru.New[string, []byte](size)
			},
			Close: func() {
				lruCache.Purge()
				lruCache = nil
			},
			Get: func(key []byte) bool {
				_, ok := lruCache.Get(string(key))
				return ok
			},
			Set: func(key []byte, value []byte) bool {
				return lruCache.Add(string(key), value)
			},
		},
		"ristretto": {
			Init: func(size int) {
				ristrettoCache, _ = ristretto.NewCache(&ristretto.Config{
					// Keeps track of when keys are requested for LFU
					// Recommendation is to make this 10x the max items in the cache, each key uses ~3 bytes
					NumCounters: int64(size) * 10,
					// Max cost of the cache should be the approximate size of each item * the number of items we want in the cache
					MaxCost:     int64(size) * itemCost,
					BufferItems: 64,
				})
			},
			Close: ristrettoCache.Close,
			Get: func(key []byte) bool {
				_, ok := ristrettoCache.Get(key)
				return ok
			},
			Set: func(key, value []byte) bool {
				ok := ristrettoCache.Set(key, value, 0)
				return ok
			},
		},
		"fastcache": {
			Init: func(size int) {
				fastcacheCache = fastcache.New(size * itemCost)
			},
			Close: func() {
				fastcacheCache.Reset()
				fastcacheCache = nil
			},
			Get: func(key []byte) bool {
				if v := fastcacheCache.Get(nil, key); v == nil {
					return false
				}
				return true
			},
			Set: func(key, value []byte) bool {
				fastcacheCache.Set(key, value)
				return true
			},
		},
		"theine": {
			Init: func(size int) {
				theineCache, _ = theine.NewBuilder[string, []byte](int64(size)).Build()
			},
			Close: func() {
				theineCache.Close()
				theineCache = nil
			},
			Get: func(key []byte) bool {
				if _, ok := theineCache.Get(string(key)); ok {
					return true
				}
				return false
			},
			Set: func(key, value []byte) bool {
				return theineCache.Set(string(key), value, int64(len(value)))
			},
		},
		"clockpro": {
			Init: func(size int) {
				clockproCache = clockpro.New(size)
			},
			Close: func() {
				clockproCache = nil
			},
			Get: func(key []byte) bool {
				if v := clockproCache.Get(string(key)); v != nil {
					return true
				}
				return false
			},
			Set: func(key, value []byte) bool {
				clockproCache.Set(string(key), value)
				return true
			},
		},
	}
)

// Benchmark to specifically test cache hits/misses on LRU
func BenchmarkCaches(b *testing.B) {
	for name, cache := range caches {
		b.Run(name, func(b *testing.B) {
			// hitrates := make([]float64, len(cacheSize)*len(multiplier))
			for _, cacheSize := range cacheSize {
				b.Run(fmt.Sprintf("cacheSize_%d", cacheSize), func(b *testing.B) {
					for _, multiplier := range multiplier {
						b.Run(fmt.Sprintf("possibleKeys_%d", multiplier*cacheSize), func(b *testing.B) {
							b.StopTimer()
							var hits, misses float64
							cache.Init(cacheSize)
							gen := NewScrambledZipfian(cacheSize*multiplier, 0)
							var hitrate float64
							b.StartTimer()
							defer func() {
								b.StopTimer()
								cache.Close()
								b.StartTimer()
							}()
							for i := 0; i < b.N; i++ {
								b.StopTimer()
								key := gen.Next()
								value := []byte(fmt.Sprintf("just a bunch of dummy text for key %s", key))
								b.StartTimer()
								if cache.Get([]byte(key)) {
									hits++
								} else {
									misses++
									cache.Set([]byte(key), value)
								}
							}
							b.StopTimer()
							hitrate = hits / (hits + misses)
							b.Logf("sample size: %d, hits: %.2f, misses: %.2f, hitrate: %.2f", b.N, hits, misses, hitrate)
						})
					}
				})
			}
		})
	}
}
