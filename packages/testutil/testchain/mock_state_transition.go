package testchain

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
)

type MockedStateTransition struct {
	t           *testing.T
	chainKey    *cryptolib.KeyPair
	onNextState func(virtualState state.VirtualStateAccess, tx *iotago.Transaction)
	onVMResult  func(virtualState state.VirtualStateAccess, tx *iotago.TransactionEssence)
}

func NewMockedStateTransition(t *testing.T, chainKey *cryptolib.KeyPair) *MockedStateTransition {
	return &MockedStateTransition{
		t:        t,
		chainKey: chainKey,
	}
}

func (c *MockedStateTransition) NextState(vs state.VirtualStateAccess, chainOutput *iotago.AliasOutput, ts time.Time, reqs ...iscp.Calldata) {
	panic("TODO implement")
	// if c.chainKey != nil {
	// 	require.True(c.t, chainOutput.GetStateAddress().Equals(iotago.NewED25519Address(c.chainKey.PublicKey)))
	// }

	// nextvs := vs.Copy()
	// prevBlockIndex := vs.BlockIndex()
	// counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")

	// counterBin, err := nextvs.KVStore().Get(counterKey)
	// require.NoError(c.t, err)

	// counter, err := codec.DecodeUint64(counterBin, 0)
	// require.NoError(c.t, err)

	// suBlockIndex := state.NewStateUpdateWithBlocklogValues(prevBlockIndex+1, time.Time{}, vs.StateCommitment())

	// suCounter := state.NewStateUpdate()
	// counterBin = codec.EncodeUint64(counter + 1)
	// suCounter.Mutations().Set(counterKey, counterBin)

	// suReqs := state.NewStateUpdate()
	// for i, req := range reqs {
	// 	key := kv.Key(blocklog.NewRequestLookupKey(vs.BlockIndex()+1, uint16(i)).Bytes())
	// 	suReqs.Mutations().Set(key, req.ID().Bytes())
	// }

	// nextvs.ApplyStateUpdates(suBlockIndex, suCounter, suReqs)
	// require.EqualValues(c.t, prevBlockIndex+1, nextvs.BlockIndex())

	// nextStateHash := nextvs.StateCommitment()

	// txBuilder := utxoutil.NewBuilder(chainOutput).WithTimestamp(ts)
	// err = txBuilder.AddAliasOutputAsRemainder(chainOutput.GetAliasAddress(), nextStateHash[:])
	// require.NoError(c.t, err)

	// if c.chainKey != nil {
	// 	tx, err := txBuilder.BuildWithED25519(c.chainKey)
	// 	require.NoError(c.t, err)
	// 	c.onNextState(nextvs, tx)
	// } else {
	// 	tx, _, err := txBuilder.BuildEssence()
	// 	require.NoError(c.t, err)
	// 	c.onVMResult(nextvs, tx)
	// }
}

func (c *MockedStateTransition) OnNextState(f func(virtualStats state.VirtualStateAccess, tx *iotago.Transaction)) {
	c.onNextState = f
}

func (c *MockedStateTransition) OnVMResult(f func(virtualStats state.VirtualStateAccess, tx *iotago.TransactionEssence)) {
	c.onVMResult = f
}
