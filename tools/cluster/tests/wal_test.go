package tests

import (
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

func TestWriteToWAL(t *testing.T) {
	e := setupWithNoChain(t, waspClusterOpts{nNodes: 1})

	chain, err := e.Clu.DeployDefaultChain()
	require.NoError(t, err)
	require.NoError(t, err)

	walDir := walDirFromDataPath(e.Clu.DataPath, chain.ChainID.String())
	require.True(t, walDirectoryCreated(walDir))

	nodeIndex := 0
	blockIndex, _ := chain.BlockIndex(nodeIndex)
	checkCreatedFilenameMatchesBlockIndex(t, walDir, blockIndex)

	segName := latestSegName(walDir)
	segPath := path.Join(walDir, segName)
	blockBytes := getBytesFromSegment(t, segPath)
	block, err := state.BlockFromBytes(blockBytes)
	require.NoError(t, err)
	require.EqualValues(t, blockIndex, block.BlockIndex())

	v, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID, blocklog.Contract.Hname(), blocklog.ViewGetBlockInfo.Name,
		dict.Dict{
			blocklog.ParamBlockIndex: codec.EncodeUint32(blockIndex),
		})
	require.NoError(t, err)

	blockInfo, err := blocklog.BlockInfoFromBytes(blockIndex, v.MustGet(blocklog.ParamBlockInfo))
	require.NoError(t, err)

	require.EqualValues(t, blockInfo.BlockIndex, block.BlockIndex())
	require.EqualValues(t, blockInfo.Timestamp, block.Timestamp())
	require.True(t, state.EqualCommitments(blockInfo.PreviousL1Commitment.TrieRoot, block.PreviousL1Commitment().TrieRoot))
}

func walDirectoryCreated(walDir string) bool {
	_, err := os.Stat(walDir)
	return !os.IsNotExist(err)
}

func walDirFromDataPath(dataPath, chainID string) string {
	return path.Join(dataPath, "wasp0", "wal", chainID)
}

func checkCreatedFilenameMatchesBlockIndex(t *testing.T, walDir string, blockIndex uint32) {
	latestSegmentName := latestSegName(walDir)
	index, _ := strconv.ParseUint(latestSegmentName, 10, 32)
	t.Logf("Index: %d", index)
	require.EqualValues(t, blockIndex, uint32(index))
}

func latestSegName(walDir string) string {
	files, _ := os.ReadDir(walDir)
	return files[len(files)-1].Name()
}

func getBytesFromSegment(t *testing.T, segPath string) []byte {
	data, err := os.ReadFile(segPath)
	require.NoError(t, err)
	return data
}
