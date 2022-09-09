package state

import (
	"encoding/hex"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/stretchr/testify/require"
)

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

func genDifferent() []string {
	orig := genRnd4()
	// filter different
	unique := make(map[string]bool)
	for _, s := range orig {
		unique[s] = true
	}
	ret := make([]string, 0)
	for s := range unique {
		ret = append(ret, s)
	}
	return ret
}

func genRndBlocks(start, num int) []Block {
	strs := genDifferent()
	blocks := make([]Block, num)
	millis := rand.Int63()
	const numMutations = 20
	for blkNum := range blocks {
		vc := RandL1Commitment()
		upd := NewStateUpdateWithBlockLogValues(uint32(blkNum+start), time.UnixMilli(millis+int64(blkNum+100)), vc)
		for i := 0; i < numMutations; i++ {
			s := "1111" + strs[rand.Intn(len(strs))]
			if rand.Intn(1000) < 100 {
				upd.Mutations().Del(kv.Key(s))
			} else {
				upd.Mutations().Set(kv.Key(s), []byte(s))
			}
		}
		blocks[blkNum], _ = newBlock(upd.Mutations())
	}
	return blocks
}

func TestRnd(t *testing.T) {
	chainID := isc.RandomChainID()

	const numBlocks = 100
	const numRepeat = 100
	cs := make([]trie.VCommitment, 0)
	blocks := genRndBlocks(2, numBlocks)
	//for bn, blk := range blocks {
	//	t.Logf("--------- #%d\nDELS: %v", bn,
	//		blk.(*blockImpl).stateUpdate.mutations.Dels)
	//}

	// blocks := genBlocks(2, numBlocks)
	t.Logf("num blocks: %d", len(blocks))
	upd1 := NewStateUpdateWithBlockLogValues(1, time.UnixMilli(0), RandL1Commitment())
	var exists bool
	store := make([]kvstore.KVStore, numRepeat)
	rndCommits := make([][]bool, numRepeat)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range rndCommits {
		rndCommits[i] = make([]bool, numBlocks)
		for j := range rndCommits[i] {
			rndCommits[i][j] = rng.Intn(1000) < 100
		}
	}
	// var badBlock int
	// var badKey kv.Key
	var round int
	for round = 0; round < numRepeat; round++ {
		t.Logf("------------------ round: %d", round)
		store[round] = mapdb.NewMapDB()
		vs, err := CreateOriginState(store[round], chainID)
		require.NoError(t, err)
		vs.ApplyStateUpdate(upd1)
		vs.Commit()
		c1 := trie.RootCommitment(vs.TrieNodeStore())
		err = vs.Save()
		require.NoError(t, err)
		c2 := trie.RootCommitment(vs.TrieNodeStore())
		require.True(t, EqualCommitments(c1, c2))
		for bn, b := range blocks {
			require.EqualValues(t, vs.BlockIndex()+1, b.BlockIndex())
			err = vs.ApplyBlock(b)
			require.NoError(t, err)
			if rndCommits[round][bn] {
				// t.Logf("           commit at block: #%d", bn)
				err = vs.Save()
				require.NoError(t, err)
				cc1 := trie.RootCommitment(vs.TrieNodeStore())
				vs, exists, err = LoadSolidState(store[round], chainID)
				require.NoError(t, err)
				require.True(t, exists)
				cc2 := trie.RootCommitment(vs.TrieNodeStore())

				diff := vs.ReconcileTrie()
				if len(diff) > 0 {
					t.Logf("============== reconcile failed: %v", diff)
				}

				require.True(t, EqualCommitments(cc1, cc2))
			}
		}
		vs.Commit()
		c1 = trie.RootCommitment(vs.TrieNodeStore())
		err = vs.Save()
		require.NoError(t, err)
		c2 = trie.RootCommitment(vs.TrieNodeStore())
		require.True(t, EqualCommitments(c1, c2))

		vstmp, exists, err := LoadSolidState(store[round], chainID)
		require.NoError(t, err)
		require.True(t, exists)
		diff := vstmp.ReconcileTrie()
		if len(diff) > 0 {
			t.Logf("============== reconcile failed: %v", diff)
		}

		cs = append(cs, trie.RootCommitment(vs.TrieNodeStore()))
		if round > 0 {
			require.True(t, EqualCommitments(cs[round-1], cs[round]))
		}
	}
}

