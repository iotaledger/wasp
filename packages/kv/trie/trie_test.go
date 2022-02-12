package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestNode(t *testing.T) {
	t.Run("base1", func(t *testing.T) {
		n := NewNode()
		var buf bytes.Buffer
		n.Write(&buf)
		t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
		require.EqualValues(t, Size(n), len(buf.Bytes()))

		nBack, err := MerkleTrieSetup.NodeFromBytes(buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), Bytes(nBack))

		h := MerkleTrieSetup.CommitToChildren(n)
		hBack := MerkleTrieSetup.CommitToChildren(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
	t.Run("base short terminal", func(t *testing.T) {
		n := NewNode()
		n.pathFragment = []byte("kuku")
		n.terminalCommitment = MerkleTrieSetup.CommitToData([]byte("data"))

		var buf bytes.Buffer
		n.Write(&buf)
		t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
		require.EqualValues(t, Size(n), len(buf.Bytes()))

		nBack, err := MerkleTrieSetup.NodeFromBytes(buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), Bytes(nBack))

		h := MerkleTrieSetup.CommitToChildren(n)
		hBack := MerkleTrieSetup.CommitToChildren(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
	t.Run("base long terminal", func(t *testing.T) {
		n := NewNode()
		n.pathFragment = []byte("kuku")
		n.terminalCommitment = MerkleTrieSetup.CommitToData([]byte(strings.Repeat("data", 1000)))

		var buf bytes.Buffer
		n.Write(&buf)
		t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
		require.EqualValues(t, Size(n), len(buf.Bytes()))

		nBack, err := MerkleTrieSetup.NodeFromBytes(buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), Bytes(nBack))

		h := MerkleTrieSetup.CommitToChildren(n)
		hBack := MerkleTrieSetup.CommitToChildren(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
}

func TestTrieBase(t *testing.T) {
	var data1 = []string{"", "1", "2"}
	var data2 = []string{"a", "ab", "ac", "abc", "abd", "ad", "ada", "adb", "adc", "c"}

	t.Run("base1", func(t *testing.T) {
		store := dict.New()
		tr := NewTrie(MerkleTrieSetup, store, nil)
		require.EqualValues(t, nil, tr.RootCommitment())

		tr.Update([]byte(data1[0]), []byte(data1[0]))
		tr.Commit()
		t.Logf("root0 = %s", tr.RootCommitment())
		c := tr.RootCommitment()
		rootNode, ok := tr.GetNode(nil)
		require.True(t, ok)
		require.EqualValues(t, c, tr.setup.CommitToChildren(rootNode))
	})
	t.Run("base2", func(t *testing.T) {
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data1 {
			tr1.Update([]byte(data1[i]), []byte(data1[i]))
			tr1.Commit()
		}
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		for i := range data1 {
			tr2.Update([]byte(data1[i]), []byte(data1[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("base3", func(t *testing.T) {
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data2 {
			tr1.Update([]byte(data2[i]), []byte(data2[i]))
			tr1.Commit()
		}
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		for i := range data2 {
			tr2.Update([]byte(data2[i]), []byte(data2[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()
		require.True(t, c1.Equal(c2))
	})
	t.Run("base4", func(t *testing.T) {
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data2 {
			tr1.Update([]byte(data2[i]), []byte(data2[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		for i := len(data2) - 1; i >= 0; i-- {
			tr2.Update([]byte(data2[i]), []byte(data2[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
}

func genRnd1() []string {
	str := "0123456789abcdef"
	ret := make([]string, 0, len(str)*len(str)*len(str))
	for i := range str {
		for j := range str {
			for k := range str {
				ret = append(ret, string([]byte{str[i], str[j], str[k]}))
			}
		}
	}
	return ret
}

func genRnd2() []string {
	str := "0123456789abcdef"
	ret := make([]string, 0, len(str)*len(str)*len(str))
	for i := range str {
		for j := range str {
			for k := range str {
				s := string([]byte{str[i], str[j], str[k]})
				ret = append(ret, s+s+s+s)
			}
		}
	}
	return ret
}

func genRnd3() []string {
	str := "0123456789abcdef"
	ret := make([]string, 0, len(str)*len(str)*len(str))
	for i := range str {
		for j := range str {
			for k := range str {
				s := string([]byte{str[i], str[j], str[k]})
				s = s + s + s + s
				r1 := rand.Intn(len(s))
				r2 := rand.Intn(len(s))
				if r2 < r1 {
					r1, r2 = r2, r1
				}
				ret = append(ret, s[r1:r2])
			}
		}
	}
	return ret
}

func genRnd4() []string {
	str := "0123456789abcdef"
	ret := make([]string, 0, len(str)*len(str)*len(str))
	for i := range str {
		for j := range str {
			for k := range str {
				for l := range str {
					s := string([]byte{str[i], str[j], str[k], str[l]})
					s = s + s + s + s
					r1 := rand.Intn(len(s))
					r2 := rand.Intn(len(s))
					if r2 < r1 {
						r1, r2 = r2, r1
					}
					ret = append(ret, s[r1:r2])
				}
			}
		}
	}
	if len(ret) > math.MaxUint16 {
		ret = ret[:math.MaxUint16]
	}
	return ret
}

func TestTrieRnd(t *testing.T) {
	t.Run("rnd1", func(t *testing.T) {
		data := genRnd1()
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		for i := len(data) - 1; i >= 0; i-- {
			tr2.Update([]byte(data[i]), []byte(data[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
	t.Run("determinism1", func(t *testing.T) {
		data := genRnd1()
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		permutation := util.NewPermutation16(uint16(len(data)), nil)
		permutation.ForEach(func(i uint16) bool {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			return true
		})
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
	t.Run("determinism2", func(t *testing.T) {
		data := genRnd2()
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		permutation := util.NewPermutation16(uint16(len(data)), nil)
		permutation.ForEach(func(i uint16) bool {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			return true
		})
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
	t.Run("determinism3", func(t *testing.T) {
		data := genRnd3()
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		permutation := util.NewPermutation16(uint16(len(data)), nil)
		permutation.ForEach(func(i uint16) bool {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			return true
		})
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
	t.Run("determinism4", func(t *testing.T) {
		data := genRnd4()
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		permutation := util.NewPermutation16(uint16(len(data)), nil)
		permutation.ForEach(func(i uint16) bool {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			return true
		})
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))

		tr2.FlushCache(store2)
		trieSize := len(store2.Bytes())
		t.Logf("key entries = %d", len(data))
		t.Logf("trie entries = %d", len(store2))
		t.Logf("trie bytes = %d KB", trieSize/1024)
		t.Logf("trie bytes/entry = %d ", trieSize/len(store2))
	})
	t.Run("determinism5", func(t *testing.T) {
		data := genRnd4()
		store1 := dict.New()
		tr1 := NewTrie(MerkleTrieSetup, store1, nil)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
			tr1.Commit()
		}
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := NewTrie(MerkleTrieSetup, store2, nil)

		permutation := util.NewPermutation16(uint16(len(data)), nil)
		permutation.ForEach(func(i uint16) bool {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			return true
		})
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
}
