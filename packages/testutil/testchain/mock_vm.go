package testchain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

type MockedVMRunner struct {
	log *logger.Logger
	t   *testing.T
}

var _ vm.VMRunner = &MockedVMRunner{}

func NewMockedVMRunner(t *testing.T, log *logger.Logger) *MockedVMRunner {
	ret := &MockedVMRunner{
		log: log.Named("vm"),
		t:   t,
	}
	ret.log.Debugf("Mocked VM runner created")
	return ret
}

func (r *MockedVMRunner) Run(task *vm.VMTask) error {
	r.log.Debugf("Mocked VM runner: VM started for trie root %v output %v",
		task.StateDraft.BaseL1Commitment().TrieRoot, isc.OID(task.AnchorOutputID.UTXOInput()))
	draft, block, txEssence, inputsCommitment := nextState(r.t, task.Store, task.AnchorOutput, task.AnchorOutputID, task.TimeAssumption, task.Requests)
	task.StateDraft = draft
	task.RotationAddress = nil
	task.ResultTransactionEssence = txEssence
	task.ResultInputsCommitment = inputsCommitment
	task.Results = make([]*vm.RequestResult, len(task.Requests))
	for i := range task.Results {
		task.Results[i] = &vm.RequestResult{
			Request: task.Requests[i],
			Return:  dict.New(),
			Receipt: &blocklog.RequestReceipt{
				Request: task.Requests[i],
				Error:   nil,
			},
		}
	}
	r.log.Debugf("Mocked VM runner: VM completed; state %v commitment %v received", draft.BlockIndex(), block.TrieRoot())
	return nil
}

func nextState(
	t *testing.T,
	store state.Store,
	consumedOutput *iotago.AliasOutput,
	consumedOutputID iotago.OutputID,
	timeAssumption time.Time,
	reqs []isc.Request,
) (state.StateDraft, state.Block, *iotago.TransactionEssence, []byte) {
	timeAssumption = timeAssumption.Add(time.Duration(len(reqs)) * time.Nanosecond)

	prev, err := state.L1CommitmentFromBytes(consumedOutput.StateMetadata)
	require.NoError(t, err)

	draft, err := store.NewStateDraft(timeAssumption, &prev)
	require.NoError(t, err)

	for i, req := range reqs {
		key := kv.Key(blocklog.NewRequestLookupKey(draft.BlockIndex(), uint16(i)).Bytes())
		draft.Set(key, req.ID().Bytes())
	}

	block := store.ExtractBlock(draft)

	aliasID := consumedOutput.AliasID
	inputs := iotago.OutputIDs{consumedOutputID}
	txEssence := &iotago.TransactionEssence{
		NetworkID: tpkg.TestNetworkID,
		Inputs:    inputs.UTXOInputs(),
		Outputs: []iotago.Output{
			&iotago.AliasOutput{
				Amount:         consumedOutput.Amount,
				NativeTokens:   consumedOutput.NativeTokens,
				AliasID:        aliasID,
				StateIndex:     consumedOutput.StateIndex + 1,
				StateMetadata:  block.L1Commitment().Bytes(),
				FoundryCounter: consumedOutput.FoundryCounter,
				Conditions:     consumedOutput.Conditions,
				Features:       consumedOutput.Features,
			},
		},
		Payload: nil,
	}

	inputsCommitment := iotago.Outputs{consumedOutput}.MustCommitment()

	store.Commit(draft)
	err = store.SetLatest(block.TrieRoot())
	require.NoError(t, err)

	return draft, block, txEssence, inputsCommitment
}

func NextState(
	t *testing.T,
	chainKey *cryptolib.KeyPair,
	store state.Store,
	chainOutput *isc.AliasOutputWithID,
	ts time.Time,
) (state.Block, *iotago.Transaction, *iotago.UTXOInput) {
	if chainKey != nil {
		require.True(t, chainOutput.GetStateAddress().Equal(chainKey.GetPublicKey().AsEd25519Address()))
	}

	_, block, txEssence, inputsCommitment := nextState(t, store, chainOutput.GetAliasOutput(), chainOutput.OutputID(), ts, nil)

	signatures, err := txEssence.Sign(
		inputsCommitment,
		chainKey.GetPrivateKey().AddressKeys(chainOutput.GetStateAddress()),
	)
	require.NoError(t, err)
	tx := &iotago.Transaction{
		Essence: txEssence,
		Unlocks: []iotago.Unlock{&iotago.SignatureUnlock{Signature: signatures[0]}},
	}

	txID, err := tx.ID()
	require.NoError(t, err)
	aliasOutputID := iotago.OutputIDFromTransactionIDAndIndex(txID, 0).UTXOInput()

	return block, tx, aliasOutputID
}
