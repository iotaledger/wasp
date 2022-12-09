package trie

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	store := NewInMemoryKVStore()

	var root0 Hash
	{
		root0 = MustInitRoot(store)
	}
	{
		state, err := NewTrieReader(store, root0)
		require.NoError(t, err)
		require.EqualValues(t, []byte(nil), state.Get([]byte("a")))
	}

	var root1 Hash
	{
		tr, err := NewTrieUpdatable(store, root0)
		require.NoError(t, err)
		tr.Update([]byte("a"), []byte("a"))
		tr.Update([]byte("b"), []byte("b"))
		root1 = tr.Commit(store)
	}

	var root2 Hash
	{
		tr, err := NewTrieUpdatable(store, root1)
		require.NoError(t, err)
		tr.Update([]byte("a"), nil)
		tr.Update([]byte("b"), []byte("bb"))
		tr.Update([]byte("cccddd"), []byte("c"))
		tr.Update([]byte("ccceee"), bytes.Repeat([]byte("c"), 70))
		root2 = tr.Commit(store)
		require.NoError(t, err)

		require.Nil(t, tr.Get([]byte("a")))
	}

	state, err := NewTrieReader(store, root2)
	require.NoError(t, err)
	require.Nil(t, state.Get([]byte("a")))

	var root3 Hash
	{
		tr, err := NewTrieUpdatable(store, root2)
		require.NoError(t, err)
		tr.Update([]byte("b"), nil)
		tr.Update([]byte("cccddd"), nil)
		tr.Update([]byte("ccceee"), nil)
		root3 = tr.Commit(store)
		require.NoError(t, err)

		require.Nil(t, tr.Get([]byte("a")))
		require.Nil(t, tr.Get([]byte("b")))
		require.Nil(t, tr.Get([]byte("cccddd")))
		require.Nil(t, tr.Get([]byte("ccceee")))
	}
	// trie is now empty, so hash3 should be equl to hash0
	require.Equal(t, root0, root3)
}

func TestBasic2(t *testing.T) {
	store := NewInMemoryKVStore()

	root0 := MustInitRoot(store)

	var root1 Hash
	{
		tr, err := NewTrieUpdatable(store, root0)
		require.NoError(t, err)
		tr.Update([]byte{0x00}, []byte{0})
		tr.Update([]byte{0x01}, []byte{0})
		tr.Update([]byte{0x10}, []byte{0})
		root1 = tr.Commit(store)
	}

	tr, err := NewTrieReader(store, root1)
	require.NoError(t, err)
	require.True(t, tr.Has([]byte{0x00}))
	require.True(t, tr.Has([]byte{0x01}))
	require.True(t, tr.Has([]byte{0x10}))
}

func TestBasic3(t *testing.T) {
	store := NewInMemoryKVStore()

	root0 := MustInitRoot(store)

	var root1 Hash
	{
		tr, err := NewTrieUpdatable(store, root0)
		require.NoError(t, err)
		tr.Update([]byte{0x30}, []byte{1})
		tr.Update([]byte{0x31}, []byte{1})
		tr.Update([]byte{0xb0}, []byte{1})
		tr.Update([]byte{0xb2}, []byte{1})
		root1 = tr.Commit(store)
	}

	tr, err := NewTrieReader(store, root1)
	require.NoError(t, err)
	require.Equal(t, []byte{1}, tr.Get([]byte{0x30}))
	require.Equal(t, []byte{1}, tr.Get([]byte{0x31}))
	require.Equal(t, []byte{1}, tr.Get([]byte{0xb0}))
	require.Equal(t, []byte{1}, tr.Get([]byte{0xb2}))
}

