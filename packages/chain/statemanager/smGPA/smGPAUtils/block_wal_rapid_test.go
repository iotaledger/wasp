package smGPAUtils

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

//const constTestFolder = "basicWALTest"

type blockWALTestSM struct { // State machine for block WAL property based Rapid tests
	bw            BlockWAL
	chainID       *isc.ChainID
	ao            *isc.AliasOutputWithID
	vs            state.VirtualStateAccess
	blocks        map[state.BlockHash]state.Block
	blocksMoved   []state.BlockHash
	blocksDamaged []state.BlockHash
	log           *logger.Logger
}

func (bwtsmT *blockWALTestSM) Init(t *rapid.T) {
	var err error
	bwtsmT.chainID, bwtsmT.ao, bwtsmT.vs = GetOriginState(t)
	bwtsmT.log = testlogger.NewLogger(t)
	bwtsmT.bw, err = NewBlockWAL(constTestFolder, bwtsmT.chainID, bwtsmT.log)
	require.NoError(t, err)
	bwtsmT.blocks = make(map[state.BlockHash]state.Block)
	bwtsmT.blocksMoved = make([]state.BlockHash, 0)
	bwtsmT.blocksDamaged = make([]state.BlockHash, 0)
}

func (bwtsmT *blockWALTestSM) Cleanup() {
	bwtsmT.log.Sync()
	os.RemoveAll(constTestFolder)
}

func (bwtsmT *blockWALTestSM) Check(t *rapid.T) {
	bwtsmT.invariantAllWrittenBlocksExist(t)
}

func (bwtsmT *blockWALTestSM) WriteBlock(t *rapid.T) {
	block, aliasOutput, virtualState := GetNextState(t, bwtsmT.vs, bwtsmT.ao)
	bwtsmT.blocks[block.GetHash()] = block
	bwtsmT.ao = aliasOutput
	bwtsmT.vs = virtualState
	err := bwtsmT.bw.Write(block)
	require.NoError(t, err)
	t.Logf("Block %s written", block.GetHash())
}

// Correct the damaged block file
func (bwtsmT *blockWALTestSM) ReWriteBlock(t *rapid.T) {
	var takeFrom int
	if len(bwtsmT.blocksMoved) == 0 {
		if len(bwtsmT.blocksDamaged) == 0 {
			t.Skip()
		} else {
			takeFrom = 1
		}
	} else if len(bwtsmT.blocksDamaged) == 0 {
		takeFrom = 0
	} else {
		takeFrom = rapid.IntRange(0, 1).Example()
	}
	var blockHash state.BlockHash
	if takeFrom == 0 {
		blockHash = rapid.SampledFrom(bwtsmT.blocksMoved).Example()
	} else {
		require.Equal(t, 1, takeFrom)
		blockHash = rapid.SampledFrom(bwtsmT.blocksDamaged).Example()
	}
	block, ok := bwtsmT.blocks[blockHash]
	require.True(t, ok)
	err := bwtsmT.bw.Write(block)
	require.NoError(t, err)
	if takeFrom == 0 {
		bwtsmT.blocksMoved = DeleteBlockHash(blockHash, bwtsmT.blocksMoved)
	} else {
		bwtsmT.blocksDamaged = DeleteBlockHash(blockHash, bwtsmT.blocksDamaged)
	}
	t.Logf("Block %s rewritten", blockHash)
}

