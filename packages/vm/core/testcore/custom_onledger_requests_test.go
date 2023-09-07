package testcore

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestNoSenderFeature(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	wallet, addr := env.NewKeyPairWithFunds()

	// ----------------------------------------------------------------

	// mint some NTs and withdraw them
	gasFee := 10 * gas.LimitsDefault.MinGasPerRequest
	withdrawAmount := 3 * gas.LimitsDefault.MinGasPerRequest
	err := ch.DepositAssetsToL2(isc.NewAssetsBaseTokens(withdrawAmount+gasFee), wallet)
	require.NoError(t, err)
	nativeTokenAmount := big.NewInt(123)
	sn, nativeTokenID, err := ch.NewFoundryParams(1234).
		WithUser(wallet).
		CreateFoundry()
	require.NoError(t, err)
	// mint some tokens for the user
	err = ch.MintTokens(sn, nativeTokenAmount, wallet)
	require.NoError(t, err)

	// withdraw native tokens to L1
	allowance := withdrawAmount
	baseTokensToSend := allowance + gasFee
	_, err = ch.PostRequestOffLedger(solo.NewCallParams(
		accounts.Contract.Name, accounts.FuncWithdraw.Name,
	).
		AddBaseTokens(baseTokensToSend).
		AddAllowanceBaseTokens(allowance).
		AddAllowanceNativeTokensVect(&iotago.NativeToken{
			ID:     nativeTokenID,
			Amount: nativeTokenAmount,
		}).
		WithGasBudget(gasFee),
		wallet)
	require.NoError(t, err)

	nft, _, err := ch.Env.MintNFTL1(wallet, addr, []byte("foobar"))
	require.NoError(t, err)

	// ----------------------------------------------------------------

	payoutAgentIDBalanceBefore := ch.L2Assets(ch.OriginatorAgentID)

	// send a custom request with Base tokens / NTs / NFT (but no sender feature)
	allOuts, allOutIDs := ch.Env.GetUnspentOutputs(addr)
	tx, err := transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    wallet,
		SenderAddress:    addr,
		UnspentOutputs:   allOuts,
		UnspentOutputIDs: allOutIDs,
		Request: &isc.RequestParameters{
			TargetAddress: ch.ChainID.AsAddress(),
			Assets: &isc.Assets{
				BaseTokens:   5 * isc.Million,
				NativeTokens: []*iotago.NativeToken{{ID: nativeTokenID, Amount: nativeTokenAmount}},
				NFTs:         []iotago.NFTID{nft.ID},
			},
			Metadata: &isc.SendMetadata{
				TargetContract: inccounter.Contract.Hname(),
				EntryPoint:     inccounter.FuncIncCounter.Hname(),
				GasBudget:      math.MaxUint64,
			},
		},
		NFT: nft,
	})
	require.NoError(t, err)

	// tweak the tx before adding to the ledger, so the request output has no sender feature
	for i, out := range tx.Essence.Outputs {
		if out.FeatureSet().MetadataFeature() == nil {
			// skip if not the request output
			continue
		}
		customOut := out.Clone().(*iotago.NFTOutput)                                   // must be NFT output because we're sending an NFT
		customOut.Features = iotago.Features{customOut.FeatureSet().MetadataFeature()} // keep metadata feature only
		tx.Essence.Outputs[i] = customOut
	}

	tx, err = transaction.CreateAndSignTx(tx.Essence.Inputs, tx.Essence.InputsCommitment[:], tx.Essence.Outputs, wallet, parameters.L1().Protocol.NetworkID())
	require.NoError(t, err)
	err = ch.Env.AddToLedger(tx)
	require.NoError(t, err)

	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	results := ch.RunRequestsSync(reqs, "post") // under normal circumstances this request won't reach the mempool
	require.Len(t, results, 1)
	require.NotNil(t, results[0].Receipt.Error)
	err = ch.ResolveVMError(results[0].Receipt.Error)
	testmisc.RequireErrorToBe(t, err, "sender unknown")

	// assert the assets were credited to the payout address
	payoutAgentIDBalance := ch.L2Assets(ch.OriginatorAgentID)
	require.Greater(t, payoutAgentIDBalance.BaseTokens, payoutAgentIDBalanceBefore.BaseTokens)
	require.EqualValues(t, payoutAgentIDBalance.NativeTokens[0].ID, nativeTokenID)
	require.EqualValues(t, payoutAgentIDBalance.NativeTokens[0].Amount.Uint64(), nativeTokenAmount.Uint64())
	require.EqualValues(t, payoutAgentIDBalance.NFTs[0], nft.ID)
}

