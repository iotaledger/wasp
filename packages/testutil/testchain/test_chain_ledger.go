// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testchain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

////////////////////////////////////////////////////////////////////////////////
// TestChainLedger

type TestChainLedger struct {
	t *testing.T
	//utxoDB      *utxodb.UtxoDB
	governor    *cryptolib.KeyPair
	chainID     isc.ChainID
	fetchedReqs map[cryptolib.AddressKey]map[iotago.OutputID]bool
}

func NewTestChainLedger(t *testing.T /*utxoDB *utxodb.UtxoDB*/, originator *cryptolib.KeyPair) *TestChainLedger {
	return &TestChainLedger{
		t: t,
		//utxoDB:      utxoDB,
		governor:    originator,
		fetchedReqs: map[cryptolib.AddressKey]map[iotago.OutputID]bool{},
	}
}

// Only set after MakeTxChainOrigin.
func (tcl *TestChainLedger) ChainID() isc.ChainID {
	return tcl.chainID
}

func (tcl *TestChainLedger) MakeTxChainOrigin(committeeAddress *cryptolib.Address) (*iotago.Transaction, *isc.StateAnchor, isc.ChainID) {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(tcl.governor.Address())
	panic("refactor me: origin.NewChainOriginTransaction")
	var originTX *iotago.Transaction
	var chainID isc.ChainID

	err := errors.New("refactor me: deployChain")
	_ = outs
	_ = outIDs

	require.NoError(tcl.t, err)
	stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	require.NoError(tcl.t, err)
	require.NotNil(tcl.t, stateAnchor)
	require.NotNil(tcl.t, aliasOutput)
	originAO := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(originTX))
	tcl.chainID = chainID
	return originTX, originAO, chainID
}

func (tcl *TestChainLedger) MakeTxAccountsDeposit(account *cryptolib.KeyPair) []isc.Request {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(account.Address())
	tx, err := transaction.NewRequestTransaction(
		transaction.NewRequestTransactionParams{
			SenderKeyPair:    account,
			SenderAddress:    account.Address(),
			UnspentOutputs:   outs,
			UnspentOutputIDs: outIDs,
			Request: &isc.RequestParameters{
				TargetAddress: tcl.chainID.AsAddress(),
				Assets:        isc.NewAssets(100_000_000),
				Metadata: &isc.SendMetadata{
					Message:   accounts.FuncDeposit.Message(),
					GasBudget: 2 * gas.LimitsDefault.MinGasPerRequest,
				},
			},
		},
	)
	require.NoError(tcl.t, err)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(tx))
	return tcl.findChainRequests(tx)
}

func (tcl *TestChainLedger) FakeStateTransition(baseAO *iscmove.AnchorWithRef, stateCommitment *state.L1Commitment) *iscmove.AnchorWithRef {
	stateMetadata := transaction.NewStateMetadata(
		stateCommitment,
		gas.DefaultFeePolicy(),
		0,
		"",
	)
	anchorOutput := &iotago.AliasOutput{
		Amount:        baseAO.GetAliasOutput().Deposit(),
		AliasID:       tcl.chainID.AsAliasID(),
		StateIndex:    baseAO.GetStateIndex() + 1,
		StateMetadata: stateMetadata.Bytes(),
		Conditions: iotago.UnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: tcl.governor.Address().AsIotagoAddress()},
			&iotago.GovernorAddressUnlockCondition{Address: tcl.governor.Address().AsIotagoAddress()},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: tcl.chainID.AsAddress().AsIotagoAddress(),
			},
		},
	}
	return isc.NewAliasOutputWithID(anchorOutput, iotago.OutputID{byte(anchorOutput.StateIndex)})
}

func (tcl *TestChainLedger) FakeRotationTX(baseAO *iscmove.AnchorWithRef, nextCommitteeAddr *cryptolib.Address) (*iscmove.AnchorWithRef, *iotago.Transaction) {
	tx, err := transaction.NewRotateChainStateControllerTx(
		tcl.chainID.AsAliasID(),
		nextCommitteeAddr,
		baseAO.OutputID(),
		baseAO.GetAliasOutput(),
		tcl.governor,
	)
	if err != nil {
		panic(err)
	}
	outputs, err := tx.OutputsSet()
	if err != nil {
		panic(err)
	}
	for outputID, output := range outputs {
		if output.Type() == iotago.OutputAlias {
			ao := output.(*iotago.AliasOutput)
			ao.StateIndex = baseAO.GetStateIndex() + 1 // Fake next state index, just for tests.
			return isc.NewAliasOutputWithID(ao, outputID), tx
		}
	}
	panic("alias output not found")
}

func (tcl *TestChainLedger) findChainRequests(tx *iotago.Transaction) []isc.Request {
	reqs := []isc.Request{}
	outputs, err := tx.OutputsSet()
	require.NoError(tcl.t, err)
	for outputID, output := range outputs {
		// If that's alias output of the chain, then it is not a request.
		if output.Type() == iotago.OutputAlias {
			outAsAlias := output.(*iotago.AliasOutput)
			if outAsAlias.AliasID == tcl.chainID.AsAliasID() {
				continue // That's our alias output, not the request, skip it here.
			}
			if outAsAlias.AliasID.Empty() {
				implicitAliasID := iotago.AliasIDFromOutputID(outputID)
				if implicitAliasID == tcl.chainID.AsAliasID() {
					continue // That's our origin alias output, not the request, skip it here.
				}
			}
		}
		//
		// Otherwise check the receiving address.
		outAddr := output.UnlockConditionSet().Address()
		if outAddr == nil {
			continue
		}
		if !cryptolib.NewAddressFromIotago(outAddr.Address).Equals(tcl.chainID.AsAddress()) {
			continue
		}
		req, err := isc.OnLedgerFromUTXO(output, outputID)
		if err != nil {
			continue
		}
		reqs = append(reqs, req)
	}
	return reqs
}
