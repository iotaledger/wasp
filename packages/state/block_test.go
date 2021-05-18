package state

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBlockBasic(t *testing.T) {
	t.Run("fail no arguments", func(t *testing.T) {
		_, err := NewBlock()
		require.Error(t, err)
	})
	t.Run("fail no state index", func(t *testing.T) {
		su := NewStateUpdate()
		_, err := NewBlock(su)
		require.Error(t, err)
	})
	t.Run("ok block index", func(t *testing.T) {
		su := NewStateUpdateWithBlockIndexMutation(42)
		b1, err := NewBlock(su)
		require.NoError(t, err)
		require.EqualValues(t, 42, b1.BlockIndex())
		require.True(t, b1.Timestamp().IsZero())
	})
	t.Run("with timestamp", func(t *testing.T) {
		nowis := time.Now()
		su := NewStateUpdateWithBlockIndexMutation(42, nowis)
		b1, err := NewBlock(su)
		require.NoError(t, err)
		require.EqualValues(t, 42, b1.BlockIndex())
		require.True(t, nowis.Equal(b1.Timestamp()))
	})
	t.Run("several state updates", func(t *testing.T) {
		nowis := time.Now()
		su1 := NewStateUpdateWithBlockIndexMutation(42, nowis)
		su2 := NewStateUpdateWithBlockIndexMutation(10)
		b1, err := NewBlock(su1, su2)
		require.NoError(t, err)
		require.EqualValues(t, 10, b1.BlockIndex())
		require.True(t, nowis.Equal(b1.Timestamp()))
	})
}

func TestBatches(t *testing.T) {
	suBlock := NewStateUpdateWithBlockIndexMutation(2)
	su1 := NewStateUpdate()
	su2 := NewStateUpdate()

	block1, err := NewBlock(suBlock, su1, su2)
	require.NoError(t, err)
	assert.EqualValues(t, 3, block1.Size())
	assert.EqualValues(t, 2, block1.BlockIndex())
	assert.True(t, block1.Timestamp().IsZero())

	block1Bin := block1.Bytes()
	block2, err := BlockFromBytes(block1Bin)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, block2.Size())
	assert.EqualValues(t, 2, block2.BlockIndex())
	assert.EqualValues(t, block1Bin, block2.Bytes())
	assert.EqualValues(t, block1.EssenceBytes(), block2.EssenceBytes())

	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 1"))
	outID := ledgerstate.NewOutputID(txid1, 0)
	block1.SetApprovingOutputID(outID)
	assert.EqualValues(t, block1.EssenceBytes(), block2.EssenceBytes())

	block2.SetApprovingOutputID(outID)
	assert.EqualValues(t, block1.EssenceBytes(), block2.EssenceBytes())
	assert.EqualValues(t, block1.Bytes(), block2.Bytes())

	assert.EqualValues(t, util.GetHashValue(block1), util.GetHashValue(block2))
}

func TestOriginBlock(t *testing.T) {
	txid1 := ledgerstate.TransactionID{}
	outID1 := ledgerstate.NewOutputID(txid1, 0)
	txid2 := ledgerstate.TransactionID(hashing.RandomHash(nil))
	outID2 := ledgerstate.NewOutputID(txid1, 0)
	require.NotEqualValues(t, txid1, txid2)
	b := NewOriginBlock()
	b1 := NewOriginBlock()
	b1.SetApprovingOutputID(outID1)
	b2 := NewOriginBlock()
	b2.SetApprovingOutputID(outID2)
	require.EqualValues(t, b1.EssenceBytes(), b2.EssenceBytes())

	require.EqualValues(t, 0, b.BlockIndex())
	require.EqualValues(t, 0, b1.BlockIndex())
	require.EqualValues(t, 0, b2.BlockIndex())

	require.True(t, b.Timestamp().IsZero())
	require.True(t, b1.Timestamp().IsZero())
	require.True(t, b2.Timestamp().IsZero())
}
