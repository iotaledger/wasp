// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package apilib

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

// TODO: move to level1 client
func PostRequestTransaction(client level1.Level1Client, par sctransaction.NewRequestTransactionParams) (*ledgerstate.Transaction, error) {
	var err error

	if len(par.UnspentOutputs) == 0 {
		addr := ledgerstate.NewED25519Address(par.SenderKeyPair.PublicKey)
		par.UnspentOutputs, err = client.GetConfirmedOutputs(addr)
		if err != nil {
			return nil, fmt.Errorf("can't get outputs from the node: %v", err)
		}
	}

	tx, err := sctransaction.NewRequestTransaction(par)
	if err != nil {
		return nil, err
	}

	if err = client.PostTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}