// Damage the block by overwriting its file with bytes of other block
func (bwtsmT *blockWALTestSM) MoveBlock(t *rapid.T) {
	blockHashes := bwtsmT.getGoodBlockHashes()
	if len(blockHashes) < 2 {
		t.Skip()
	}
	gen := rapid.SampledFrom(blockHashes)
	blockHashOrig := gen.Example(0)
	blockHashToDamage := gen.Example(1)
	if blockHashOrig.Equals(blockHashToDamage) {
		t.Skip()
	}
	fileOrigPath := bwtsmT.pathFromHash(blockHashOrig)
	fileToDamagePath := bwtsmT.pathFromHash(blockHashToDamage)
	data, err := os.ReadFile(fileOrigPath)
	require.NoError(t, err)
	err = os.WriteFile(fileToDamagePath, data, 0644)
	require.NoError(t, err)
	bwtsmT.blocksMoved = append(bwtsmT.blocksMoved, blockHashToDamage)
	t.Logf("Block %s damaged: block %s written instead", blockHashToDamage, blockHashOrig)
}

// Damage the block by writting random bytes to it
func (bwtsmT *blockWALTestSM) DamageBlock(t *rapid.T) {
	blockHashes := bwtsmT.getGoodBlockHashes()
	if len(blockHashes) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(blockHashes).Example()
	filePath := bwtsmT.pathFromHash(blockHash)
	data := make([]byte, 50)
	_, err := rand.Read(data)
	require.NoError(t, err)
	err = os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)
	bwtsmT.blocksDamaged = append(bwtsmT.blocksDamaged, blockHash)
	t.Logf("Block %s damaged: 50 random bytes written instead", blockHash)
}

func (bwtsmT *blockWALTestSM) ReadGoodBlock(t *rapid.T) {
	blockHashes := bwtsmT.getGoodBlockHashes()
	if len(blockHashes) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(blockHashes).Example()
	block, err := bwtsmT.bw.Read(blockHash)
	require.NoError(t, err)
	require.True(t, block.Equals(bwtsmT.blocks[blockHash]))
	t.Logf("Block %s read", blockHash)
}

func (bwtsmT *blockWALTestSM) ReadMovedBlock(t *rapid.T) {
	if len(bwtsmT.blocksMoved) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bwtsmT.blocksMoved).Example()
	block, err := bwtsmT.bw.Read(blockHash)
	require.NoError(t, err)
	require.False(t, block.Equals(bwtsmT.blocks[blockHash]))
	t.Logf("Moved block %s read", blockHash)
}

func (bwtsmT *blockWALTestSM) ReadDamagedBlock(t *rapid.T) {
	if len(bwtsmT.blocksDamaged) == 0 {
		t.Skip()
	}
	blockHash := rapid.SampledFrom(bwtsmT.blocksDamaged).Example()
	_, err := bwtsmT.bw.Read(blockHash)
	require.Error(t, err)
	t.Logf("Damaged block %s read", blockHash)
}

func (bwtsmT *blockWALTestSM) Restart(t *rapid.T) {
	var err error
	bwtsmT.bw, err = NewBlockWAL(constTestFolder, bwtsmT.chainID, bwtsmT.log)
	require.NoError(t, err)
	t.Logf("Block WAL restarted")
}

func (bwtsmT *blockWALTestSM) getGoodBlockHashes() []state.BlockHash {
	result := make([]state.BlockHash, 0)
	for blockHash, _ := range bwtsmT.blocks { //nolint:gofmt,gofumpt,revive,gosimple
		if !ContainsBlockHash(blockHash, bwtsmT.blocksMoved) && !ContainsBlockHash(blockHash, bwtsmT.blocksDamaged) {
			result = append(result, blockHash)
		}
	}
	return result
}

func (bwtsmT *blockWALTestSM) pathFromHash(blockHash state.BlockHash) string {
	return filepath.Join(constTestFolder, bwtsmT.chainID.String(), fileName(blockHash))
}

func (bwtsmT *blockWALTestSM) invariantAllWrittenBlocksExist(t *rapid.T) {
	for blockHash, _ := range bwtsmT.blocks { //nolint:gofmt,gofumpt,revive,gosimple
		require.True(t, bwtsmT.bw.Contains(blockHash))
	}
}

func TestBlockWALPropBased(t *testing.T) {
	rapid.Check(t, rapid.Run[*blockWALTestSM]())
}