func TestSendBack(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(inccounter.Processor)
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10*isc.Million, nil)
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
	counter, err := codec.DecodeInt64(ret.Get(inccounter.VarCounter))
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
			TargetAddress: ch.ChainID.AsAddress(),
			Assets:        &isc.Assets{BaseTokens: 1 * isc.Million},
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
			Amount:        1 * isc.Million,
		}
		customOut.Conditions = append(customOut.Conditions, sendBackCondition)
		tx.Essence.Outputs[i] = customOut
	}

	tx, err = transaction.CreateAndSignTx(tx.Essence.Inputs, tx.Essence.InputsCommitment[:], tx.Essence.Outputs, wallet, parameters.L1().Protocol.NetworkID())
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
	counter, err = codec.DecodeInt64(ret.Get(inccounter.VarCounter))
	require.NoError(t, err)
	require.EqualValues(t, 1, counter)
}

func TestBadMetadata(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	wallet, addr := env.NewKeyPairWithFunds()

	// send a custom request
	allOuts, allOutIDs := ch.Env.GetUnspentOutputs(addr)
	tx, err := transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    wallet,
		SenderAddress:    addr,
		UnspentOutputs:   allOuts,
		UnspentOutputIDs: allOutIDs,
		Request: &isc.RequestParameters{
			TargetAddress: ch.ChainID.AsAddress(),
			Assets:        &isc.Assets{BaseTokens: 1 * isc.Million},
			Metadata: &isc.SendMetadata{
				TargetContract: inccounter.Contract.Hname(),
				EntryPoint:     inccounter.FuncIncCounter.Hname(),
				GasBudget:      math.MaxUint64,
			},
		},
	})
	require.NoError(t, err)

	// tweak the tx before adding to the ledger, set bad metadata
	for i, out := range tx.Essence.Outputs {
		if out.FeatureSet().MetadataFeature() == nil {
			// skip if not the request output
			continue
		}
		customOut := out.Clone().(*iotago.BasicOutput)
		for ii, f := range customOut.Features {
			if mf, ok := f.(*iotago.MetadataFeature); ok {
				mf.Data = []byte("foobar")
				customOut.Features[ii] = mf
			}
		}
		tx.Essence.Outputs[i] = customOut
	}

	tx, err = transaction.CreateAndSignTx(tx.Essence.Inputs, tx.Essence.InputsCommitment[:], tx.Essence.Outputs, wallet, parameters.L1().Protocol.NetworkID())
	require.NoError(t, err)
	require.Zero(t, ch.L2BaseTokens(isc.NewAddressAgentID(addr)))
	err = ch.Env.AddToLedger(tx)
	require.NoError(t, err)

	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	results := ch.RunRequestsSync(reqs, "post")
	// assert request was processed with an error
	require.Len(t, results, 1)
	require.NotNil(t, results[0].Receipt.Error)
	err = ch.ResolveVMError(results[0].Receipt.Error)
	testmisc.RequireErrorToBe(t, err, "contract with hname 3030303030303030 not found")

	// assert funds were credited to the sender
	require.Positive(t, ch.L2BaseTokens(isc.NewAddressAgentID(addr)))
}