func TestCreateTrie(t *testing.T) {
	t.Run("ok init-"+"", func(t *testing.T) {
		rootC1 := MustInitRoot(NewInMemoryKVStore())
		require.NotNil(t, rootC1)

		rootC2 := MustInitRoot(NewInMemoryKVStore())
		require.NotNil(t, rootC2)

		require.Equal(t, rootC1, rootC2)
	})
	t.Run("update 1"+"", func(t *testing.T) {
		store := NewInMemoryKVStore()
		const (
			key   = "key"
			value = "value"
		)

		rootInitial := MustInitRoot(store)
		require.NotNil(t, rootInitial)

		tr, err := NewTrieUpdatable(store, rootInitial)
		require.NoError(t, err)

		require.Empty(t, tr.GetStr(""))

		tr.UpdateStr(key, value)
		rootCnext := tr.Commit(store)
		t.Logf("initial root commitment: %s", rootInitial)
		t.Logf("next root commitment: %s", rootCnext)

		trInit, err := NewTrieReader(store, rootInitial)
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

		rootInitial := MustInitRoot(store)
		require.NotNil(t, rootInitial)

		tr, err := NewTrieUpdatable(store, rootInitial)
		require.NoError(t, err)

		require.Empty(t, tr.GetStr(""))

		tr.UpdateStr(key, strings.Repeat(value, 500))
		rootCnext := tr.Commit(store)
		t.Logf("initial root commitment: %s", rootInitial)
		t.Logf("next root commitment: %s", rootCnext)

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
			rootInitial := MustInitRoot(store)
			require.NotNil(t, rootInitial)

			tr, err := NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			// data = data[:2]
			for _, key := range data {
				value := strings.Repeat(key, 5)
				tr.UpdateStr(key, value)
			}
			rootNext := tr.Commit(store)
			t.Logf("after commit: %s", rootNext)

			err = tr.setRoot(rootNext)
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

func runUpdateScenario(trie *TrieUpdatable, store KVWriter, scenario []string) (map[string]string, Hash) {
	checklist := make(map[string]string)
	uncommitted := false
	var ret Hash
	for _, cmd := range scenario {
		if len(cmd) == 0 {
			continue
		}
		if cmd == "*" {
			ret = trie.Commit(store)
			if traceScenarios {
				fmt.Printf("+++ commit. Root: '%s'\n", ret)
			}
			uncommitted = false
			continue
		}
		var key, value []byte
		before, after, found := strings.Cut(cmd, "/")
		if found {
			if len(before) == 0 {
				continue // key must not be empty
			}
			key = []byte(before)
			if len(after) > 0 {
				value = []byte(after)
			}
		} else {
			key = []byte(cmd)
			value = []byte(cmd)
		}
		trie.Update(key, value)
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
		ret = trie.Commit(store)
		if traceScenarios {
			fmt.Printf("+++ commit. Root: '%s'\n", ret)
		}
	}
	if traceScenarios {
		fmt.Printf("+++ return root: '%s'\n", ret)
	}
	return checklist, trie.Root()
}

func checkResult(t *testing.T, trie *TrieUpdatable, checklist map[string]string) {
	for key, expectedValue := range checklist {
		v := trie.GetStr(key)
		if traceScenarios {
			if len(v) > 0 {
				fmt.Printf("FOUND '%s': '%s' (expected '%s')\n", key, v, expectedValue)
			} else {
				fmt.Printf("NOT FOUND '%s' (expected '%s')\n", key, func() string {
					if len(expectedValue) > 0 {
						return "FOUND"
					} else {
						return "NOT FOUND"
					}
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
			rootInitial := MustInitRoot(store)
			require.NotNil(t, rootInitial)

			tr, err := NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			checklist, _ := runUpdateScenario(tr, store, data)
			checkResult(t, tr, checklist)
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
		beginRoot := MustInitRoot(store)
		tr, err := NewTrieUpdatable(store, beginRoot)
		require.NoError(t, err)
		t.Logf("TestDeletionLoop: model: '%s', init='%s', scenario='%s'", "", initScenario, scenario)
		_, beginRoot = runUpdateScenario(tr, store, initScenario)
		_, endRoot := runUpdateScenario(tr, store, scenario)
		require.Equal(t, beginRoot, endRoot)
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
			store1 := NewInMemoryKVStore()
			initRoot1 := MustInitRoot(store1)

			tr1, err := NewTrieUpdatable(store1, initRoot1)
			require.NoError(t, err)

			checklist1, root1 := runUpdateScenario(tr1, store1, scenario1)
			checkResult(t, tr1, checklist1)

			store2 := NewInMemoryKVStore()
			initRoot2 := MustInitRoot(store2)

			tr2, err := NewTrieUpdatable(store2, initRoot2)
			require.NoError(t, err)

			checklist2, root2 := runUpdateScenario(tr2, store2, scenario2)
			checkResult(t, tr2, checklist2)

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

func TestIterate(t *testing.T) {
	iterTest := func(scenario []string) func(t *testing.T) {
		return func(t *testing.T) {
			store := NewInMemoryKVStore()
			rootInitial := MustInitRoot(store)
			require.NotNil(t, rootInitial)

			tr, err := NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			checklist, root := runUpdateScenario(tr, store, scenario)
			checkResult(t, tr, checklist)

			trr, err := NewTrieReader(store, root, 0)
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
			rootInitial := MustInitRoot(store)
			require.NotNil(t, rootInitial)

			tr, err := NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			_, root := runUpdateScenario(tr, store, scenario)

			trr, err := NewTrieReader(store, root, 0)
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
			rootInitial := MustInitRoot(store)
			require.NotNil(t, rootInitial)

			tr, err := NewTrieUpdatable(store, rootInitial)
			require.NoError(t, err)

			_, root := runUpdateScenario(tr, store, scenario)

			tr, err = NewTrieUpdatable(store, root, 0)
			require.NoError(t, err)

			deleted := tr.DeletePrefix([]byte(prefix))
			tr.Commit(store)

			tr.Iterator([]byte(prefix)).Iterate(func(k []byte, v []byte) bool {
				if traceScenarios {
					fmt.Printf("---- iter --- '%s': '%s'\n", string(k), string(v))
				}
				require.NotEmpty(t, k)
				if deleted && len(prefix) != 0 {
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
	rnd := rand.New(rand.NewSource(1))
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

func TestSnapshot1(t *testing.T) {
	runTest := func(name string, data []string) {
		t.Run(name, func(t *testing.T) {
			store1 := NewInMemoryKVStore()
			initRoot1 := MustInitRoot(store1)
			tr1, err := NewTrieUpdatable(store1, initRoot1)
			require.NoError(t, err)

			_, root1 := runUpdateScenario(tr1, store1, data)
			storeData := NewInMemoryKVStore()
			tr1.SnapshotData(storeData)

			store2 := NewInMemoryKVStore()
			initRoot2 := MustInitRoot(store2)
			tr2, err := NewTrieUpdatable(store2, initRoot2)
			require.NoError(t, err)

			storeData.Iterate(func(k, v []byte) bool {
				tr2.Update(k, v)
				return true
			})
			root2 := tr2.Commit(store2)

			require.Equal(t, root1, root2)
		})
	}
	runTest("1", []string{"a", "ab", "abc", "1", "2", "3", "11"})
	runTest("rnd", genRnd3())
}

func TestSnapshot2(t *testing.T) {
	runTest := func(data []string) {
		store1 := NewInMemoryKVStore()
		initRoot1 := MustInitRoot(store1)
		tr1, err := NewTrieUpdatable(store1, initRoot1)
		require.NoError(t, err)

		_, root1 := runUpdateScenario(tr1, store1, data)
		store2 := NewInMemoryKVStore()
		tr1.Snapshot(store2)

		tr2, err := NewTrieUpdatable(store2, root1)
		require.NoError(t, err)

		sc1 := []string{"@", "#$%%^", "____++++", "~~~~~"}
		sc2 := []string{"@", "#$%%^", "*", "____++++", "~~~~~"}
		_, r1 := runUpdateScenario(tr1, store1, sc1)
		_, r2 := runUpdateScenario(tr2, store2, sc2)
		require.Equal(t, r1, r2)
	}
	{
		data := []string{"a", "ab", "abc", "1", "2", "3", "11"}
		runTest(data)
	}
	{
		data := genRnd3()
		runTest(data)
	}
}
