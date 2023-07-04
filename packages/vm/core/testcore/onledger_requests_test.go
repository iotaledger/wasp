package testcore

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/iotaledger/wasp/packages/transaction"
)

func TestNoSenderFeature(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	ch := env.NewChain()

	wallet, addr := env.NewKeyPairWithFunds()

	// ----------------------------------------------------------------

	// mint some NTs and withdraw them
	err := ch.DepositAssetsToL2(isc.NewAssetsBaseTokens(10*isc.Million), wallet)
	require.NoError(t, err)
	nativeTokenAmount := big.NewInt(123)
	sn, nativeTokenID, err := ch.NewFoundryParams(1234).
		WithUser(wallet).
		CreateFoundry()
	require.NoError(t, err)
	// mint some tokens for the user
	err = ch.MintTokens(sn, nativeTokenAmount, wallet)
	require.NoError(t, err)

	// withdraw NTs to L1
	req := solo.NewCallParams("accounts", "withdraw").
		AddAllowanceBaseTokens(5 * isc.Million).
		AddAllowanceNativeTokensVect(&iotago.NativeToken{
			ID:     nativeTokenID,
			Amount: nativeTokenAmount,
		})
	_, err = ch.PostRequestOffLedger(req, wallet)
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

	tx, err = transaction.CreateAndSignTx(allOutIDs, tx.Essence.InputsCommitment[:], tx.Essence.Outputs, wallet, parameters.L1().Protocol.NetworkID())
	require.NoError(t, err)
	err = ch.Env.AddToLedger(tx)
	require.NoError(t, err)

	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	results := ch.RunRequestsSync(reqs, "post")
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
