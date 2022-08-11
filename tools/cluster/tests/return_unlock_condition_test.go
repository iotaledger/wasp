package tests

import (
	"math"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/stretchr/testify/require"
)

// buils a normal tx to post a request to inccounter, optionally adds SDRC
func buildTX(t *testing.T, env *ChainEnv, addr iotago.Address, keyPair *cryptolib.KeyPair, addSDRC bool) *iotago.Transaction {
	outputs, err := env.Clu.L1Client().OutputMap(addr)
	require.NoError(t, err)

	outputIDs := make(iotago.OutputIDs, len(outputs))
	i := 0
	for id := range outputs {
		outputIDs[i] = id
		i++
	}

	tx, err := transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    keyPair,
		SenderAddress:    addr,
		UnspentOutputs:   outputs,
		UnspentOutputIDs: outputIDs,
		Request: &isc.RequestParameters{
			TargetAddress:  env.Chain.ChainAddress(),
			FungibleTokens: &isc.FungibleTokens{BaseTokens: 1 * isc.Million},
			Metadata: &isc.SendMetadata{
				TargetContract: nativeIncCounterSCHname,
				EntryPoint:     inccounter.FuncIncCounter.Hname(),
				GasBudget:      math.MaxUint64,
			},
		},
	})
	require.NoError(t, err)

	if !addSDRC {
		return tx
	}

	// tweak the tx , so the request output has a StorageDepositReturn unlock condition
	for i, out := range tx.Essence.Outputs {
		if out.FeatureSet().MetadataFeature() == nil {
			// skip if not the request output
			continue
		}
		customOut := out.Clone().(*iotago.BasicOutput)
		sendBackCondition := &iotago.StorageDepositReturnUnlockCondition{
			ReturnAddress: addr,
			Amount:        500,
		}
		customOut.Conditions = append(customOut.Conditions, sendBackCondition)
		tx.Essence.Outputs[i] = customOut
	}

	inputsCommitment := outputIDs.OrderedSet(outputs).MustCommitment()
	tx, err = transaction.CreateAndSignTx(outputIDs, inputsCommitment, tx.Essence.Outputs, keyPair, parameters.L1.Protocol.NetworkID())
	require.NoError(t, err)
	return tx
}

func TestSDRC(t *testing.T) {
	env := setupAdvancedInccounterTest(t, 1, []int{0})

	keyPair, addr, err := env.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	initialBlockIdx, err := env.Chain.BlockIndex()
	require.NoError(t, err)

	// // send a request with Storage Deposit Return Unlock
	txSDRC := buildTX(t, env, addr, keyPair, true)
	err = env.Clu.L1Client().PostTx(txSDRC)
	require.NoError(t, err)

	// wait some time and assert that the chain has not processed the request
	time.Sleep(10 * time.Second) // don't like the sleep here, but not sure there is a better way to do this

	// make sure the request is not picked up and the chain does not process it
	currentBlockIndex, err := env.Chain.BlockIndex()
	require.NoError(t, err)
	require.EqualValues(t, initialBlockIdx, currentBlockIndex)

	require.EqualValues(t, 0, env.getNativeContractCounter(nativeIncCounterSCHname))

	// send an equivalent request without StorageDepositReturnUnlockCondition
	txNormal := buildTX(t, env, addr, keyPair, false)
	err = env.Clu.L1Client().PostTx(txNormal)
	require.NoError(t, err)

	_, err = env.Clu.MultiClient().WaitUntilAllRequestsProcessedSuccessfully(env.Chain.ChainID, txNormal, 1*time.Minute)
	require.NoError(t, err)

	require.EqualValues(t, 1, env.getNativeContractCounter(nativeIncCounterSCHname))

	currentBlockIndex2, err := env.Chain.BlockIndex()
	require.NoError(t, err)
	require.EqualValues(t, initialBlockIdx+1, currentBlockIndex2)
}
