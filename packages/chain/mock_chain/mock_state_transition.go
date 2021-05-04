package mock_chain

import (
	"testing"

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
}

func NewMockedStateTransition(t *testing.T, chainKey *ed25519.KeyPair) *MockedStateTransition {
	return &MockedStateTransition{
		t:        t,
		chainKey: chainKey,
	}
}

func (c *MockedStateTransition) NextState(virtualState state.VirtualState, chainOutput *ledgerstate.AliasOutput) {
	require.True(c.t, chainOutput.GetStateAddress().Equals(ledgerstate.NewED25519Address(c.chainKey.PublicKey)))

	nextVirtualState := virtualState.Clone()
	counterBin, err := nextVirtualState.KVStore().Get("counter")
	require.NoError(c.t, err)
	counter, _, err := codec.DecodeUint64(counterBin)
	require.NoError(c.t, err)

	su0 := state.NewStateUpdateWithBlockIndexMutation(virtualState.BlockIndex() + 1)
	su1 := state.NewStateUpdate()
	su1.Mutations().Set("counter", codec.EncodeUint64(counter+1))
	nextVirtualState.ApplyStateUpdates(su0, su1)

	nextStateHash := nextVirtualState.Hash()

	txBuilder := utxoutil.NewBuilder(chainOutput)
	err = txBuilder.AddAliasOutputAsRemainder(chainOutput.GetAliasAddress(), nextStateHash[:])
	require.NoError(c.t, err)
	tx, err := txBuilder.BuildWithED25519(c.chainKey)
	require.NoError(c.t, err)

	c.onNextState(nextVirtualState, tx)
}

func (c *MockedStateTransition) OnNextState(f func(virtualStats state.VirtualState, tx *ledgerstate.Transaction)) {
	c.onNextState = f
}
