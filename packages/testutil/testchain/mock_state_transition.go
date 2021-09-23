package testchain

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"

	"github.com/iotaledger/wasp/packages/iscp"

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

func (c *MockedStateTransition) NextState(vs state.VirtualState, chainOutput *ledgerstate.AliasOutput, ts time.Time, reqs ...iscp.Request) {
	if c.chainKey != nil {
		require.True(c.t, chainOutput.GetStateAddress().Equals(ledgerstate.NewED25519Address(c.chainKey.PublicKey)))
	}

	nextvs := vs.Clone()
	prevBlockIndex := vs.BlockIndex()
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")

	counterBin, err := nextvs.KVStore().Get(counterKey)
	require.NoError(c.t, err)

	counter, _, err := codec.DecodeUint64(counterBin)
	require.NoError(c.t, err)

	suBlockIndex := state.NewStateUpdateWithBlocklogValues(prevBlockIndex+1, time.Time{}, vs.StateCommitment())

	suCounter := state.NewStateUpdate()
	counterBin = codec.EncodeUint64(counter + 1)
	suCounter.Mutations().Set(counterKey, counterBin)

	suReqs := state.NewStateUpdate()
	for i, req := range reqs {
		key := kv.Key(blocklog.NewRequestLookupKey(vs.BlockIndex()+1, uint16(i)).Bytes())
		suReqs.Mutations().Set(key, req.ID().Bytes())
	}

	nextvs.ApplyStateUpdates(suBlockIndex, suCounter, suReqs)
	require.EqualValues(c.t, prevBlockIndex+1, nextvs.BlockIndex())

	nextStateHash := nextvs.StateCommitment()

	txBuilder := utxoutil.NewBuilder(chainOutput).WithTimestamp(ts)
	err = txBuilder.AddAliasOutputAsRemainder(chainOutput.GetAliasAddress(), nextStateHash[:])
	require.NoError(c.t, err)

	if c.chainKey != nil {
		tx, err := txBuilder.BuildWithED25519(c.chainKey)
		require.NoError(c.t, err)
		c.onNextState(nextvs, tx)
	} else {
		tx, _, err := txBuilder.BuildEssence()
		require.NoError(c.t, err)
		c.onVMResult(nextvs, tx)
	}
}

func (c *MockedStateTransition) OnNextState(f func(virtualStats state.VirtualState, tx *ledgerstate.Transaction)) {
	c.onNextState = f
}

func (c *MockedStateTransition) OnVMResult(f func(virtualStats state.VirtualState, tx *ledgerstate.TransactionEssence)) {
	c.onVMResult = f
}
