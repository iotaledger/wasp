package trie_merkle

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestNode(t *testing.T) {
	t.Run("base1", func(t *testing.T) {
		n := trie.NewNode(nil)
		var buf bytes.Buffer
		err := n.Write(&buf)
		require.NoError(t, err)
		t.Logf("size() = %d, size(serialize) = %d", trie.MustSize(n), len(buf.Bytes()))
		require.EqualValues(t, trie.MustSize(n), len(buf.Bytes()))

		nBack, err := trie.NodeFromBytes(CommitmentLogic, buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), trie.MustBytes(nBack))

		h := CommitmentLogic.CommitToNode(n)
		hBack := CommitmentLogic.CommitToNode(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
	t.Run("base short terminal", func(t *testing.T) {
		n := trie.NewNode([]byte("kuku"))
		n.Terminal = CommitmentLogic.CommitToData([]byte("data"))

		var buf bytes.Buffer
		err := n.Write(&buf)
		require.NoError(t, err)
		t.Logf("size() = %d, size(serialize) = %d", trie.MustSize(n), len(buf.Bytes()))
		require.EqualValues(t, trie.MustSize(n), len(buf.Bytes()))

		nBack, err := trie.NodeFromBytes(CommitmentLogic, buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), trie.MustBytes(nBack))

		h := CommitmentLogic.CommitToNode(n)
		hBack := CommitmentLogic.CommitToNode(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
	t.Run("base long terminal", func(t *testing.T) {
		n := trie.NewNode([]byte("kuku"))
		n.Terminal = CommitmentLogic.CommitToData([]byte(strings.Repeat("data", 1000)))
		var buf bytes.Buffer
		err := n.Write(&buf)
		require.NoError(t, err)
		t.Logf("size() = %d, size(serialize) = %d", trie.MustSize(n), len(buf.Bytes()))
		require.EqualValues(t, trie.MustSize(n), len(buf.Bytes()))

		nBack, err := trie.NodeFromBytes(CommitmentLogic, buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), trie.MustBytes(nBack))

		h := CommitmentLogic.CommitToNode(n)
		hBack := CommitmentLogic.CommitToNode(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
}

func TestTrieBase(t *testing.T) {
	var data1 = []string{"", "1", "2"}
	var data2 = []string{"a", "ab", "ac", "abc", "abd", "ad", "ada", "adb", "adc", "c"}

	t.Run("base1", func(t *testing.T) {
		store := dict.New()
		tr := trie.New(CommitmentLogic, store)
		require.EqualValues(t, nil, tr.RootCommitment())

		tr.Update([]byte(data1[0]), []byte(data1[0]))
		tr.Commit()
		t.Logf("root0 = %s", tr.RootCommitment())
		_, ok := tr.GetNode(nil)
		require.False(t, ok)

		tr.Update([]byte(""), []byte("0"))
		tr.Commit()
		t.Logf("root0 = %s", tr.RootCommitment())
		c := tr.RootCommitment()
		rootNode, ok := tr.GetNode(nil)
		require.True(t, ok)
		require.EqualValues(t, c, tr.CommitToNode(rootNode))
	})
	t.Run("base2", func(t *testing.T) {
		store1 := dict.New()
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data1 {
			tr1.Update([]byte(data1[i]), []byte(data1[i]))
			tr1.Commit()
		}
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

		for i := range data1 {
			tr2.Update([]byte(data1[i]), []byte(data1[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("base3", func(t *testing.T) {
		store1 := dict.New()
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data2 {
			tr1.Update([]byte(data2[i]), []byte(data2[i]))
			tr1.Commit()
		}
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

		for i := range data2 {
			tr2.Update([]byte(data2[i]), []byte(data2[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()
		require.True(t, c1.Equal(c2))
	})
	t.Run("base4", func(t *testing.T) {
		store1 := dict.New()
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data2 {
			tr1.Update([]byte(data2[i]), []byte(data2[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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

func genDels(data []string, num int) []string {
	ret := make([]string, 0, num)
	for i := 0; i < num; i++ {
		ret = append(ret, data[rand.Intn(len(data))])
	}
	return ret
}

func gen2different(n int) ([]string, []string) {
	orig := genRnd4()
	// filter different
	unique := make(map[string]bool)
	for _, s := range orig {
		unique[s] = true
	}
	ret1 := make([]string, 0)
	ret2 := make([]string, 0)
	for s := range unique {
		if rand.Intn(10000) > 1000 {
			ret1 = append(ret1, s)
		} else {
			ret2 = append(ret2, s)
		}
		if len(ret1)+len(ret2) > n {
			break
		}
	}
	return ret1, ret2
}

func TestTrieRnd(t *testing.T) {
	t.Run("rnd1", func(t *testing.T) {
		data := genRnd1()
		store1 := dict.New()
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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

		tr2.FlushDelta(store2)
		trieSize := len(store2.Bytes())
		t.Logf("key entries = %d", len(data))
		t.Logf("Trie entries = %d", len(store2))
		t.Logf("Trie bytes = %d KB", trieSize/1024)
		t.Logf("Trie bytes/entry = %d ", trieSize/len(store2))
	})
	t.Run("determinism5", func(t *testing.T) {
		data := genRnd4()
		store1 := dict.New()
		tr1 := trie.New(CommitmentLogic, store1)

		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
			tr1.Commit()
		}
		c1 := tr1.RootCommitment()

		store2 := dict.New()
		tr2 := trie.New(CommitmentLogic, store2)

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

func TestTrieWithDeletion(t *testing.T) {
	data := []string{"0", "1", "2", "3", "4", "5"}
	var tr1, tr2 *trie.Trie
	initTest := func() {
		store1 := dict.New()
		tr1 = trie.New(CommitmentLogic, store1)
		store2 := dict.New()
		tr2 = trie.New(CommitmentLogic, store2)
	}
	t.Run("del1", func(t *testing.T) {
		initTest()
		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		for i := range data {
			tr2.Update([]byte(data[i]), []byte(data[i]))
		}
		tr2.Delete([]byte(data[1]))
		tr2.Update([]byte(data[1]), []byte(data[1]))
		tr2.Commit()
		c2 := tr1.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("del2", func(t *testing.T) {
		initTest()
		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		for i := range data {
			tr2.Update([]byte(data[i]), []byte(data[i]))
		}
		tr2.Commit()
		tr2.Delete([]byte(data[1]))
		tr2.Update([]byte(data[1]), []byte(data[1]))
		tr2.Commit()
		c2 := tr1.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("del3", func(t *testing.T) {
		initTest()
		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		for i := range data {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			tr2.Commit()
		}
		tr2.Delete([]byte(data[1]))
		tr2.Update([]byte(data[1]), []byte(data[1]))
		tr2.Commit()
		c2 := tr1.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("del4", func(t *testing.T) {
		initTest()
		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		for i := range data {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			tr2.Commit()
		}
		tr2.Delete([]byte(data[1]))
		tr2.Commit()
		tr2.Update([]byte(data[1]), []byte(data[1]))
		tr2.Commit()
		c2 := tr1.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("del5", func(t *testing.T) {
		initTest()
		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		for i := range data {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			tr2.Commit()
		}
		c2 := tr1.RootCommitment()
		require.True(t, c1.Equal(c2))

		tr2.Delete([]byte(data[1]))
		tr2.Commit()
		c2 = tr2.RootCommitment()
		require.False(t, c1.Equal(c2))

		tr2.Update([]byte(data[1]), []byte(data[1]))
		tr2.Commit()
		c2 = tr1.RootCommitment()

		require.True(t, c1.Equal(c2))
	})
	t.Run("del determ", func(t *testing.T) {
		initTest()
		data = genRnd4()
		dels := genDels(data, 1000)

		posDel := 0
		for i := range data {
			tr1.Update([]byte(data[i]), []byte(data[i]))
			tr1.Commit()
			if i%10 == 10 {
				tr1.Delete([]byte(dels[posDel]))
				posDel = (posDel + 1) % len(dels)
			}
		}
		tr1.Commit()
		for i := range dels {
			tr1.Delete([]byte(dels[i]))
		}
		tr1.Commit()
		c1 := tr1.RootCommitment()

		permutation := util.NewPermutation16(uint16(len(data)), nil)
		permutation.ForEach(func(i uint16) bool {
			tr2.Update([]byte(data[i]), []byte(data[i]))
			return true
		})
		for i := range dels {
			tr2.Delete([]byte(dels[i]))
		}
		tr2.Commit()
		c2 := tr2.RootCommitment()
		t.Logf("root1 = %s", c1)
		t.Logf("root2 = %s", c2)
		require.True(t, c1.Equal(c2))
	})
}

func TestTrieProof(t *testing.T) {
	//data1 := []string{"", "1", "2"}

	t.Run("proof empty tie", func(t *testing.T) {
		store := dict.New()
		tr := trie.New(CommitmentLogic, store)
		require.EqualValues(t, nil, tr.RootCommitment())

		proof := tr.ProofPath(nil)
		require.EqualValues(t, 0, len(proof.Path))
	})
	t.Run("proof one entry 1", func(t *testing.T) {
		store := dict.New()
		tr := trie.New(CommitmentLogic, store)

		tr.Update(nil, []byte("1"))
		tr.Commit()

		proofPath := tr.ProofPath(nil)
		require.EqualValues(t, 1, len(proofPath.Path))

		proof := CommitmentLogic.Proof(proofPath)
		rootC := tr.RootCommitment()
		err := proof.Validate(rootC)
		require.NoError(t, err)

		t.Logf("proof presence size = %d bytes", trie.MustSize(proof))

		key, term, _ := proof.MustKeyTerminal()
		c, _ := commitToTerminal([]byte("1")).value()
		require.EqualValues(t, 0, len(key))
		require.EqualValues(t, term, c[:])

		proofPath = tr.ProofPath([]byte("a"))
		require.EqualValues(t, 1, len(proofPath.Path))

		proof = CommitmentLogic.Proof(proofPath)
		rootC = tr.RootCommitment()
		err = proof.Validate(rootC)
		require.NoError(t, err)
		require.True(t, proof.MustIsProofOfAbsence())
		t.Logf("proof absence size = %d bytes", trie.MustSize(proof))
	})
	t.Run("proof one entry 2", func(t *testing.T) {
		store := dict.New()
		tr := trie.New(CommitmentLogic, store)

		tr.Update([]byte("1"), []byte("2"))
		tr.Commit()
		proofPath := tr.ProofPath(nil)
		require.EqualValues(t, 1, len(proofPath.Path))

		proof := CommitmentLogic.Proof(proofPath)
		rootC := tr.RootCommitment()
		err := proof.Validate(rootC)
		require.NoError(t, err)
		require.True(t, proof.MustIsProofOfAbsence())

		proofPath = tr.ProofPath([]byte("1"))
		require.EqualValues(t, 1, len(proofPath.Path))

		proof = CommitmentLogic.Proof(proofPath)
		err = proof.Validate(rootC)
		require.NoError(t, err)
		require.False(t, proof.MustIsProofOfAbsence())

		_, term, _ := proof.MustKeyTerminal()
		c, _ := commitToTerminal([]byte("2")).value()
		require.EqualValues(t, term, c)
	})
}

func TestTrieProofWithDeletes(t *testing.T) {
	var tr *trie.Trie
	var rootC trie.VectorCommitment

	initTrie := func(dataAdd []string) {
		store := dict.New()
		tr = trie.New(CommitmentLogic, store)
		for _, s := range dataAdd {
			tr.Update([]byte(s), []byte(s+"++"))
		}
	}
	deleteKeys := func(keysDelete []string) {
		for _, s := range keysDelete {
			tr.Update([]byte(s), nil)
		}
	}
	commitTrie := func() trie.VectorCommitment {
		tr.Commit()
		return tr.RootCommitment()
	}
	data := []string{"a", "ab", "ac", "abc", "abd", "ad", "ada", "adb", "adc", "c", "ad+dddgsssisd"}
	t.Run("proof many entries 1", func(t *testing.T) {
		initTrie(data)
		rootC = commitTrie()
		for _, s := range data {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.False(t, proof.MustIsProofOfAbsence())
			t.Logf("key: '%s', proof lenPlus1: %d", s, len(proofPath.Path))
			t.Logf("proof presence size = %d bytes", trie.MustSize(proof))
		}
	})
	t.Run("proof many entries 2", func(t *testing.T) {
		delKeys := []string{"1", "2", "3", "12345", "ab+", "ada+"}
		initTrie(data)
		deleteKeys(delKeys)
		rootC = commitTrie()

		for _, s := range data {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.False(t, proof.MustIsProofOfAbsence())
			t.Logf("key: '%s', proof presence lenPlus1: %d", s, len(proofPath.Path))
			t.Logf("proof presence size = %d bytes", trie.MustSize(proof))
		}
		for _, s := range delKeys {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.True(t, proof.MustIsProofOfAbsence())
			t.Logf("key: '%s', proof absence lenPlus1: %d", s, len(proofPath.Path))
			t.Logf("proof absence size = %d bytes", trie.MustSize(proof))
		}
	})
	t.Run("proof many entries 3", func(t *testing.T) {
		delKeys := []string{"1", "2", "3", "12345", "ab+", "ada+"}
		allData := make([]string, 0, len(data)+len(delKeys))
		allData = append(allData, data...)
		allData = append(allData, delKeys...)
		initTrie(allData)
		deleteKeys(delKeys)
		rootC = commitTrie()

		for _, s := range data {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.False(t, proof.MustIsProofOfAbsence())
			t.Logf("key: '%s', proof presence lenPlus1: %d", s, len(proofPath.Path))
			sz := trie.MustSize(proof)
			t.Logf("proof presence size = %d bytes", sz)

			proofBin := trie.MustBytes(proof)
			require.EqualValues(t, len(proofBin), sz)
			proofBack, err := ProofFromBytes(proofBin)
			require.NoError(t, err)
			err = proofBack.Validate(rootC)
			require.NoError(t, err)
			require.EqualValues(t, proof.Key, proofBack.Key)
			require.False(t, proofBack.MustIsProofOfAbsence())
		}
		for _, s := range delKeys {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.True(t, proof.MustIsProofOfAbsence())
			t.Logf("key: '%s', proof absence lenPlus1: %d", s, len(proofPath.Path))
			sz := trie.MustSize(proof)
			t.Logf("proof absence size = %d bytes", sz)

			proofBin := trie.MustBytes(proof)
			require.EqualValues(t, len(proofBin), sz)
			proofBack, err := ProofFromBytes(proofBin)
			require.NoError(t, err)
			err = proofBack.Validate(rootC)
			require.NoError(t, err)
			require.EqualValues(t, proof.Key, proofBack.Key)
			require.True(t, proofBack.MustIsProofOfAbsence())
		}
	})
	t.Run("proof many entries rnd", func(t *testing.T) {
		addKeys, delKeys := gen2different(100000)
		t.Logf("lenPlus1 adds: %d, lenPlus1 dels: %d", len(addKeys), len(delKeys))
		allData := make([]string, 0, len(addKeys)+len(delKeys))
		allData = append(allData, addKeys...)
		allData = append(allData, delKeys...)
		initTrie(allData)
		deleteKeys(delKeys)
		rootC = commitTrie()

		lenStats := make(map[int]int)
		size100Stats := make(map[int]int)
		for _, s := range addKeys {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.False(t, proof.MustIsProofOfAbsence())
			lenP := len(proofPath.Path)
			sizeP100 := trie.MustSize(proof) / 100
			//t.Logf("key: '%s', proof presence lenPlus1: %d", s, )
			t.Logf("proof presence size = %d bytes", trie.MustSize(proof))

			l := lenStats[lenP]
			lenStats[lenP] = l + 1
			sz := size100Stats[sizeP100]
			size100Stats[sizeP100] = sz + 1
		}
		for _, s := range delKeys {
			proofPath := tr.ProofPath([]byte(s))
			proof := CommitmentLogic.Proof(proofPath)
			err := proof.Validate(rootC)
			require.NoError(t, err)
			require.True(t, proof.MustIsProofOfAbsence())
			//t.Logf("key: '%s', proof absence lenPlus1: %d", s, lenPlus1(proofPath.Path))
			t.Logf("proof absence size = %d bytes", trie.MustSize(proof))
		}
		for i := 0; i < 5000; i++ {
			if i < 10 {
				t.Logf("len[%d] = %d", i, lenStats[i])
			}
			if size100Stats[i] != 0 {
				t.Logf("size[%d] = %d", i*100, size100Stats[i])
			}
		}
	})
}
