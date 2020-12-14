// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes/cbalances"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
)

type RequestSectionParams struct {
	TargetContractID coretypes.ContractID
	EntryPointCode   coretypes.Hname
	Timelock         uint32
	Transfer         map[balance.Color]int64 // should not not include request token. It is added automatically
	Vars             dict.Dict
}

type CreateRequestTransactionParams struct {
	NodeClient           nodeclient.NodeClient
	SenderSigScheme      signaturescheme.SignatureScheme
	RequestSectionParams []RequestSectionParams
	Mint                 map[address.Address]int64
	Post                 bool
	WaitForConfirmation  bool
}

func CreateRequestTransaction(par CreateRequestTransactionParams) (*sctransaction.Transaction, error) {
	senderAddr := par.SenderSigScheme.Address()
	allOuts, err := par.NodeClient.GetConfirmedAccountOutputs(&senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	for targetAddress, amount := range par.Mint {
		// TODO: check that targetAddress is not any target address in request blocks
		err = txb.MintColor(targetAddress, balance.ColorIOTA, amount)
		if err != nil {
			return nil, err
		}
	}

	for _, sectPar := range par.RequestSectionParams {
		reqSect := sctransaction.NewRequestSectionByWallet(sectPar.TargetContractID, sectPar.EntryPointCode).
			WithTimelock(sectPar.Timelock).
			WithTransfer(cbalances.NewFromMap(sectPar.Transfer))

		reqSect.WithArgs(sectPar.Vars)

		err = txb.AddRequestSection(reqSect)
		if err != nil {
			return nil, err
		}
	}
	tx, err := txb.Build(false)

	//dump := txb.Dump()

	if err != nil {
		return nil, err
	}
	tx.Sign(par.SenderSigScheme)

	// semantic check just in case
	if _, err := tx.Properties(); err != nil {
		return nil, err
	}
	//fmt.Printf("$$$$ dumping builder for %s\n%s\n", tx.ID().String(), dump)

	if !par.Post {
		return tx, nil
	}

	if !par.WaitForConfirmation {
		if err = par.NodeClient.PostTransaction(tx.Transaction); err != nil {
			return nil, err
		}
		return tx, nil
	}

	err = par.NodeClient.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
