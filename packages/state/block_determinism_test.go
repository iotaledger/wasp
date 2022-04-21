package state

import (
	"encoding/hex"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func readBlocks(t *testing.T, dir string) ([]Block, []trie.VCommitment, []hashing.HashValue) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, info.Name())
		//t.Logf("-- %s", info.Name())
		return nil
	})
	require.NoError(t, err)
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
		stateCommitment, err := CommitmentModel.VectorCommitmentFromBytes(vcbin)
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

	chainID := testmisc.RandChainID()
	runRound := func(saveYN func(i uint16) bool) {
		vs, err := CreateOriginState(mapdb.NewMapDB(), chainID)
		require.NoError(t, err)
		require.True(t, trie.EqualCommitments(trie.RootCommitment(vs.TrieNodeStore()), OriginStateCommitment()))
		require.EqualValues(t, calcOriginStateHash(), trie.RootCommitment(vs.TrieNodeStore()))

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

				require.True(t, trie.EqualCommitments(stateCommitments[i], trie.RootCommitment(vs.TrieNodeStore())))
				blockToSave = blockToSave[:0]
			}
		}
		err = vs.Save()
		t.Logf("    committed at %+v", commits)
		require.NoError(t, err)
		require.True(t, trie.EqualCommitments(stateCommitments[len(stateCommitments)-1], trie.RootCommitment(vs.TrieNodeStore())))
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
