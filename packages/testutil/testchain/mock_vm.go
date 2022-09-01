package testchain

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/stretchr/testify/require"
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
	r.log.Debugf("Mocked VM runner: VM started for state %v commitment %v output %v",
		task.VirtualStateAccess.BlockIndex(), trie.RootCommitment(task.VirtualStateAccess.TrieNodeStore()), isc.OID(task.AnchorOutputID.UTXOInput()))
	nextvs, txEssence, inputsCommitment := nextState(r.t, task.VirtualStateAccess, task.AnchorOutput, task.AnchorOutputID, task.TimeAssumption, task.Requests...)
	task.VirtualStateAccess = nextvs
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
	r.log.Debugf("Mocked VM runner: VM completed; state %v commitment %v received", nextvs.BlockIndex(), trie.RootCommitment(nextvs.TrieNodeStore()))
	return nil
}

func nextState(
	t *testing.T,
	vs state.VirtualStateAccess,
	consumedOutput *iotago.AliasOutput,
	consumedOutputID iotago.OutputID,
	timeAssumption time.Time,
	reqs ...isc.Request,
) (state.VirtualStateAccess, *iotago.TransactionEssence, []byte) {
	nextvs := vs.Copy()
	prevBlockIndex := vs.BlockIndex()
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")

	counterBin, err := nextvs.KVStore().Get(counterKey)
	require.NoError(t, err)

	counter, err := codec.DecodeUint64(counterBin, 0)
	require.NoError(t, err)

	prev, err := state.L1CommitmentFromBytes(consumedOutput.StateMetadata)
	require.NoError(t, err)
	suBlockIndex := state.NewStateUpdateWithBlockLogValues(prevBlockIndex+1, timeAssumption.Add(time.Duration(len(reqs))*time.Nanosecond), &prev)

	suCounter := state.NewStateUpdate()
	counterBin = codec.EncodeUint64(counter + 1)
	suCounter.Mutations().Set(counterKey, counterBin)

	suReqs := state.NewStateUpdate()
	for i, req := range reqs {
		key := kv.Key(blocklog.NewRequestLookupKey(vs.BlockIndex()+1, uint16(i)).Bytes())
		suReqs.Mutations().Set(key, req.ID().Bytes())
	}

	nextvs.ApplyStateUpdate(suBlockIndex)
	nextvs.ApplyStateUpdate(suCounter)
	nextvs.ApplyStateUpdate(suReqs)
	nextvs.Commit()
	require.EqualValues(t, prevBlockIndex+1, nextvs.BlockIndex())

	block, err := nextvs.ExtractBlock()
	require.NoError(t, err)

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
				StateMetadata:  state.NewL1Commitment(trie.RootCommitment(nextvs.TrieNodeStore()), state.BlockHashFromData(block.EssenceBytes())).Bytes(),
				FoundryCounter: consumedOutput.FoundryCounter,
				Conditions:     consumedOutput.Conditions,
				Features:       consumedOutput.Features,
			},
		},
		Payload: nil,
	}

	inputsCommitment := iotago.Outputs{consumedOutput}.MustCommitment()

	return nextvs, txEssence, inputsCommitment
}

func NextState(
	t *testing.T,
	chainKey *cryptolib.KeyPair,
	vs state.VirtualStateAccess,
	chainOutput *isc.AliasOutputWithID,
	ts time.Time,
) (state.VirtualStateAccess, *iotago.Transaction, *iotago.UTXOInput) {
	if chainKey != nil {
		require.True(t, chainOutput.GetStateAddress().Equal(chainKey.GetPublicKey().AsEd25519Address()))
	}

	nextvs, txEssence, inputsCommitment := nextState(t, vs, chainOutput.GetAliasOutput(), chainOutput.OutputID(), ts)

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

	return nextvs, tx, aliasOutputID
}
