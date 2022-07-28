package testcore

import (
	"math"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/stretchr/testify/require"
)

func TestSendBack(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true}).
		WithNativeContract(inccounter.Processor)
	ch := env.NewChain(nil, "chain1")

	err := ch.DepositBaseTokensToL2(10*isc.Mi, nil)
	require.NoError(t, err)

	err = ch.DeployContract(nil, inccounter.Contract.Name, inccounter.Contract.ProgramHash, inccounter.VarCounter, 0)
	require.NoError(t, err)

	// send a normal request
	wallet, addr := env.NewKeyPairWithFunds()

	req := solo.NewCallParams(inccounter.Contract.Name, inccounter.FuncIncCounter.Name).WithMaxAffordableGasBudget()
	_, _, err = ch.PostRequestSyncTx(req, wallet)
	require.NoError(t, err)

	// check counter increments
	ret, err := ch.CallView(inccounter.Contract.Name, inccounter.ViewGetCounter.Name)
	require.NoError(t, err)
	counter, err := codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)

	// send a custom request
	allOuts, allOutIDs := ch.Env.GetUnspentOutputs(addr)
	tx, err := transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    wallet,
		SenderAddress:    addr,
		UnspentOutputs:   allOuts,
		UnspentOutputIDs: allOutIDs,
		Request: &isc.RequestParameters{
			TargetAddress:  ch.ChainID.AsAddress(),
			FungibleTokens: &isc.FungibleTokens{BaseTokens: 1 * isc.Mi},
			Metadata: &isc.SendMetadata{
				TargetContract: inccounter.Contract.Hname(),
				EntryPoint:     inccounter.FuncIncCounter.Hname(),
				GasBudget:      math.MaxUint64,
			},
		},
	})
	require.NoError(t, err)

	// tweak the tx before adding to the ledger, so the request output has a StorageDepositReturn unlock condition
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

	inputsCommitment := allOutIDs.OrderedSet(allOuts).MustCommitment()
	tx, err = transaction.CreateAndSignTx(allOutIDs, inputsCommitment, tx.Essence.Outputs, wallet, parameters.L1.Protocol.NetworkID())
	require.NoError(t, err)
	err = ch.Env.AddToLedger(tx)
	require.NoError(t, err)

	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	results := ch.RunRequestsSync(reqs, "post")
	// TODO for now the request must be skipped, in the future this needs to be refactored, so that the request is handled as expected
	require.Len(t, results, 0)

	// check counter is still the same (1)
	ret, err = ch.CallView(inccounter.Contract.Name, inccounter.ViewGetCounter.Name)
	require.NoError(t, err)
	counter, err = codec.DecodeInt64(ret.MustGet(inccounter.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}
