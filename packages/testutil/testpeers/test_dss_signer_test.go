package testpeers_test

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
)

func TestDSSSigner(t *testing.T) {
	//
	// Infrastructure.
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	n := 4
	f := 1
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodeIDs := gpa.MakeTestNodeIDs(n)
	addr, dkRegs := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Create the signer.
	signer := testpeers.NewTestDSSSigner(addr, dkRegs, nodeIDs, peerIdentities, log)
	//
	// Use it.
	msg := []byte{1, 2, 3}
	signature := lo.Must(signer.Sign(msg))
	require.True(t, signature.Validate(msg))
}