func readBlocks(t *testing.T, dir string) ([]Block, []trie.VCommitment, []hashing.HashValue) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, info.Name())
		// t.Logf("-- %s", info.Name())
		return nil
	})
	if err != nil {
		t.Logf("filepath.Walk error: %v", err)
		return nil, nil, nil
	}
	retBlocks := make([]Block, len(files))
	retCommitments := make([]trie.VCommitment, len(files))
	retBlockHashes := make([]hashing.HashValue, len(files))

	for _, fn := range files {
		part := strings.Split(fn, ".")
		require.EqualValues(t, 4, len(part))
		require.EqualValues(t, "mut", part[3])
		n, err := strconv.Atoi(part[0])
		require.NoError(t, err)
		// expected all numbers from 1 to len()-1
		require.True(t, n >= 1 && n <= len(retBlocks) && retBlocks[n-1] == nil)
		vcbin, err := hex.DecodeString(part[1])
		require.NoError(t, err)
		stateCommitment, err := VCommitmentFromBytes(vcbin)
		require.NoError(t, err)
		retCommitments[n-1] = stateCommitment
		blockHash, err := hashing.HashValueFromHex(part[2])
		require.NoError(t, err)
		retBlockHashes[n-1] = blockHash

		blockBin, err := os.ReadFile(filepath.Join(dir, fn))
		require.NoError(t, err)
		blk, err := BlockFromBytes(blockBin)
		require.NoError(t, err)
		retBlocks[n-1] = blk

		require.EqualValues(t, blockHash, hashing.HashData(blockBin))
		require.EqualValues(t, n, blk.BlockIndex())
	}

	return retBlocks, retCommitments, retBlockHashes
}

const directory = "testdata/test1"

func TestBlockDeterminism(t *testing.T) {
	blocks, stateCommitments, _ := readBlocks(t, directory)
	if len(blocks) == 0 {
		t.Logf("no test data has been found in %s", directory)
		return
	}

	chainID := testmisc.RandChainID()
	runRound := func(saveYN func(i uint16) bool) {
		vs, err := CreateOriginState(mapdb.NewMapDB(), chainID)
		require.NoError(t, err)
		require.True(t, EqualCommitments(trie.RootCommitment(vs.TrieNodeStore()), OriginStateCommitment()))
		require.EqualValues(t, calcOriginStateCommitment(), trie.RootCommitment(vs.TrieNodeStore()))

		commits := make([]int, 0)
		blockToSave := make([]Block, 0, len(blocks))
		for i, blk := range blocks {
			err = vs.ApplyBlock(blk)
			require.NoError(t, err)
			blockToSave = append(blockToSave, blk)

			if saveYN(uint16(i)) {
				commits = append(commits, i)
				err = vs.Save(blockToSave...)
				require.NoError(t, err)

				require.True(t, EqualCommitments(stateCommitments[i], trie.RootCommitment(vs.TrieNodeStore())))
				blockToSave = blockToSave[:0]
			}
		}
		err = vs.Save()
		t.Logf("    committed at %+v", commits)
		require.NoError(t, err)
		require.True(t, EqualCommitments(stateCommitments[len(stateCommitments)-1], trie.RootCommitment(vs.TrieNodeStore())))
	}
	runRound(func(i uint16) bool {
		return true
	})

	// all combinations
	for m := uint16(0); m < uint16(0x1)<<len(blocks); m++ {
		t.Logf("%x", m)
		runRound(func(i uint16) bool {
			mask := uint16(0x1) << i
			return m&mask != 0
		})
	}
}
