// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	_ "github.com/iotaledger/wasp/packages/sctransaction/properties"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
)

type RequestSectionParams struct {
	TargetContractID coretypes.ContractID
	EntryPointCode   coretypes.Hname
	TimeLock         uint32
	Transfer         coretypes.ColoredBalances // should not not include request token. It is added automatically
	Args             requestargs.RequestArgs
}

type CreateRequestTransactionParams struct {
	Level1Client         level1.Level1Client
	SenderSigScheme      signaturescheme.SignatureScheme
	RequestSectionParams []RequestSectionParams
	Mint                 map[address.Address]int64 // free tokens to be minted from IOTA color
	Post                 bool
	WaitForConfirmation  bool
}

func CreateRequestTransaction(par CreateRequestTransactionParams) (*sctransaction.Transaction, error) {
	senderAddr := par.SenderSigScheme.Address()
	allOuts, err := par.Level1Client.GetConfirmedAccountOutputs(&senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	for _, sectPar := range par.RequestSectionParams {
		reqSect := sctransaction.NewRequestSectionByWallet(sectPar.TargetContractID, sectPar.EntryPointCode).
			WithTimelock(sectPar.TimeLock).
			WithTransfer(sectPar.Transfer)

		reqSect.WithArgs(sectPar.Args)

		err = txb.AddRequestSection(reqSect)
		if err != nil {
			return nil, err
		}
	}
	txb.AddMinting(par.Mint)
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
		if err = par.Level1Client.PostTransaction(tx.Transaction); err != nil {
			return nil, err
		}
		return tx, nil
	}

	err = par.Level1Client.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
