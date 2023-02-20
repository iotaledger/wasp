// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

////////////////////////////////////////////////////////////////////////////////
// TestChainLedger

type TestChainLedger struct {
	t           *testing.T
	utxoDB      *utxodb.UtxoDB
	governor    *cryptolib.KeyPair
	originator  *cryptolib.KeyPair
	chainID     isc.ChainID
	fetchedReqs map[iotago.Address]map[iotago.OutputID]bool
}

func NewTestChainLedger(t *testing.T, utxoDB *utxodb.UtxoDB, governor, originator *cryptolib.KeyPair) *TestChainLedger {
	return &TestChainLedger{
		t:           t,
		utxoDB:      utxoDB,
		governor:    governor,
		originator:  originator,
		fetchedReqs: map[iotago.Address]map[iotago.OutputID]bool{},
	}
}

// Only set after MakeTxChainOrigin.
func (tcl *TestChainLedger) ChainID() isc.ChainID {
	return tcl.chainID
}

func (tcl *TestChainLedger) MakeTxChainOrigin(committeeAddress iotago.Address) (*isc.AliasOutputWithID, isc.ChainID) {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(tcl.originator.Address())
	originTX, chainID, err := transaction.NewChainOriginTransaction(
		tcl.originator,
		committeeAddress,
		tcl.governor.Address(),
		1_000_000,
		outs,
		outIDs,
	)
	require.NoError(tcl.t, err)
	stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	require.NoError(tcl.t, err)
	require.NotNil(tcl.t, stateAnchor)
	require.NotNil(tcl.t, aliasOutput)
	originAO := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(originTX))
	tcl.chainID = chainID
	return originAO, chainID
}

func (tcl *TestChainLedger) MakeTxChainInit() []isc.Request {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(tcl.originator.Address())
	initTX, err := transaction.NewRootInitRequestTransaction(tcl.originator, tcl.chainID, "my test chain", outs, outIDs)
	require.NoError(tcl.t, err)
	require.NotNil(tcl.t, initTX)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(initTX))
	return tcl.findChainRequests(initTX)
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
				TargetAddress:                 tcl.chainID.AsAddress(),
				Assets:                        isc.NewAssetsBaseTokens(100_000_000),
				AdjustToMinimumStorageDeposit: false,
				Metadata: &isc.SendMetadata{
					TargetContract: accounts.Contract.Hname(),
					EntryPoint:     accounts.FuncDeposit.Hname(),
					GasBudget:      10_000,
				},
			},
		},
	)
	require.NoError(tcl.t, err)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(tx))
	return tcl.findChainRequests(tx)
}

func (tcl *TestChainLedger) MakeTxDeployIncCounterContract() []isc.Request {
	sender := tcl.originator
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(sender.Address())
	tx, err := transaction.NewRequestTransaction(
		transaction.NewRequestTransactionParams{
			SenderKeyPair:    sender,
			SenderAddress:    sender.Address(),
			UnspentOutputs:   outs,
			UnspentOutputIDs: outIDs,
			Request: &isc.RequestParameters{
				TargetAddress:                 tcl.chainID.AsAddress(),
				Assets:                        isc.NewAssetsBaseTokens(2_000_000),
				AdjustToMinimumStorageDeposit: false,
				Metadata: &isc.SendMetadata{
					TargetContract: root.Contract.Hname(),
					EntryPoint:     root.FuncDeployContract.Hname(),
					Params: codec.MakeDict(map[string]interface{}{
						root.ParamProgramHash: inccounter.Contract.ProgramHash,
						root.ParamDescription: "inccounter",
						root.ParamName:        inccounter.Contract.Name,
						inccounter.VarCounter: 0,
					}),
					GasBudget: 10_000,
				},
			},
		},
	)
	require.NoError(tcl.t, err)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(tx))
	return tcl.findChainRequests(tx)
}

func (tcl *TestChainLedger) FakeTX(baseAO *isc.AliasOutputWithID, nextCommitteeAddr iotago.Address) (*isc.AliasOutputWithID, *iotago.Transaction) {
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
		if !outAddr.Address.Equal(tcl.chainID.AsAddress()) {
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
