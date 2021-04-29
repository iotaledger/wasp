package state

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatches(t *testing.T) {
	su1 := NewStateUpdate()
	su2 := NewStateUpdate()

	block1 := NewBlock(2, su1, su2)
	assert.EqualValues(t, 3, block1.Size())
	assert.EqualValues(t, 2, block1.BlockIndex())

	block1Bin := block1.Bytes()
	block2, err := BlockFromBytes(block1Bin)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, block2.Size())
	assert.EqualValues(t, 2, block2.BlockIndex())
	assert.EqualValues(t, block1Bin, block2.Bytes())
	assert.EqualValues(t, block1.EssenceHash(), block2.EssenceHash())

	txid1 := ledgerstate.TransactionID(hashing.HashStrings("test string 1"))
	outID := ledgerstate.NewOutputID(txid1, 0)
	block1.WithApprovingOutputID(outID)
	assert.EqualValues(t, block1.EssenceHash(), block2.EssenceHash())

	block2.WithApprovingOutputID(outID)
	assert.EqualValues(t, block1.EssenceHash(), block2.EssenceHash())
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
	require.EqualValues(t, coreutil.OriginBlockEssenceHashBase58, b.EssenceHash().String())

	b1 := NewOriginBlock().WithApprovingOutputID(outID1)
	b2 := NewOriginBlock().WithApprovingOutputID(outID2)
	require.EqualValues(t, b1.EssenceHash(), b2.EssenceHash())
}
