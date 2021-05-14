package testchain

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/stretchr/testify/require"
)

type MockedStateTransition struct {
	t           *testing.T
	chainKey    *ed25519.KeyPair
	onNextState func(virtualState state.VirtualState, tx *ledgerstate.Transaction)
	onVMResult  func(virtualState state.VirtualState, tx *ledgerstate.TransactionEssence)
}

func NewMockedStateTransition(t *testing.T, chainKey *ed25519.KeyPair) *MockedStateTransition {
	return &MockedStateTransition{
		t:        t,
		chainKey: chainKey,
	}
}

func (c *MockedStateTransition) NextState(virtualState state.VirtualState, chainOutput *ledgerstate.AliasOutput, ts time.Time, reqs ...coretypes.Request) {
	if c.chainKey != nil {
		require.True(c.t, chainOutput.GetStateAddress().Equals(ledgerstate.NewED25519Address(c.chainKey.PublicKey)))
	}

	nextVirtualState := virtualState.Clone()
	prevBlockIndex := virtualState.BlockIndex()
	counterBin, err := nextVirtualState.KVStore().Get("counter")
	require.NoError(c.t, err)
	counter, _, err := codec.DecodeUint64(counterBin)
	require.NoError(c.t, err)

	sus := make([]state.StateUpdate, 0, len(reqs)+2)

	suBlockIndex := state.NewStateUpdateWithBlockIndexMutation(prevBlockIndex + 1)
	sus = append(sus, suBlockIndex)

	suCounter := state.NewStateUpdate()
	counterBin = codec.EncodeUint64(counter + 1)
	suCounter.Mutations().Set("counter", counterBin)
	sus = append(sus, suCounter)

	for _, req := range reqs {
		sureq := state.NewStateUpdate()
		sureq.Mutations().Set(kv.Key(req.ID().Bytes()), counterBin)
		sus = append(sus, sureq)
	}

	nextVirtualState.ApplyStateUpdates(sus...)
	require.EqualValues(c.t, prevBlockIndex+1, nextVirtualState.BlockIndex())

	nextStateHash := nextVirtualState.Hash()

	txBuilder := utxoutil.NewBuilder(chainOutput).WithTimestamp(ts)
	err = txBuilder.AddAliasOutputAsRemainder(chainOutput.GetAliasAddress(), nextStateHash[:])
	require.NoError(c.t, err)

	if c.chainKey != nil {
		tx, err := txBuilder.BuildWithED25519(c.chainKey)
		require.NoError(c.t, err)
		c.onNextState(nextVirtualState, tx)
	} else {
		tx, _, err := txBuilder.BuildEssence()
		require.NoError(c.t, err)
		c.onVMResult(nextVirtualState, tx)
	}
}

func (c *MockedStateTransition) OnNextState(f func(virtualStats state.VirtualState, tx *ledgerstate.Transaction)) {
	c.onNextState = f
}

func (c *MockedStateTransition) OnVMResult(f func(virtualStats state.VirtualState, tx *ledgerstate.TransactionEssence)) {
	c.onVMResult = f
}
