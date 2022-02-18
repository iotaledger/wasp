package testchain

import (
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"

	//"github.com/iotaledger/wasp/packages/transaction"
	//"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/stretchr/testify/require"
	//"golang.org/x/crypto/blake2b"
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

func (c *MockedStateTransition) NextState(vs state.VirtualStateAccess, chainOutput *iscp.AliasOutputWithID, ts time.Time /*, reqs ...iscp.Calldata*/) {
	if c.chainKey != nil {
		require.True(c.t, chainOutput.GetStateAddress().Equal(c.chainKey.GetPublicKey().AsEd25519Address()))
	}

	nextvs := vs.Copy()
	prevBlockIndex := vs.BlockIndex()
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")

	counterBin, err := nextvs.KVStore().Get(counterKey)
	require.NoError(c.t, err)

	counter, err := codec.DecodeUint64(counterBin, 0)
	require.NoError(c.t, err)

	suBlockIndex := state.NewStateUpdateWithBlocklogValues(prevBlockIndex+1, time.Time{}, vs.StateCommitment())

	suCounter := state.NewStateUpdate()
	counterBin = codec.EncodeUint64(counter + 1)
	suCounter.Mutations().Set(counterKey, counterBin)

	/*suReqs := state.NewStateUpdate()
	for i, req := range reqs {
		key := kv.Key(blocklog.NewRequestLookupKey(vs.BlockIndex()+1, uint16(i)).Bytes())
		suReqs.Mutations().Set(key, req.ID().Bytes())
	}*/

	nextvs.ApplyStateUpdates(suBlockIndex, suCounter /*, suReqs*/)
	require.EqualValues(c.t, prevBlockIndex+1, nextvs.BlockIndex())

	//nextStateHash := nextvs.StateCommitment()

	consumedOutput := chainOutput.GetAliasOutput()
	/*var aliasID iotago.AliasID
	if consumedOutput.AliasEmpty() {
		id := chainOutput.OutputID()
		hash, err := blake2b.New(160, id[:])
		require.NoError(c.t, err)
		hashBytes := hash.Sum([]byte{})
		copy(aliasID[:], hashBytes[:iotago.AliasIDLength])
	} else {
		aliasID = consumedOutput.AliasID
	}*/
	aliasID := consumedOutput.AliasID
	inputs := iotago.OutputIDs{chainOutput.OutputID()}
	txEssence := &iotago.TransactionEssence{
		NetworkID: parameters.NetworkID,
		Inputs:    inputs.UTXOInputs(),
		Outputs: []iotago.Output{
			&iotago.AliasOutput{
				Amount:         consumedOutput.Amount,
				NativeTokens:   consumedOutput.NativeTokens,
				AliasID:        aliasID,
				StateIndex:     consumedOutput.StateIndex + 1,
				StateMetadata:  nextvs.StateCommitment().Bytes(),
				FoundryCounter: consumedOutput.FoundryCounter,
				Conditions:     consumedOutput.Conditions,
				Blocks:         consumedOutput.Blocks,
			},
		},
		Payload: nil,
	}
	signatures, err := txEssence.Sign(
		iotago.Outputs{chainOutput.GetAliasOutput()}.MustCommitment(),
		c.chainKey.GetPrivateKey().AddressKeys(chainOutput.GetStateAddress()),
	)
	require.NoError(c.t, err)
	tx := &iotago.Transaction{
		Essence:      txEssence,
		UnlockBlocks: []iotago.UnlockBlock{&iotago.SignatureUnlockBlock{Signature: signatures[0]}},
	}

	/*tx, err = transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    cryptolib.HiveKeyPairToCryptolibKeyPair(*c.chainKey),
		UnspentOutputs:   []iotago.Output{chainOutput.GetAliasOutput()},
		UnspentOutputIDs: []*iotago.UTXOInput{chainOutput.ID()},
		Requests:         []*iscp.RequestParameters{}, // TODO
		//RentStructure:                 *iotago.RentStructure
		//DisableAutoAdjustDustDeposit bool // if true, the minimal dust deposit won't be adjusted automatically
	})
	require.NoError(c.t, err)*/

	/*txBuilder := utxoutil.NewBuilder(chainOutput).WithTimestamp(ts)
	err = txBuilder.AddAliasOutputAsRemainder(chainOutput.GetAliasAddress(), nextStateHash[:])
	require.NoError(c.t, err)*/

	if c.chainKey != nil {
		//tx, err := txBuilder.BuildWithED25519(c.chainKey)
		//require.NoError(c.t, err)
		c.onNextState(nextvs, tx)
	} else {
		//tx, _, err := txBuilder.BuildEssence()
		//require.NoError(c.t, err)
		c.onVMResult(nextvs, tx.Essence)
	}
}

func (c *MockedStateTransition) OnNextState(f func(virtualStats state.VirtualStateAccess, tx *iotago.Transaction)) {
	c.onNextState = f
}

func (c *MockedStateTransition) OnVMResult(f func(virtualStats state.VirtualStateAccess, tx *iotago.TransactionEssence)) {
	c.onVMResult = f
}
