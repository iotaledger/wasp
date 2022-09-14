// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
)

// NOTE: this test tests verification of off ledger requests rather than chainimpl.
// It should possibly be moved to other place.
func TestValidateOffledger(t *testing.T) {
	c := &chainObj{
		chainID: isc.RandomChainID(),
	}
	req := testutil.DummyOffledgerRequest(c.chainID)
	require.NoError(t, c.validateRequest(req))
	req.(isc.UnsignedOffLedgerRequest).WithNonce(999) // signature must be invalid after request content changes
	require.Error(t, c.validateRequest(req))

	wrongChainReq := testutil.DummyOffledgerRequest(isc.RandomChainID())
	require.Error(t, c.validateRequest(wrongChainReq))
}
