package state

import (
	"testing"
	"time"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockBasic(t *testing.T) {
	t.Run("fail no state index", func(t *testing.T) {
		su := NewStateUpdate()
		_, err := newBlock(su.Mutations())
		require.Error(t, err)
	})
	t.Run("ok block index", func(t *testing.T) {
		h := testmisc.RandVectorCommitment()
		su := NewStateUpdateWithBlockLogValues(42, time.Time{}, h)
		b1, err := newBlock(su.Mutations())
		require.NoError(t, err)
		require.EqualValues(t, 42, b1.BlockIndex())
		require.True(t, b1.Timestamp().IsZero())
		require.True(t, trie.EqualCommitments(h, b1.PreviousStateCommitment(CommitmentModel)))
	})
	t.Run("with timestamp", func(t *testing.T) {
		currentTime := time.Now()
		ph := testmisc.RandVectorCommitment()
		su := NewStateUpdateWithBlockLogValues(42, currentTime, ph)
		b1, err := newBlock(su.Mutations())
		require.NoError(t, err)
		require.EqualValues(t, 42, b1.BlockIndex())
		require.True(t, currentTime.Equal(b1.Timestamp()))
		require.EqualValues(t, ph, b1.PreviousStateCommitment(CommitmentModel))
	})
}

func TestBatches(t *testing.T) {
	suBlock := NewStateUpdateWithBlockLogValues(2, time.Time{}, testmisc.RandVectorCommitment())

	block1, err := newBlock(suBlock.Mutations())
	require.NoError(t, err)
	assert.EqualValues(t, 2, block1.BlockIndex())
	assert.True(t, block1.Timestamp().IsZero())

	block1Bin := block1.Bytes()
	block2, err := BlockFromBytes(block1Bin)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, block2.BlockIndex())
	assert.EqualValues(t, block1Bin, block2.Bytes())
	assert.EqualValues(t, block1.EssenceBytes(), block2.EssenceBytes())

	outID := tpkg.RandOutputID(0).UTXOInput()
	block1.SetApprovingOutputID(outID)
	assert.EqualValues(t, block1.EssenceBytes(), block2.EssenceBytes())

	block2.SetApprovingOutputID(outID)
	assert.EqualValues(t, block1.EssenceBytes(), block2.EssenceBytes())
	assert.EqualValues(t, block1.Bytes(), block2.Bytes())

	assert.EqualValues(t, util.GetHashValue(block1), util.GetHashValue(block2))
}
