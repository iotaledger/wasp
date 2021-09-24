package state

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockBasic(t *testing.T) {
	t.Run("fail no state index", func(t *testing.T) {
		su := NewStateUpdate()
		_, err := newBlock(su)
		require.Error(t, err)
	})
	t.Run("ok block index", func(t *testing.T) {
		su := NewStateUpdateWithBlocklogValues(42, time.Time{}, hashing.NilHash)
		b1, err := newBlock(su)
		require.NoError(t, err)
		require.EqualValues(t, 42, b1.BlockIndex())
		require.True(t, b1.Timestamp().IsZero())
		require.EqualValues(t, hashing.NilHash, b1.PreviousStateHash())
	})
	t.Run("with timestamp", func(t *testing.T) {
		nowis := time.Now()
		ph := hashing.HashStrings("dummy-dummy")
		su := NewStateUpdateWithBlocklogValues(42, nowis, ph)
		b1, err := newBlock(su)
		require.NoError(t, err)
		require.EqualValues(t, 42, b1.BlockIndex())
		require.True(t, nowis.Equal(b1.Timestamp()))
		require.EqualValues(t, ph, b1.PreviousStateHash())
	})
}

func TestBatches(t *testing.T) {
	suBlock := NewStateUpdateWithBlocklogValues(2, time.Time{}, hashing.NilHash)

	block1, err := newBlock(suBlock)
	require.NoError(t, err)
	assert.EqualValues(t, 2, block1.BlockIndex())
	assert.True(t, block1.Timestamp().IsZero())

	block1Bin := block1.Bytes()
	block2, err := BlockFromBytes(block1Bin)
	assert.NoError(t, err)
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
	b := newOriginBlock()
	b1 := newOriginBlock()
	b1.SetApprovingOutputID(outID1)
	b2 := newOriginBlock()
	b2.SetApprovingOutputID(outID2)
	require.EqualValues(t, b1.EssenceBytes(), b2.EssenceBytes())

	require.EqualValues(t, 0, b.BlockIndex())
	require.EqualValues(t, 0, b1.BlockIndex())
	require.EqualValues(t, 0, b2.BlockIndex())

	require.True(t, b.Timestamp().IsZero())
	require.True(t, b1.Timestamp().IsZero())
	require.True(t, b2.Timestamp().IsZero())

	require.EqualValues(t, hashing.NilHash, b.PreviousStateHash())
	require.EqualValues(t, hashing.NilHash, b1.PreviousStateHash())
	require.EqualValues(t, hashing.NilHash, b2.PreviousStateHash())
}
