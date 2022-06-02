// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/stretchr/testify/require"
)

// NOTE: this test tests verification of off ledger requests rather than chainimpl.
// It should possibly be moved to other place.
func TestValidateOffledger(t *testing.T) {
	c := &chainObj{
		chainID: iscp.RandomChainID(),
	}
	req := testutil.DummyOffledgerRequest(c.chainID)
	require.NoError(t, c.validateRequest(req))
	req.(iscp.UnsignedOffLedgerRequest).WithNonce(999) // signature must be invalid after request content changes
	require.Error(t, c.validateRequest(req))

	wrongChainReq := testutil.DummyOffledgerRequest(iscp.RandomChainID())
	require.Error(t, c.validateRequest(wrongChainReq))
}
