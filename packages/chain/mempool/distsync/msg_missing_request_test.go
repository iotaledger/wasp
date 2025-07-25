// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestMsgMissingRequestSerialization(t *testing.T) {
	senderKP := cryptolib.NewKeyPair()
	contract := governance.Contract.Hname()
	entryPoint := governance.FuncAddCandidateNode.Hname()
	gasBudget := gas.LimitsDefault.MaxGasPerRequest
	req := isc.NewOffLedgerRequest(isctest.RandomChainID(), isc.NewMessage(contract, entryPoint, nil), 0, gasBudget).Sign(senderKP)

	msg := &msgMissingRequest{
		gpa.BasicMessage{},
		isc.RequestRefFromRequest(req),
	}

	bcs.TestCodec(t, msg)

	msg = &msgMissingRequest{
		gpa.BasicMessage{},
		isc.RequestRefFromRequest(
			isc.NewOffLedgerRequest(
				isctest.TestChainID,
				isc.NewMessage(contract, entryPoint, nil),
				0,
				gasBudget,
			).Sign(
				cryptolib.TestKeyPair,
			),
		),
	}

	bcs.TestCodecAndHash(t, msg, "2e3a31ac716d")
}
