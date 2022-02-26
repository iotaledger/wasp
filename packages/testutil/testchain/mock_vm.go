package testchain

import (
	// "strings"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"

	// "github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/stretchr/testify/require"
)

type MockedVMRunner struct {
	stateTransition *MockedStateTransition
	nextState       state.VirtualStateAccess
	tx              *iotago.TransactionEssence
	log             *logger.Logger
	t               *testing.T
}

func NewMockedVMRunner(t *testing.T, log *logger.Logger) *MockedVMRunner {
	ret := &MockedVMRunner{
		stateTransition: NewMockedStateTransition(t, nil),
		log:             log,
		t:               t,
	}
	ret.stateTransition.OnVMResult(func(vs state.VirtualStateAccess, tx *iotago.TransactionEssence) {
		ret.nextState = vs
		ret.tx = tx
	})
	return ret
}

func (r *MockedVMRunner) Run(task *vm.VMTask) {
	panic("TODO: implement")
	/*reqstr := strings.Join(iscp.ShortRequestIDs(iscp.TakeRequestIDs(task.Requests...)), ",")

	r.log.Debugf("VM input: state hash: %s, chain input: %s, requests: [%s]",
		task.VirtualStateAccess.StateCommitment(), iscp.OID(&task.AnchorOutputID), reqstr)

	calldata := make([]iscp.Calldata, len(task.Requests))
	for i := range calldata {
		calldata[i] = task.Requests[i]
	}
	r.stateTransition.NextState(task.VirtualStateAccess, task.AnchorOutput, task.TimeAssumption.Time, calldata...)
	task.ResultTransactionEssence = r.tx
	task.VirtualStateAccess = r.nextState
	newOut := transaction.GetAliasOutputFromEssence(task.ResultTransactionEssence, task.AnchorOutput.GetAliasAddress())
	require.NotNil(r.t, newOut)
	require.EqualValues(r.t, task.AnchorOutput.StateIndex+1, newOut.StateIndex)
	// essenceHash := hashing.HashData(task.ResultTransactionEssence.Bytes())
	// r.log.Debugf("mockedVMRunner: new state produced: stateIndex: #%d state hash: %s, essence hash: %s stateOutput: %s\n essence : %s",
	//	r.nextState.BlockIndex(), r.nextState.Hash().String(), essenceHash.String(), iscp.OID(newOut.ID()), task.ResultTransactionEssence.String())
	task.OnFinish(nil, nil, nil)*/
}

func NextState(
	t *testing.T,
	chainKey *cryptolib.KeyPair,
	vs state.VirtualStateAccess,
	chainOutput *iscp.AliasOutputWithID,
	ts time.Time,
	/*reqs ...iscp.Calldata,*/
) (nextvs state.VirtualStateAccess, tx *iotago.Transaction, aliasOutputID *iotago.UTXOInput) {
	if chainKey != nil {
		require.True(t, chainOutput.GetStateAddress().Equal(chainKey.GetPublicKey().AsEd25519Address()))
	}

	nextvs = vs.Copy()
	prevBlockIndex := vs.BlockIndex()
	counterKey := kv.Key(coreutil.StateVarBlockIndex + "counter")

	counterBin, err := nextvs.KVStore().Get(counterKey)
	require.NoError(t, err)

	counter, err := codec.DecodeUint64(counterBin, 0)
	require.NoError(t, err)

	suBlockIndex := state.NewStateUpdateWithBlockLogValues(prevBlockIndex+1, time.Time{}, vs.StateCommitment())

	suCounter := state.NewStateUpdate()
	counterBin = codec.EncodeUint64(counter + 1)
	suCounter.Mutations().Set(counterKey, counterBin)

	/*suReqs := state.NewStateUpdate()
	for i, req := range reqs {
		key := kv.Key(blocklog.NewRequestLookupKey(vs.BlockIndex()+1, uint16(i)).Bytes())
		suReqs.Mutations().Set(key, req.ID().Bytes())
	}*/

	nextvs.ApplyStateUpdate(suBlockIndex, suCounter /*, suReqs*/)
	require.EqualValues(t, prevBlockIndex+1, nextvs.BlockIndex())

	consumedOutput := chainOutput.GetAliasOutput()
	aliasID := consumedOutput.AliasID
	inputs := iotago.OutputIDs{chainOutput.OutputID()}
	txEssence := &iotago.TransactionEssence{
		NetworkID: 0,
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
		chainKey.GetPrivateKey().AddressKeys(chainOutput.GetStateAddress()),
	)
	require.NoError(t, err)
	tx = &iotago.Transaction{
		Essence:      txEssence,
		UnlockBlocks: []iotago.UnlockBlock{&iotago.SignatureUnlockBlock{Signature: signatures[0]}},
	}

	txID, err := tx.ID()
	require.NoError(t, err)
	aliasOutputID = iotago.OutputIDFromTransactionIDAndIndex(*txID, 0).UTXOInput()

	return nextvs, tx, aliasOutputID
}
